package redis

import (
	baseredis "github.com/redis/go-redis/v9"
	"go.uber.org/fx"
	"strings"
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

var Module = fx.Options(fx.Provide(NewRedis))
