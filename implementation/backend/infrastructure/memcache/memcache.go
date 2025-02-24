package memcache

import (
	"github.com/dgraph-io/ristretto"
)

type Memcache struct {
	Cache *ristretto.Cache
}

func NewMemcache() (*Memcache, error) {
	cache, err := ristretto.NewCache(&ristretto.Config{
		// todo update the max values
		NumCounters: 1e5,     // number of keys to track frequency of (10M).
		MaxCost:     1 << 25, // maximum cost of cache (32MB).
		BufferItems: 1,       // number of keys per Get buffer.
	})

	if err != nil {
		return nil, err
	}

	return &Memcache{
		Cache: cache,
	}, nil
}
