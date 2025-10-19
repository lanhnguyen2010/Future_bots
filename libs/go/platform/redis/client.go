package redis

import (
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// Config captures connection options for Redis.
type Config struct {
	Addr         string
	Username     string
	Password     string
	DB           int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

func (c Config) withDefaults() Config {
	cfg := c
	if cfg.Addr == "" {
		cfg.Addr = "localhost:6379"
	}
	if cfg.DialTimeout == 0 {
		cfg.DialTimeout = 5 * time.Second
	}
	if cfg.ReadTimeout == 0 {
		cfg.ReadTimeout = 3 * time.Second
	}
	if cfg.WriteTimeout == 0 {
		cfg.WriteTimeout = 3 * time.Second
	}
	return cfg
}

// NewClient builds a go-redis client using the provided config.
func NewClient(cfg Config) *goredis.Client {
	cfg = cfg.withDefaults()
	return goredis.NewClient(&goredis.Options{
		Addr:         cfg.Addr,
		Username:     cfg.Username,
		Password:     cfg.Password,
		DB:           cfg.DB,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	})
}
