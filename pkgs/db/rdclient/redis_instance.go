package rdclient

import (
	"context"
	"fmt"
	"log/slog"
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

func NewRedisClient(cfg *config.AppConfig) (RedisClient, error) {
	if len(cfg.Redis.Address) == 0 {
		slog.Error("redis address list is empty")
		return nil, fmt.Errorf("redis address list is empty")
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
			return nil, fmt.Errorf("redis cluster mode requires at least one node address")
		}
	case "sentinel": // Sentinel mode requires master name and sentinel addresses
		if cfg.Redis.MasterName == "" {
			slog.Error("redis sentinel mode requires a master name")
			return nil, fmt.Errorf("redis sentinel mode requires a master name")
		}
		connOpts.MasterName = cfg.Redis.MasterName
	default:
		if len(connOpts.Addrs) == 0 { // Treat any other value as standalone
			slog.Error("redis standalone mode requires a node address")
			return nil, fmt.Errorf("redis standalone mode requires a node address")
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
		return nil, fmt.Errorf("redis client ping failed: %w", err)
	}

	slog.Info("redis connection created",
		slog.String("mode", cfg.Redis.Mode),
		slog.Any("addresses", cfg.Redis.Address),
		slog.String("master_name", cfg.Redis.MasterName),
	)

	return &redisClient{
		_client: client,
	}, nil
}
