package redis

import (
	"context"
	"errors"
	baseredis "github.com/redis/go-redis/v9"
	"go.uber.org/fx"
	"strings"
	"time"
	"tugas-akhir/backend/infrastructure/config"
)

type Redis struct {
	Client *baseredis.ClusterClient
}

func NewRedis(config *config.Config) (*Redis, error) {
	hosts := strings.Split(config.RedisHosts, ",")

	opts := baseredis.ClusterOptions{
		Addrs: hosts,
	}

	if config.RedisPassword != "" {
		opts.Password = config.RedisPassword
	}

	rdb := baseredis.NewClusterClient(&opts)

	return &Redis{
		Client: rdb,
	}, nil
}

func (r *Redis) GetOrSetWithEx(
	ctx context.Context,
	key string,
	setValue string,
	expiration time.Duration,
) (string, error) {

	// Create a transaction pipeline
	txf := func(tx *baseredis.Tx) error {
		// Get the current value
		_, err := tx.Get(ctx, key).Result()

		if errors.Is(err, baseredis.Nil) {
			// Key doesn't exist, set it with expiration
			_, err = tx.TxPipelined(ctx, func(pipe baseredis.Pipeliner) error {
				pipe.Set(ctx, key, setValue, expiration)
				return nil
			})
			if err != nil {
				return err
			}
			return nil
		} else if err != nil {
			// Some other error occurred
			return err
		}

		// Key exists, nothing to do in the transaction
		return nil
	}

	// Execute the transaction with optimistic locking
	// Retry up to 10 times if there's a race condition
	for i := 0; i < 10; i++ {
		err := r.Client.Watch(ctx, txf, key)
		if err == nil {
			break
		}
		if errors.Is(err, baseredis.TxFailedErr) {
			// Optimistic lock lost, retry
			continue
		}
		return "", err
	}

	// Read the final value after transaction
	val, err := r.Client.Get(ctx, key).Result()
	if err != nil {
		return "", err
	}

	return val, nil

}

var Module = fx.Options(fx.Provide(NewRedis))
