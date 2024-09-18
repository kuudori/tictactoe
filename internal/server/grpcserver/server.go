package grpcserver

import (
	"TicTacToe/internal/grpc/interceptors"
	"TicTacToe/internal/server/gameserver"
	"fmt"
	"google.golang.org/grpc"
	"log/slog"
	"net"
)

// GRPCServer manages the gRPC server instance.
type GRPCServer struct {
	Server *grpc.Server
	port   int
}

// NewGRPCServer initializes a new GRPCServer instance.
func NewGRPCServer(port int, gameSrv *gameserver.GameServer) *GRPCServer {
	interceptor := interceptors.AuthInterceptor(gameSrv)

	grpcSrv := grpc.NewServer(grpc.UnaryInterceptor(interceptor))

	return &GRPCServer{
		Server: grpcSrv,
		port:   port,
	}
}

// Start starts the gRPC server on the specified port.
func (g *GRPCServer) Start() error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", g.port))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	slog.Info("Starting gRPC server", "port", g.port)

	if err := g.Server.Serve(listener); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}

// Stop gracefully stops the gRPC server.
func (g *GRPCServer) Stop() {
	g.Server.GracefulStop()
	slog.Info("gRPC server stopped gracefully")
}
