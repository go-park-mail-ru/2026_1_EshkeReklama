package app

import (
	"database/sql"
	"fmt"
	"time"

	"eshkere/internal/config"

	redis "github.com/gomodule/redigo/redis"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func initDB(cfg config.PostgresConfig) (*sql.DB, error) {
	dataSource := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
		cfg.Host,
		cfg.Username,
		cfg.Password,
		cfg.Database,
		cfg.Port,
	)

	db, err := sql.Open("pgx", dataSource)
	if err != nil {
		return nil, fmt.Errorf("create pool of connections to database: %w", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}

	return db, nil
}

func initRedis(cfg config.RedisConfig) (*redis.Pool, error) {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	dialOpts := []redis.DialOption{
		redis.DialConnectTimeout(cfg.ConnectTimeout),
		redis.DialReadTimeout(cfg.ReadTimeout),
		redis.DialWriteTimeout(cfg.WriteTimeout),
	}
	if cfg.Password != "" {
		dialOpts = append(dialOpts, redis.DialPassword(cfg.Password))
	}
	if cfg.DB > 0 {
		dialOpts = append(dialOpts, redis.DialDatabase(cfg.DB))
	}

	pool := &redis.Pool{
		MaxIdle:     cfg.MaxIdle,
		MaxActive:   cfg.MaxActive,
		IdleTimeout: cfg.IdleTimeout,
		Wait:        cfg.Wait,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", addr, dialOpts...)
		},
		TestOnBorrow: func(conn redis.Conn, lastUsed time.Time) error {
			if cfg.PingAfterIdleFor <= 0 || time.Since(lastUsed) < cfg.PingAfterIdleFor {
				return nil
			}
			_, err := conn.Do("PING")
			return err
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
