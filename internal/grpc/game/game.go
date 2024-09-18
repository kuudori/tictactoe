package game

import (
	tictactoev1 "TicTacToe/api/tictactoe"
	"TicTacToe/internal/game"
	"TicTacToe/internal/server/gameserver"
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type serverAPI struct {
	tictactoev1.UnimplementedGameServiceServer
	gameServer *gameserver.GameServer
}

func Register(gRPC *grpc.Server, gameSrv *gameserver.GameServer) {
	tictactoev1.RegisterGameServiceServer(gRPC, &serverAPI{
		gameServer: gameSrv,
	})
}

func (s *serverAPI) Login(ctx context.Context, req *tictactoev1.LoginRequest) (*tictactoev1.PlayerData, error) {
	player, err := s.gameServer.LoginPlayer(ctx, req.GetPlayerName())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &tictactoev1.PlayerData{
		PlayerId:   player.ID,
		PlayerName: player.Name,
	}, nil
}

func (s *serverAPI) CreateGame(ctx context.Context, req *tictactoev1.CreateGameRequest) (*tictactoev1.GameData, error) {
	player, ok := ctx.Value("player").(*game.Player)
	if !ok {
		return nil, status.Error(codes.Internal, "auth error")
	}
	gameData, err := s.gameServer.CreateGame(ctx, player, req.GetPassword())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	gameProto := game.GameToProto(gameData)

	return gameProto, nil
}

func (s *serverAPI) JoinGame(ctx context.Context, req *tictactoev1.JoinGameRequest) (*tictactoev1.GameData, error) {
	player, ok := ctx.Value("player").(*game.Player)
	if !ok {
		return nil, status.Error(codes.Internal, "auth error")
	}
	gameData, err := s.gameServer.JoinGame(ctx, req.GetGameId(), player, req.GetPassword())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	protoGame := game.GameToProto(gameData)
	gameData.Updates <- protoGame

	return protoGame, nil
}

func (s *serverAPI) LeaveGame(ctx context.Context, req *tictactoev1.LeaveGameRequest) (*tictactoev1.GameData, error) {
	player, ok := ctx.Value("player").(*game.Player)
	if !ok {
		return nil, status.Error(codes.Internal, "auth error")
	}
	gameData, err := s.gameServer.LeaveGame(ctx, req.GameId, player.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	protoGame := game.GameToProto(gameData)

	return protoGame, nil
}

func (s *serverAPI) MakeMove(ctx context.Context, req *tictactoev1.MoveRequest) (*tictactoev1.GameData, error) {
	player, ok := ctx.Value("player").(*game.Player)
	if !ok {
		return nil, status.Error(codes.Internal, "auth error")
	}
	gameData, err := s.gameServer.MakeMove(ctx, req.GetGameId(), player, req.GetPosition())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	protoGame := game.GameToProto(gameData)
	gameData.Updates <- protoGame

	return protoGame, nil
}

func (s *serverAPI) GetGameState(req *tictactoev1.GameRequest, stream tictactoev1.GameService_GetGameStateServer) error {
	md, ok := metadata.FromIncomingContext(stream.Context())
	if !ok {
		return status.Error(codes.InvalidArgument, "missing metadata")
	}
	playerIds := md.Get("player-id")
	if len(playerIds) == 0 {
		return status.Error(codes.InvalidArgument, "missing player name")
	}
	updateChan, err := s.gameServer.GetGameData(stream.Context(), req.GameId, playerIds[0])
	if err != nil {
		return status.Error(codes.NotFound, err.Error())
	}

	for {
		select {
		case update, ok := <-updateChan:
			if !ok {
				return nil
			}
			if err := stream.Send(update); err != nil {
				return status.Error(codes.Internal, "failed to send game state update")
			}
		case <-stream.Context().Done():
			return nil
		}
	}
}
