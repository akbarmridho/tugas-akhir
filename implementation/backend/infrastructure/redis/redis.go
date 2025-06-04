package redis

import (
	"context"
	"errors"
	"net"
	"strings"
	"time"
	"tugas-akhir/backend/infrastructure/config"
	"tugas-akhir/backend/pkg/logger"

	errors2 "github.com/pkg/errors"
	baseredis "github.com/redis/go-redis/v9"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var RedisUnhealthy = errors.New("redis cluster unhealthy")

type Redis struct {
	Client *baseredis.ClusterClient
}

func NewRedis(config *config.Config) (*Redis, error) {
	hosts := strings.Split(config.RedisHosts, ",")

	opts := baseredis.ClusterOptions{
		Addrs:        hosts,
		PoolSize:     500,
		MinIdleConns: 20,

		DialTimeout:  10 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		PoolTimeout:  6 * time.Second,
	}

	if config.RedisPassword != "" {
		opts.Password = config.RedisPassword
	}

	if len(config.RedisHostsMap) != 0 {
		hostsMap := strings.Split(config.RedisHostsMap, ",")

		mapping := make(map[string]string)

		for i, mapped := range hostsMap {
			mapping[mapped] = hosts[i]
		}

		baseDialer := &net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 5 * time.Minute,
		}

		// Create the custom dialer function using the map from the cluster setup
		customDialer := func(ctx context.Context, network, addr string) (net.Conn, error) {

			if !strings.Contains(addr, "localhost") {
				externalAddr, found := mapping[addr]

				if found {
					addr = externalAddr // Use the mapped external address
				} else {
					return nil, errors2.Errorf("go-redis Dialer: Dialing address '%s' directly (not found in internal map)", addr)
				}
			}

			return baseDialer.DialContext(ctx, network, addr)
		}

		opts.Dialer = customDialer
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

func (r *Redis) IsHealthy(baseCtx context.Context) error {
	l := logger.FromCtx(baseCtx).With(zap.String("service", "redis-health-check"))

	ctx, cancel := context.WithTimeout(baseCtx, 5*time.Second)
	defer cancel()

	_, err := r.Client.Ping(ctx).Result()
	if err != nil {
		return err
	}

	clusterInfo, err := r.Client.ClusterInfo(ctx).Result()

	if err != nil {
		return err
	}

	if !strings.Contains(clusterInfo, "cluster_state:ok") {
		l.Warn(clusterInfo)
		return errors2.WithMessage(RedisUnhealthy, "cluster state is not ok")
	}

	clusterNodes, err := r.Client.ClusterNodes(ctx).Result()
	if err != nil {
		return err
	}

	lines := strings.Split(clusterNodes, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		parts := strings.Fields(line)

		if len(parts) < 3 {
			continue
		}

		flags := strings.Split(parts[2], ",")

		// Check if node is a master (should be master since we only have masters)
		isMaster := false
		isFailing := false

		for _, flag := range flags {
			if flag == "master" {
				isMaster = true
			}
			if flag == "fail" || flag == "fail?" {
				isFailing = true
			}
		}

		// If node is failing or not a master, the cluster is unhealthy
		if isFailing || !isMaster {
			l.Warn(clusterNodes)
			return errors2.WithMessage(RedisUnhealthy, "master node is failing")
		}
	}

	return nil
}

func (r *Redis) Stop() error {
	return r.Client.Close()
}

var Module = fx.Options(
	fx.Provide(fx.Annotate(NewRedis,
		fx.OnStop(func(r *Redis) error {
			return r.Stop()
		}),
	)),
)
