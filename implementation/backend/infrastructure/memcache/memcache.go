package memcache

import (
	"github.com/allegro/bigcache"
)

type Memcache struct {
	Cache *bigcache.BigCache
}

func NewMemcache() (*Memcache, error) {
	config := bigcache.Config{
		HardMaxCacheSize: 256,
	}

	cache, err := bigcache.NewBigCache(config)

	if err != nil {
		return nil, err
	}

	return &Memcache{
		Cache: cache,
	}, nil
}
