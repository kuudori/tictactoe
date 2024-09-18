package storage

import (
	"TicTacToe/internal/game"
	"context"
)

type GameStorage interface {
	CreatePlayer(ctx context.Context, player *game.Player) error
	GetPlayer(ctx context.Context, playerID string) (*game.Player, bool)
	CreateGame(ctx context.Context, game *game.Game) error
	GetGame(ctx context.Context, gameID string) (*game.Game, bool)
	UpdateGame(ctx context.Context, game *game.Game) error
	DeleteGame(ctx context.Context, gameID string) error
}
