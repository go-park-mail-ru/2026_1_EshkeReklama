package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"eshkere/internal/auth/vkid"
	redis "github.com/gomodule/redigo/redis"
	_ "github.com/jackc/pgx/v5/stdlib"
	"google.golang.org/grpc"

	authrepo "eshkere/internal/auth/repository/postgres"
	authserver "eshkere/internal/auth/server"
	authsvc "eshkere/internal/auth/service"
	authsession "eshkere/internal/auth/session"
	authv1 "eshkere/pkg/pb/auth/v1"
)

func main() {
	addr := envOr("AUTH_GRPC_ADDR", ":50051")
	pgDSN := envOr("AUTH_PG_DSN", "postgres://eshkere:eshkere@localhost:5432/eshkere?sslmode=disable")
	redisAddr := envOr("AUTH_REDIS_ADDR", "localhost:6379")
	redisPassword := envOr("AUTH_REDIS_PASSWORD", "")
	sessionTTL := envDuration("AUTH_SESSION_TTL", 24*time.Hour)
	vkIDClientID := envInt64("AUTH_VKID_CLIENT_ID", 0)
	vkIDRedirectURI := envOr("AUTH_VKID_REDIRECT_URI", "")
	vkIDDomain := envOr("AUTH_VKID_DOMAIN", "id.vk.ru")
	vkIDTimeout := envDuration("AUTH_VKID_TIMEOUT", 5*time.Second)

	db, err := initPostgres(pgDSN)
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	defer db.Close()

	redisPool, err := initRedis(redisAddr, redisPassword)
	if err != nil {
		log.Fatalf("redis: %v", err)
	}
	defer redisPool.Close()

	credsRepo := authrepo.NewCredentialsRepository(db)
	credsSvc := authsvc.NewCredentialsService(credsRepo, initVKIDClient(vkIDClientID, vkIDRedirectURI, vkIDDomain, vkIDTimeout))

	store := authsession.NewRedisStore(redisPool)
	sessionMgr := authsession.NewManager(store, sessionTTL)

	grpcServer := grpc.NewServer()
	authv1.RegisterAuthServiceServer(grpcServer, authserver.New(credsSvc, sessionMgr))

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("listen %s: %v", addr, err)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-quit
		log.Println("shutting down auth gRPC server...")
		grpcServer.GracefulStop()
	}()

	log.Printf("auth gRPC listening on %s", addr)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("serve: %v", err)
	}
}

func initPostgres(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}
	return db, nil
}

func initRedis(addr, password string) (*redis.Pool, error) {
	dialOpts := []redis.DialOption{
		redis.DialConnectTimeout(5 * time.Second),
		redis.DialReadTimeout(3 * time.Second),
		redis.DialWriteTimeout(3 * time.Second),
	}
	if password != "" {
		dialOpts = append(dialOpts, redis.DialPassword(password))
	}

	pool := &redis.Pool{
		MaxIdle:     10,
		MaxActive:   100,
		IdleTimeout: 240 * time.Second,
		Wait:        true,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", addr, dialOpts...)
		},
	}
	conn := pool.Get()
	defer conn.Close()
	if _, err := conn.Do("PING"); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping redis: %w", err)
	}
	return pool, nil
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envDuration(key string, def time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return def
	}
	return d
}

func envInt64(key string, def int64) int64 {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return def
	}
	return n
}

func initVKIDClient(clientID int64, redirectURI, domain string, timeout time.Duration) *vkid.Client {
	if clientID <= 0 || redirectURI == "" {
		return nil
	}
	if _, err := url.ParseRequestURI(redirectURI); err != nil {
		log.Printf("vk id disabled: invalid redirect uri: %v", err)
		return nil
	}
	return vkid.New(clientID, redirectURI, domain, timeout)
}
