package inmem

import (
	"TicTacToe/internal/game"
	"TicTacToe/internal/storage"
	"context"
	"errors"
	"sync"
)

type GameStorage struct {
	games   map[string]*game.Game
	players map[string]*game.Player
	mu      sync.RWMutex
}

func NewGameStorage() storage.GameStorage {
	return &GameStorage{
		games:   make(map[string]*game.Game),
		players: make(map[string]*game.Player),
	}
}

func (s *GameStorage) CreatePlayer(ctx context.Context, player *game.Player) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.games[player.ID]; exists {
		return errors.New("player already exists")
	}

	s.players[player.ID] = player

	return nil

}

func (s *GameStorage) GetPlayer(ctx context.Context, ID string) (*game.Player, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	player, exists := s.players[ID]

	return player, exists

}

func (s *GameStorage) CreateGame(ctx context.Context, game *game.Game) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.games[game.ID]; exists {
		return errors.New("game already exists")
	}

	s.games[game.ID] = game

	return nil
}

func (s *GameStorage) GetGame(ctx context.Context, gameID string) (*game.Game, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	game, exists := s.games[gameID]

	return game, exists
}

func (s *GameStorage) UpdateGame(ctx context.Context, game *game.Game) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.games[game.ID]; !exists {
		return errors.New("game not found")
	}

	s.games[game.ID] = game
	return nil
}

func (s *GameStorage) DeleteGame(ctx context.Context, gameID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.games, gameID)
	return nil
}
