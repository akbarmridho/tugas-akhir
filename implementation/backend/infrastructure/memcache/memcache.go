package memcache

import (
	"github.com/allegro/bigcache"
	"go.uber.org/fx"
	"time"
)

type Memcache struct {
	Cache *bigcache.BigCache
}

func NewMemcache() (*Memcache, error) {
	config := bigcache.Config{
		HardMaxCacheSize: 256,
		LifeWindow:       5 * time.Minute,
	}

	cache, err := bigcache.NewBigCache(config)

	if err != nil {
		return nil, err
	}

	return &Memcache{
		Cache: cache,
	}, nil
}

var Module = fx.Options(fx.Provide(NewMemcache))
