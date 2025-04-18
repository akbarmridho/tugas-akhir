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
		Shards:             1024,
		HardMaxCacheSize:   256,
		LifeWindow:         5 * time.Minute,
		CleanWindow:        5 * time.Minute,
		MaxEntriesInWindow: 1000 * 10 * 60,
		MaxEntrySize:       10000, // 10kb
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
