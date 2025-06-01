package memcache

import (
	gocache "github.com/patrickmn/go-cache"
	"go.uber.org/fx"
	"time"
)

type Memcache struct {
	Cache *gocache.Cache
}

func NewMemcache() (*Memcache, error) {
	cache := gocache.New(15*time.Minute, 10*time.Minute)

	return &Memcache{
		Cache: cache,
	}, nil
}

var Module = fx.Options(fx.Provide(NewMemcache))
