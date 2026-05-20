package rdclient

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/tdatIT/go-template/config"
)

type RedisClient interface {
	Client() redis.UniversalClient
	Close() error
}

type redisClient struct {
	_client redis.UniversalClient
}

func (r *redisClient) Client() redis.UniversalClient {
	return r._client
}

func (r *redisClient) Close() error {
	return r._client.Close()
}

func NewRedisClient(cfg *config.AppConfig) RedisClient {
	if len(cfg.Redis.Address) == 0 {
		slog.Error("redis address list is empty")
		os.Exit(1)
	}

	connOpts := redis.UniversalOptions{
		Addrs:        cfg.Redis.Address,
		Username:     cfg.Redis.Username,
		Password:     cfg.Redis.Password,
		PoolSize:     cfg.Redis.PoolSize,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		MaxRetries:   3,
	}

	switch cfg.Redis.Mode {
	case "cluster": // Cluster mode relies on multiple node addresses
		if len(connOpts.Addrs) == 0 {
			slog.Error("redis cluster mode requires at least one node address")
			os.Exit(1)
		}
	case "sentinel": // Sentinel mode requires master name and sentinel addresses
		if cfg.Redis.MasterName == "" {
			slog.Error("redis sentinel mode requires a master name")
			os.Exit(1)
		}
		connOpts.MasterName = cfg.Redis.MasterName
	default:
		if len(connOpts.Addrs) == 0 { // Treat any other value as standalone
			slog.Error("redis standalone mode requires a node address")
		}
		// Ensure only the primary address is used for standalone setups
		connOpts.Addrs = []string{connOpts.Addrs[0]}
	}

	client := redis.NewUniversalClient(&connOpts)

	// Test the connection with a ping to ensure it's working
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		slog.Error("redis client ping failed", slog.String("error", err.Error()))
	}

	slog.Info("redis client connected successfully")

	return &redisClient{
		_client: client,
	}
}
