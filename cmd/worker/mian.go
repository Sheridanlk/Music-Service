package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/Sheridanlk/Music-Service/internal/app/worker"
	"github.com/Sheridanlk/Music-Service/internal/config"
	"github.com/Sheridanlk/Music-Service/internal/logger"
)

func main() {
	cfg := config.Load()
	log := logger.SetupLogger(cfg.Env)

	log.Info("starting worker", "env", cfg.Env)

	app := worker.New(log, cfg.PostgreSQL, cfg.MinIOClient, cfg.MinioStorage, cfg.RabbitMQ)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	app.Start(ctx)

	<-ctx.Done()

	app.Stop()

	log.Info("worker stopped")
}
