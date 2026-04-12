package app

import (
	"eshkere/internal/config"
	"fmt"
	"time"

	redis "github.com/gomodule/redigo/redis"
)

func initRedis(cfg config.RedisConfig) (*redis.Pool, error) {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	pool := &redis.Pool{
		MaxIdle:     4,
		MaxActive:   16,
		IdleTimeout: 5 * time.Minute,
		Wait:        true,
		Dial: func() (redis.Conn, error) {
			return redis.Dial(
				"tcp",
				addr,
				redis.DialDatabase(cfg.DB),
				redis.DialPassword(cfg.Password),
				redis.DialConnectTimeout(5*time.Second),
				redis.DialReadTimeout(5*time.Second),
				redis.DialWriteTimeout(5*time.Second),
			)
		},
		TestOnBorrow: func(c redis.Conn, lastUsed time.Time) error {
			if time.Since(lastUsed) < time.Minute {
				return nil
			}

			_, err := c.Do("PING")
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
