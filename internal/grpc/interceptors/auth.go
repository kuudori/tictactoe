package interceptors

import (
	"TicTacToe/internal/server/gameserver"
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func AuthInterceptor(gameServer *gameserver.GameServer) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if info.FullMethod == "/game.GameService/Login" {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "no metadata provided")
		}

		playerIDs := md.Get("player-id")
		if len(playerIDs) == 0 {
			return nil, status.Error(codes.Unauthenticated, "player-id not provided")
		}

		playerID := playerIDs[0]

		player, exists := gameServer.GetPlayer(playerID)
		if !exists {
			return nil, status.Error(codes.Unauthenticated, "invalid player-id")
		}

		newCtx := context.WithValue(ctx, "player", player)

		return handler(newCtx, req)
	}
}
