package game

import (
	tictactoev1 "TicTacToe/api/tictactoe"
)

type Player struct {
	ID   string
	Name string
}

type Game struct {
	ID            string
	PlayerX       *Player
	PlayerO       *Player
	Board         []string
	CurrentPlayer *Player
	Status        tictactoev1.GameStatus
	Event         tictactoev1.GameEvent
	Password      string
	Updates       chan *tictactoev1.GameData
	Players       map[string]chan *tictactoev1.GameData
	Winner        string
}

func PlayerToProto(p *Player) *tictactoev1.PlayerData {
	if p == nil {
		return nil
	}
	return &tictactoev1.PlayerData{
		PlayerId:   p.ID,
		PlayerName: p.Name,
	}
}

func PlayerFromProto(p *tictactoev1.PlayerData) *Player {
	if p == nil {
		return nil
	}
	return &Player{
		ID:   p.PlayerId,
		Name: p.PlayerName,
	}
}

func GameToProto(g *Game) *tictactoev1.GameData {
	return &tictactoev1.GameData{
		Id:            g.ID,
		PlayerX:       PlayerToProto(g.PlayerX),
		PlayerO:       PlayerToProto(g.PlayerO),
		Board:         g.Board,
		CurrentPlayer: PlayerToProto(g.CurrentPlayer),
		Status:        g.Status,
		Event:         g.Event,
		Password:      g.Password,
		Winner:        g.Winner,
	}
}
