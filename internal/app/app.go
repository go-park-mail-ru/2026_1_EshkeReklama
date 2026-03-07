package app

import (
	"context"
	"database/sql"
	"eshkere/internal/config"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/mux"
)

type App struct {
	cfg *config.Config
	db  *sql.DB
	// TODO: closers []io.Closer
	// TODO: starters []StartAsService
}

func New(configPath string) *App {
	cfg, err := config.ReadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to read config: %v", err)
	}

	db, err := initDB(cfg.Postgres)
	if err != nil {
		log.Fatalf("Failed to init DB: %v", err)
	}

	return &App{
		cfg: cfg,
		db:  db,
	}
}

func (a *App) Run() error {
	router := mux.NewRouter()

	server := &http.Server{
		Addr:         a.cfg.HTTPServer.Listen,
		Handler:      router,
		ReadTimeout:  a.cfg.HTTPServer.ReadTimeout,
		WriteTimeout: a.cfg.HTTPServer.WriteTimeout,
	}

	serverErr := make(chan error, 1)

	go func() {
		serverErr <- server.ListenAndServe()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		if err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("http server failed: %w", err)
		}
		return nil

	case <-stop:
		ctx, cancel := context.WithTimeout(context.Background(), a.cfg.GracefulTimeout)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			return fmt.Errorf("shutdown server: %w", err)
		}

		if err := a.db.Close(); err != nil {
			return fmt.Errorf("close db: %w", err)
		}

		return nil
	}
}
