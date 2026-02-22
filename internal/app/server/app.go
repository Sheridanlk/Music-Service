package server

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

type App struct {
	log    *slog.Logger
	server *http.Server
}

func New(log *slog.Logger, router http.Handler, host string, port int, readTimeout time.Duration, writeTimeout time.Duration, idleTimeout time.Duration) *App {
	address := host + ":" + strconv.Itoa(port)
	serv := &http.Server{
		Addr:         address,
		Handler:      router,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}

	return &App{
		log:    log,
		server: serv,
	}
}

func (a *App) Start() error {
	const op = "serverapp.Start"

	log := a.log.With(
		slog.String("op", op),
	)
	log.Info("Starting HTTP server", slog.String("address", a.server.Addr))

	return a.server.ListenAndServe()
}

func (a *App) Stop() error {
	const op = "serverapp.Stop"
	log := a.log.With(
		slog.String("op", op),
	)
	log.Info("Stopping HTTP server")

	return a.server.Shutdown(context.Background())
}
