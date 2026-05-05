package app

import (
	"context"
	authclient "eshkere/internal/client/auth"
	profileclient "eshkere/internal/client/profile"
	"eshkere/internal/config"
	"eshkere/internal/handler"
	middleware2 "eshkere/internal/handler/middleware"
	"eshkere/internal/handler/v1"
	"eshkere/internal/repository/postgres"
	redisrepo "eshkere/internal/repository/redis"
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
	cfg           *config.Config
	logger        *zap.SugaredLogger
	closers       []io.Closer
	service       *service.Service
	authClient    *authclient.Client
	profileClient *profileclient.Client
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

	redisPool, err := initRedis(cfg.Redis)
	if err != nil {
		logger.Fatalf("Failed to init Redis: %v", err)
	}
	closers = append(closers, redisPool)

	advertiserRepo := postgres.NewAdvertiserRepository(db)
	partnerRepo := postgres.NewPartnerRepository(db)
	partnerSiteRepo := postgres.NewPartnerSiteRepository(db)
	partnerBlockRepo := postgres.NewPartnerBlockRepository(db)
	partnerBlockGeoRuleRepo := postgres.NewPartnerBlockGeoRuleRepository(db)
	adGroupRepo := postgres.NewAdGroupRepository(db)
	adRepo := postgres.NewAdRepository(db)
	adCampaignRepo := postgres.NewAdCampaignRepository(db)
	feedLinkRepo := postgres.NewFeedLinkRepository(db)
	appealRepo := postgres.NewAppealRepository(db)
	adRequestStore := redisrepo.NewAdRequestStore(redisPool)

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
		AdvertiserRepo:          advertiserRepo,
		PartnerRepo:             partnerRepo,
		PartnerSiteRepo:         partnerSiteRepo,
		PartnerBlockRepo:        partnerBlockRepo,
		PartnerBlockGeoRuleRepo: partnerBlockGeoRuleRepo,
		AdCampaignRepo:          adCampaignRepo,
		AdGroupRepo:             adGroupRepo,
		AdRepo:                  adRepo,
		FeedLinkRepo:            feedLinkRepo,
		AppealRepo:              appealRepo,
		AvatarStorage:           avatarStorage,
		AppealStorage:           appealStorage,
		AdStorage:               adStorage,
		AdActionRepo:            nil,
		TopicRepo:               nil,
		RegionRepo:              nil,
		ProfileClient:           nil,
		AdRequestStore:          adRequestStore,
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

	pc, err := profileclient.New(cfg.ProfileService.GRPCAddr)
	if err != nil {
		logger.Fatalf("Failed to connect to profile service: %v", err)
	}
	closers = append(closers, pc)
	svc.SetProfileClient(pc)

	return &App{
		cfg:           cfg,
		logger:        logger,
		closers:       closers,
		service:       svc,
		authClient:    ac,
		profileClient: pc,
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
		SkipPaths:  []string{"/ad/request"},
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
		VKIDConfig: v1.VKIDConfig{
			ClientID:           a.cfg.VKID.ClientID,
			RedirectURI:        a.cfg.VKID.RedirectURI,
			AuthDomain:         a.cfg.VKID.AuthDomain,
			Scope:              a.cfg.VKID.Scope,
			DefaultRedirectURL: a.cfg.VKID.DefaultRedirectURL,
			ErrorRedirectURL:   a.cfg.VKID.ErrorRedirectURL,
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
