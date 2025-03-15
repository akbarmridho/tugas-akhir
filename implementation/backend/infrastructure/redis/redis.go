package redis

import baseredis "github.com/redis/go-redis/v9"

type Cluster struct {
	Client *baseredis.ClusterClient
}

type Config struct {
	Hosts    []string
	Password *string
}

func NewRedis(config Config) (*Cluster, error) {
	opts := baseredis.ClusterOptions{
		Addrs: config.Hosts,
	}

	if config.Password != nil {
		opts.Password = *config.Password
	}

	rdb := baseredis.NewClusterClient(&opts)

	return &Cluster{
		Client: rdb,
	}, nil
}
