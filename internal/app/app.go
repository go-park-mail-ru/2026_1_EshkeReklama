package app

import (
	"context"
	"eshkere/internal/config"
	handlers "eshkere/internal/handler"
	"eshkere/internal/middleware"
	"eshkere/internal/session"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
)

type App struct {
	cfg     *config.Config
	closers []io.Closer
}

func New(configPath string) *App {
	var closers []io.Closer

	cfg, err := config.ReadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to read config: %v", err)
	}

	db, err := initDB(cfg.Postgres)
	if err != nil {
		log.Fatalf("Failed to init DB: %v", err)
	}

	closers = append([]io.Closer{db}, closers...)

	return &App{
		cfg:     cfg,
		closers: closers,
	}
}

func (a *App) Run() error {
	sessionManager := session.NewManager()
	sessionManager.StartCleanup(5 * time.Minute)

	router := mux.NewRouter().StrictSlash(true)
	handlers.Register(router, handlers.NewAPI(handlers.APIConfig{
		SessionManager: sessionManager,
	}))

	server := &http.Server{
		Addr:         a.cfg.HTTPServer.Listen,
		Handler:      middleware.CORS(a.cfg.CORS.AllowedOrigins)(router),
		ReadTimeout:  a.cfg.HTTPServer.ReadTimeout,
		WriteTimeout: a.cfg.HTTPServer.WriteTimeout,
	}

	serverErr := make(chan error, 1)

	go func() {
		log.Printf("server started on %s", server.Addr)
		serverErr <- server.ListenAndServe()
	}()

	return a.waitShutdown(server, serverErr)
}

func (a *App) waitShutdown(server *http.Server, serverErr <-chan error) error {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(stop)

	select {
	case err := <-serverErr:
		if err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("http server failed: %w", err)
		}
		return nil

	case <-stop:
		return a.shutdown(server)
	}
}

func (a *App) shutdown(server *http.Server) error {
	ctx, cancel := context.WithTimeout(context.Background(), a.cfg.GracefulTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown server: %w", err)
	}

	for _, c := range a.closers {
		if err := c.Close(); err != nil {
			fmt.Println("failed to close:", c)
		}
	}

	return nil
}
