package gameserver

import (
	tictactoev1 "TicTacToe/api/tictactoe"
	"TicTacToe/internal/game"
	"TicTacToe/internal/storage"
	"TicTacToe/internal/utils"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
)

type GameServer struct {
	storage storage.GameStorage
	mu      sync.RWMutex
}

func NewGameServer(storage storage.GameStorage) *GameServer {

	return &GameServer{
		storage: storage,
	}
}

func (gs *GameServer) LoginPlayer(ctx context.Context, playerName string) (*game.Player, error) {
	newPlayer := &game.Player{
		ID:   utils.GenerateUniqueID(),
		Name: playerName,
	}
	if err := gs.storage.CreatePlayer(ctx, newPlayer); err != nil {
		return nil, fmt.Errorf("failed to create player: %w", err)
	}

	return newPlayer, nil
}

func (gs *GameServer) GetPlayer(playerID string) (*game.Player, bool) {
	return gs.storage.GetPlayer(context.Background(), playerID)
}

func (gs *GameServer) CreateGame(ctx context.Context, creator *game.Player, password string) (*game.Game, error) {
	newGame := &game.Game{
		ID:            utils.GenerateUniqueID(),
		PlayerX:       creator,
		Board:         make([]string, 9),
		CurrentPlayer: creator,
		Status:        tictactoev1.GameStatus_WAITING_FOR_PLAYER,
		Event:         tictactoev1.GameEvent_GAME_CREATED,
		Password:      password,
		Updates:       make(chan *tictactoev1.GameData, 10),
		Players:       make(map[string]chan *tictactoev1.GameData),
	}

	if err := gs.storage.CreateGame(ctx, newGame); err != nil {
		return nil, fmt.Errorf("failed to create game: %w", err)
	}

	newGame.Players[creator.ID] = make(chan *tictactoev1.GameData, 10)
	go gs.broadcastUpdates(ctx, newGame.ID)
	return newGame, nil
}

func (gs *GameServer) JoinGame(ctx context.Context, gameID string, player *game.Player, password string) (*game.Game, error) {

	gameData, exists := gs.storage.GetGame(ctx, gameID)
	if !exists {
		return nil, errors.New("game not found")
	}

	if gameData.Status != tictactoev1.GameStatus_WAITING_FOR_PLAYER {
		return nil, errors.New("game has already started / finished")
	}

	if gameData.Password != password {
		return nil, errors.New("incorrect password")
	}

	if gameData.PlayerO != nil {
		return nil, errors.New("game is full")
	}

	gameData.Players[player.ID] = make(chan *tictactoev1.GameData, 10)

	gameData.PlayerO = player
	gameData.Status = tictactoev1.GameStatus_IN_PROGRESS
	gameData.Event = tictactoev1.GameEvent_PLAYER_JOINED
	gameData.CurrentPlayer = gameData.PlayerX // First player (X) starts

	if err := gs.storage.UpdateGame(ctx, gameData); err != nil {
		return nil, fmt.Errorf("failed to update game: %w", err)
	}

	return gameData, nil
}

func (gs *GameServer) MakeMove(ctx context.Context, gameID string, player *game.Player, position int32) (*game.Game, error) {

	gameData, exists := gs.storage.GetGame(ctx, gameID)
	if !exists {
		return nil, errors.New("game not found")
	}
	if gameData.Status == tictactoev1.GameStatus_WAITING_FOR_PLAYER {
		return nil, errors.New("game not started")
	}
	if position < 0 || position >= 9 {
		return nil, errors.New("invalid position")
	}
	if gameData.CurrentPlayer.ID != player.ID {
		return nil, errors.New("it's not your turn")
	}
	if filled := gameData.Board[position]; filled != "" {
		return nil, errors.New("can't move here")
	}

	symbol := "X"
	if player.ID == gameData.PlayerO.ID {
		symbol = "O"
	}

	gameData.Board[position] = symbol

	winner := utils.CheckWin(gameData.Board)
	if winner != "" {
		gameData.Winner = player.Name
		gameData.Status = tictactoev1.GameStatus_FINISHED
		gameData.Event = tictactoev1.GameEvent_GAME_OVER
	} else if utils.IsBoardFull(gameData.Board) {
		gameData.Status = tictactoev1.GameStatus_FINISHED
		gameData.Event = tictactoev1.GameEvent_GAME_OVER
	} else {
		if player.ID == gameData.PlayerX.ID {
			gameData.CurrentPlayer = gameData.PlayerO
		} else {
			gameData.CurrentPlayer = gameData.PlayerX
		}
		gameData.Event = tictactoev1.GameEvent_MOVE_MADE
	}

	if err := gs.storage.UpdateGame(ctx, gameData); err != nil {
		return nil, fmt.Errorf("failed to update game: %w", err)
	}

	return gameData, nil
}

func (gs *GameServer) LeaveGame(ctx context.Context, gameID string, playerID string) (*game.Game, error) {
	gameData, exists := gs.storage.GetGame(ctx, gameID)
	if !exists {
		return nil, errors.New("game not found")
	}

	if gameData.PlayerX != nil && playerID == gameData.PlayerX.ID {
		gameData.PlayerX = nil
	} else if gameData.PlayerO != nil && playerID == gameData.PlayerO.ID {
		gameData.PlayerO = nil
	} else {
		return nil, errors.New("player is not in this game")
	}

	close(gameData.Players[playerID])
	delete(gameData.Players, playerID)

	gameData.Status = tictactoev1.GameStatus_FINISHED
	gameData.Event = tictactoev1.GameEvent_PLAYER_LEAVED
	gameData.CurrentPlayer = nil

	if err := gs.storage.UpdateGame(ctx, gameData); err != nil {
		return nil, fmt.Errorf("failed to update game: %w", err)
	}

	if len(gameData.Players) == 0 {
		err := gs.storage.DeleteGame(ctx, gameID)
		if err != nil {
			slog.Error("Game is not deleted: ", gameID)
		}
	}

	return gameData, nil
}

func (gs *GameServer) GetGame(gameID string) (*game.Game, bool) {
	return gs.storage.GetGame(context.Background(), gameID)
}

func (gs *GameServer) GetGameData(ctx context.Context, gameID, playerId string) (chan *tictactoev1.GameData, error) {
	gameData, exists := gs.storage.GetGame(ctx, gameID)
	if !exists {
		return nil, errors.New("game not found")
	}

	playerChan, exists := gameData.Players[playerId]
	if !exists {
		playerChan = make(chan *tictactoev1.GameData, 10)
		gameData.Players[playerId] = playerChan
	}

	return playerChan, nil
}

func (gs *GameServer) broadcastUpdates(ctx context.Context, gameID string) {
	for {
		gameData, exists := gs.storage.GetGame(ctx, gameID)
		if !exists {
			return
		}

		update, ok := <-gameData.Updates
		if !ok {
			return
		}

		for playerID, clientChan := range gameData.Players {
			select {
			case clientChan <- update:
			default:
				close(clientChan)
				_, err := gs.LeaveGame(ctx, gameID, playerID)
				if err != nil {
					slog.Error("Leave game error: ", err)
				}
			}
		}

		if gameData.Status == tictactoev1.GameStatus_FINISHED {
			close(gameData.Updates)
			return
		}
	}
}
