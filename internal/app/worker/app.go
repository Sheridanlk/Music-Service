package worker

import (
	"context"
	"log/slog"
	"os"

	"github.com/Sheridanlk/Music-Service/internal/app/worker/consumer"
	"github.com/Sheridanlk/Music-Service/internal/broker"
	"github.com/Sheridanlk/Music-Service/internal/config"
	"github.com/Sheridanlk/Music-Service/internal/services/tracks/hls"
	"github.com/Sheridanlk/Music-Service/internal/storage/media"
	"github.com/Sheridanlk/Music-Service/internal/storage/postgresql"
)

type App struct {
	storage  *postgresql.Storage
	broker   *broker.RabbitMQ
	consumer *consumer.HlsConsumer

	hlsConsumerDone <-chan struct{}
}

func New(log *slog.Logger,
	storageCfg config.PostgreSQL,
	minioClientCfg config.MinIOClient,
	minioStorageCfg config.MinioStorage,
	rabbitCfg config.RabbitMQ,
) *App {
	storage, err := postgresql.New(storageCfg.Host, storageCfg.UserName, storageCfg.Password, storageCfg.DBName, storageCfg.Port)
	if err != nil {
		log.Error("failed to init storage", slog.String("error", err.Error()))
		os.Exit(1)
	}

	minioStorage, err := media.New(minioClientCfg.Endpoint, minioClientCfg.AccessKeyID, minioClientCfg.SecretAccessKey, minioClientCfg.UseSSL)
	if err != nil {
		log.Error("failed to init minio storage", slog.String("error", err.Error()))
		os.Exit(1)
	}

	taskBroker, err := broker.New(rabbitCfg.UserName, rabbitCfg.Password, rabbitCfg.Host, rabbitCfg.Port)
	if err != nil {
		log.Error("failed to init rabbitmq producer", slog.String("error", err.Error()))
		os.Exit(1)
		// TODO: retries
	}

	hlsService := hls.New(log, storage, minioStorage, minioStorageCfg.HLSBucket)

	msgs, _ := taskBroker.GetTrackTaskStream()

	hlsConsumer := consumer.New(log, hlsService, msgs)

	return &App{
		storage:  storage,
		broker:   taskBroker,
		consumer: hlsConsumer,
	}
}

func (a *App) Start(ctx context.Context) {
	a.hlsConsumerDone = a.consumer.Consume(ctx)
}

func (a *App) Stop() {
	// TODO: add logs
	<-a.hlsConsumerDone
	a.broker.Close()
	a.storage.Close()

}
