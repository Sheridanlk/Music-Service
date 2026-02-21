package app

import (
	"log/slog"
	"os"

	"github.com/Sheridanlk/Music-Service/internal/app/server"
	"github.com/Sheridanlk/Music-Service/internal/clients/minio"
	"github.com/Sheridanlk/Music-Service/internal/config"
	"github.com/Sheridanlk/Music-Service/internal/http/router/chi"
	"github.com/Sheridanlk/Music-Service/internal/services/tracks/stream"
	"github.com/Sheridanlk/Music-Service/internal/services/tracks/upload"
	"github.com/Sheridanlk/Music-Service/internal/storage/media"
	"github.com/Sheridanlk/Music-Service/internal/storage/postgresql"
)

type App struct {
	Storage *postgresql.Storage
	Server  *server.App
}

func New(log *slog.Logger, storageCfg config.PostgreSQL, serverCfg config.HTTPServer, minioClientCfg config.MinIOClient, minioStorageCfg config.MinioStorage) *App {
	storage, err := postgresql.New(storageCfg.Host, storageCfg.UserName, storageCfg.Password, storageCfg.DBName, storageCfg.Port)
	if err != nil {
		log.Error("failed to init storage", slog.String("error", err.Error()))
		os.Exit(1)
	}

	minioClient, err := minio.New(minioClientCfg.Endpoint, minioClientCfg.AccessKeyID, minioClientCfg.SecretAccessKey, minioClientCfg.UseSSL)
	if err != nil {
		log.Error("failed to init minio client", slog.String("error", err.Error()))
		os.Exit(1)
	}

	minioStorage := media.New(log, minioClient)

	trackUploaderService := upload.New(log, storage, minioStorage, minioStorageCfg.OriginalBucket, minioStorageCfg.HLSBucket)
	trackStreamerService := stream.New(log, storage, minioStorage)

	router := chi.Setup(log, trackUploaderService, trackStreamerService)

	server := server.New(log, router, serverCfg.Host, serverCfg.Port, serverCfg.Timeout, serverCfg.Timeout, serverCfg.IdleTimeout)
	return &App{
		Storage: storage,
		Server:  server,
	}
}
