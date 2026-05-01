package app

import (
	"context"
	authclient "eshkere/internal/client/auth"
	"eshkere/internal/config"
	"eshkere/internal/handler"
	middleware2 "eshkere/internal/handler/middleware"
	"eshkere/internal/handler/v1"
	"eshkere/internal/repository/postgres"
	"eshkere/internal/service"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	s3 "eshkere/internal/storage/s3"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type App struct {
	cfg        *config.Config
	logger     *zap.SugaredLogger
	closers    []io.Closer
	service    *service.Service
	authClient *authclient.Client
}

func New(configPath string) *App {
	var closers []io.Closer

	zapCfg := zap.NewDevelopmentConfig()
	zapCfg.DisableStacktrace = true

	baseLogger, err := zapCfg.Build()
	if err != nil {
		panic(err)
	}
	logger := baseLogger.Sugar()

	cfg, err := config.ReadConfig(configPath)
	if err != nil {
		logger.Fatalf("Failed to read config: %v", err)
	}

	db, err := initDB(cfg.Postgres)
	if err != nil {
		logger.Fatalf("Failed to init DB: %v", err)
	}

	closers = append([]io.Closer{db}, closers...)

	advertiserRepo := postgres.NewAdvertiserRepository(db)
	adGroupRepo := postgres.NewAdGroupRepository(db)
	adRepo := postgres.NewAdRepository(db)
	adCampaignRepo := postgres.NewAdCampaignRepository(db)
	feedLinkRepo := postgres.NewFeedLinkRepository(db)
	appealRepo := postgres.NewAppealRepository(db)

	s3Client, err := s3.NewClient(context.Background(), s3.Config{
		Region:          cfg.S3.Region,
		Bucket:          cfg.S3.Bucket,
		Endpoint:        cfg.S3.Endpoint,
		AccessKeyID:     cfg.S3.AccessKey,
		SecretAccessKey: cfg.S3.SecretKey,
		ForcePathStyle:  cfg.S3.ForcePathStyle,
		PublicBaseURL:   cfg.S3.PublicBaseURL,
	})
	if err != nil {
		logger.Fatalf("Failed to create S3 client: %v", err)
	}

	avatarStorage := s3.NewAvatarStorage(s3Client, "")
	appealStorage := s3.NewAppealStorage(s3Client)
	adStorage := s3.NewAdStorage(s3Client)

	svc, err := service.NewService(&service.Config{
		AdvertiserRepo:  advertiserRepo,
		PartnerRepo:     nil,
		PartnerSiteRepo: nil,
		AdCampaignRepo:  adCampaignRepo,
		AdGroupRepo:     adGroupRepo,
		AdRepo:          adRepo,
		FeedLinkRepo:    feedLinkRepo,
		AppealRepo:      appealRepo,
		AvatarStorage:   avatarStorage,
		AppealStorage:   appealStorage,
		AdStorage:       adStorage,
		AdActionRepo:    nil,
		TopicRepo:       nil,
		RegionRepo:      nil,
	})
	if err != nil {
		logger.Fatalf("Failed to init service: %v", err)
	}

	authAddr := cfg.AuthService.GRPCAddr

	ac, err := authclient.New(authAddr)
	if err != nil {
		logger.Fatalf("Failed to connect to auth service: %v", err)
	}
	closers = append(closers, ac)

	return &App{
		cfg:        cfg,
		logger:     logger,
		closers:    closers,
		service:    svc,
		authClient: ac,
	}
}

func (a *App) Run() error {
	router := mux.NewRouter().StrictSlash(true)
	router.Use(middleware2.RequestContext(a.logger))
	router.Use(middleware2.AccessLog())
	router.Use(middleware2.CSRF(middleware2.CSRFConfig{
		CookieName: "csrf_token",
		HeaderName: "X-CSRF-Token",
		Secure:     a.cfg.Session.CookieSecure,
	}))

	handler.Register(router, v1.NewAPI(v1.APIConfig{
		Service:    a.service,
		AuthClient: a.authClient,
		CookieConfig: v1.CookieConfig{
			Name:     a.cfg.Session.CookieName,
			Path:     a.cfg.Session.CookiePath,
			HTTPOnly: true,
			Secure:   a.cfg.Session.CookieSecure,
		},
	}))

	server := &http.Server{
		Addr:         a.cfg.HTTPServer.Listen,
		Handler:      middleware2.CORS(a.cfg.CORS.AllowedOrigins)(router),
		ReadTimeout:  a.cfg.HTTPServer.ReadTimeout,
		WriteTimeout: a.cfg.HTTPServer.WriteTimeout,
	}

	serverErr := make(chan error, 1)

	go func() {
		a.logger.Infow("server started", "addr", server.Addr)
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
