package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	grpcServer := grpc.NewServer()
	profilev1.RegisterProfileServiceServer(grpcServer, profileserver.New(svc))

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
	}()

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
