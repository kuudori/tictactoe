package app

import (
	"TicTacToe/internal/grpc/game"
	"TicTacToe/internal/server/gameserver"
	"TicTacToe/internal/server/grpcserver"
	storage "TicTacToe/internal/storage/inmem"
)

// App represents the main application containing the game server and gRPC server.
type App struct {
	GameServer *gameserver.GameServer
	GrpcServer *grpcserver.GRPCServer
	port       int
}

// New initializes the App with the specified port.
func New(port int) *App {
	gameStorage := storage.NewGameStorage()
	gameSrv := gameserver.NewGameServer(gameStorage)
	grpcSrv := grpcserver.NewGRPCServer(port, gameSrv)

	game.Register(grpcSrv.Server, gameSrv)

	return &App{
		GameServer: gameSrv,
		GrpcServer: grpcSrv,
		port:       port,
	}
}
