package redis

import baseredis "github.com/redis/go-redis/v9"

type Cluster struct {
	Client *baseredis.ClusterClient
}

type Config struct {
	Hosts []string
}

func NewRedis(config Config) (*Cluster, error) {
	rdb := baseredis.NewClusterClient(&baseredis.ClusterOptions{
		Addrs: config.Hosts,
	})

	return &Cluster{
		Client: rdb,
	}, nil
}
