package main

import (
	"TicTacToe/internal/app"
	"TicTacToe/internal/config"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func setupLogger() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)
}

func loadConfig() *config.Config {
	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("Failed to load configuration: ", err)
	}
	return cfg
}

func main() {
	setupLogger()
	cfg := loadConfig()

	application := app.New(cfg.GRPC.Port)

	// start grpc server with goroutine
	go func() {
		err := application.GrpcServer.Start()
		if err != nil {
			panic("could not start server")
		}
	}()

	// gracefully stop server on syscall
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	<-stop
	application.GrpcServer.Stop()
}
