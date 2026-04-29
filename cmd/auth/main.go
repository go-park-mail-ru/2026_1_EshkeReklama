package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	redis "github.com/gomodule/redigo/redis"
	_ "github.com/jackc/pgx/v5/stdlib"
	"google.golang.org/grpc"

	authrepo "eshkere/internal/auth/repository/postgres"
	authsession "eshkere/internal/auth/session"
	authserver "eshkere/internal/auth/server"
	authsvc "eshkere/internal/auth/service"
	authv1 "eshkere/pkg/pb/auth/v1"
)

func main() {
	addr := envOr("AUTH_GRPC_ADDR", ":50051")
	pgDSN := envOr("AUTH_PG_DSN", "postgres://eshkere:eshkere@localhost:5432/eshkere?sslmode=disable")
	redisAddr := envOr("AUTH_REDIS_ADDR", "localhost:6379")
	sessionTTL := envDuration("AUTH_SESSION_TTL", 24*time.Hour)

	db, err := initPostgres(pgDSN)
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	defer db.Close()

	redisPool, err := initRedis(redisAddr)
	if err != nil {
		log.Fatalf("redis: %v", err)
	}
	defer redisPool.Close()

	credsRepo := authrepo.NewCredentialsRepository(db)
	credsSvc := authsvc.NewCredentialsService(credsRepo)

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

func initRedis(addr string) (*redis.Pool, error) {
	pool := &redis.Pool{
		MaxIdle:     10,
		MaxActive:   100,
		IdleTimeout: 240 * time.Second,
		Wait:        true,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", addr,
				redis.DialConnectTimeout(5*time.Second),
				redis.DialReadTimeout(3*time.Second),
				redis.DialWriteTimeout(3*time.Second),
			)
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
