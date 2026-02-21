package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/Sheridanlk/Music-Service/internal/app"
	"github.com/Sheridanlk/Music-Service/internal/config"
	"github.com/Sheridanlk/Music-Service/internal/logger"
)

func main() {
	cfg := config.Load()
	log := logger.SetupLogger(cfg.Env)

	log.Info("starting music service", slog.String("env", cfg.Env))

	application := app.New(log, cfg.PostgreSQL, cfg.HTTPServer, cfg.MinIOClient, cfg.MinioStorage)

	go application.Server.Start()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	<-stop

	application.Server.Stop()
	application.Storage.Close()

	log.Info("application stopped")

}
