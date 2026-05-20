package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"eshkere/internal/observability"

	profileredis "eshkere/internal/profile/repository/redis"
	profileserver "eshkere/internal/profile/server"
	profileservice "eshkere/internal/profile/service"
	profilev1 "eshkere/pkg/pb/profile/v1"

	redis "github.com/gomodule/redigo/redis"
	"google.golang.org/grpc"
)

func main() {
	addr := envOr("PROFILE_GRPC_ADDR", ":50052")
	redisAddr := envOr("PROFILE_REDIS_ADDR", "localhost:6379")
	redisPassword := envOr("PROFILE_REDIS_PASSWORD", "")
	metricsAddr := envOr("PROFILE_METRICS_ADDR", "")

	redisPool, err := initRedis(redisAddr, redisPassword)
	if err != nil {
		log.Fatalf("redis: %v", err)
	}
	defer redisPool.Close()

	repo := profileredis.New(redisPool)
	svc, err := profileservice.New(repo)
	if err != nil {
		log.Fatalf("profile service: %v", err)
	}

	metrics := observability.NewMetrics("profile")
	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(metrics.UnaryServerInterceptor()))
	profilev1.RegisterProfileServiceServer(grpcServer, profileserver.New(svc))
	metricsServer := observability.NewMetricsServer(metricsAddr, metrics)

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("listen %s: %v", addr, err)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-quit
		log.Println("shutting down profile gRPC server...")
		grpcServer.GracefulStop()
		if metricsServer != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := metricsServer.Shutdown(ctx); err != nil && err != http.ErrServerClosed {
				log.Printf("shutdown profile metrics server: %v", err)
			}
		}
	}()

	if metricsServer != nil {
		go func() {
			log.Printf("profile metrics listening on %s", metricsAddr)
			if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("metrics serve: %v", err)
			}
		}()
	}

	log.Printf("profile gRPC listening on %s", addr)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("serve: %v", err)
	}
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
		_ = pool.Close()
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
