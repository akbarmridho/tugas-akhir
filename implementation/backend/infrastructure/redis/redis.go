package redis

import baseredis "github.com/redis/go-redis/v9"

type Cluster struct {
	Client *baseredis.ClusterClient
}

func NewRedis(config Config) (*Cluster, error) {
	rdb := baseredis.NewClusterClient(&baseredis.ClusterOptions{
		// todo
	})

	return &Cluster{
		Client: rdb,
	}, nil
}
