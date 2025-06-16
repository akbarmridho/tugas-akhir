package sanity

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	baseredis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"strconv"
	"strings"
	"sync"
	"tugas-akhir/backend/infrastructure/memcache"
	"tugas-akhir/backend/infrastructure/redis"
	"tugas-akhir/backend/internal/events/entity"
	"tugas-akhir/backend/internal/events/repository/availability"
	"tugas-akhir/backend/internal/orders/service/early_dropper"
	"tugas-akhir/backend/pkg/logger"
)

type RedisCheck struct {
	redis *redis.Redis
	cache *memcache.Memcache
}

const dropperCacheKey = "dropper-keys"

// GetAvailability performs a sanity check on seat availability for a given list of ticket sale IDs.
// It fetches data directly from Redis Hashes for high performance.
func (s *RedisCheck) GetAvailability(ctx context.Context, ticketSaleIDs []int64) (*AvailabilityCheck, error) {
	l := logger.FromCtx(ctx)
	result := &AvailabilityCheck{}

	if len(ticketSaleIDs) == 0 {
		return result, nil // Return empty result if no IDs are provided.
	}

	// Use a pipeline to fetch all HGetAll results in a single round-trip.
	pipe := s.redis.Client.Pipeline()
	cmds := make(map[int64]*baseredis.MapStringStringCmd, len(ticketSaleIDs))

	for _, id := range ticketSaleIDs {
		key := availability.CacheKey(id)
		cmds[id] = pipe.HGetAll(ctx, key)
	}

	_, err := pipe.Exec(ctx)
	// We can ignore a general `redis.Nil` error from Exec, as it might just mean some keys didn't exist.
	// We'll check the error for each individual command below.
	if err != nil && !errors.Is(err, baseredis.Nil) {
		l.Error("failed to execute availability pipeline", zap.Error(err))
		return nil, err
	}

	// Process the results from the pipeline.
	for _, cmd := range cmds {
		fields, err := cmd.Result()
		if err != nil {
			// If a specific hash doesn't exist, it's not an error for a sanity check, just skip it.
			if errors.Is(err, baseredis.Nil) {
				continue
			}
			l.Warn("failed to get command result from availability pipeline", zap.Error(err))
			continue // Continue checking other results
		}

		for field, value := range fields {
			parts := strings.Split(field, ":")
			if len(parts) != 3 {
				continue // Ignore malformed fields
			}

			// The last part determines if it's total or available seats.
			fieldType := parts[2]
			seatCount, convErr := strconv.ParseInt(value, 10, 32)
			if convErr != nil {
				l.Warn("could not parse seat count", zap.String("field", field), zap.String("value", value), zap.Error(convErr))
				continue
			}

			switch fieldType {
			case "total":
				result.Count += int(seatCount)
			case "available":
				result.Available += int(seatCount)
			}
		}
	}

	result.Unavailable = result.Count - result.Available
	return result, nil
}

func (s *RedisCheck) GetDropperAvailability(ctx context.Context) (*AvailabilityCheck, error) {
	l := logger.FromCtx(ctx).With(zap.String("func", "dropper_availability"))

	pattern := fmt.Sprintf("%s*", early_dropper.DropperRedisPrefix)

	keys := make([]string, 0)

	getKeys := func() error {
		var mu sync.Mutex

		// ForEachMaster will loop over every master node in the cluster.
		err := s.redis.Client.ForEachMaster(ctx, func(ctx context.Context, client *baseredis.Client) error {
			var cursor uint64 = 0
			var localKeys []string
			for {
				// Scan on the current node.
				ckeys, nextCursor, err := client.Scan(ctx, cursor, pattern, 100).Result()
				if err != nil {
					return err
				}

				localKeys = append(localKeys, ckeys...)

				if nextCursor == 0 {
					break
				}
				cursor = nextCursor
			}

			mu.Lock()
			keys = append(keys, localKeys...)
			mu.Unlock()

			return nil
		})

		if err != nil {
			return err
		}

		raw, err := json.Marshal(keys)

		if err != nil {
			logger.FromCtx(ctx).Error("Cannot marshall keys", zap.Error(err))
			return err
		}

		s.cache.Cache.SetDefault(dropperCacheKey, raw)

		return nil
	}

	cachedData, found := s.cache.Cache.Get(dropperCacheKey)

	var getKeysError error

	if found {
		rawBytes, typeOk := cachedData.([]byte) // bigcache Get returns []byte, go-cache Get returns interface{}
		if !typeOk {
			l.Error("Cached data for dropper keys is not []byte", zap.String("key", dropperCacheKey))
			// Treat as cache miss/corruption if type is wrong, and refetch
			getKeysError = getKeys()
		} else {
			marshallErr := json.Unmarshal(rawBytes, &keys)
			if marshallErr != nil {
				l.Error("Cannot unmarshal cached dropper keys", zap.Error(marshallErr))
				// Original logic: Do not set getKeysError = marshallErr here directly.
				// The len(keys) == 0 check below is the primary trigger for getKeys()
				// if unmarshalling fails and results in empty keys.
			}

			// This check mirrors the original logic: if after attempting to load from cache,
			// 'keys' is empty, then refetch. This covers:
			// 1. Successful unmarshal of an empty list from cache.
			// 2. Failed unmarshal where 'keys' remains (or becomes) empty.
			if len(keys) == 0 {
				if marshallErr != nil { // Add context if unmarshal error led to empty keys
					l.Warn("Unmarshaling cached dropper keys failed and resulted in empty keys, refetching.", zap.Error(marshallErr), zap.String("key", dropperCacheKey))
				} else {
					l.Warn("Cached dropper keys are an empty list, refetching.", zap.String("key", dropperCacheKey))
				}
				getKeysError = getKeys()
			}
		}
	} else { // Not found in cache (equivalent to bigcache.ErrEntryNotFound)
		getKeysError = getKeys()
	}

	if getKeysError != nil {
		return nil, getKeysError
	}

	//l.Sugar().Infof("got %d keys to scan", len(keys))

	result := AvailabilityCheck{
		Count:       0,
		Available:   0,
		Unavailable: 0,
	}

	buffer := make([]string, 0)
	bufferCount := 400

	batchCheck := func() error {
		//defer func() {
		//	l.Sugar().Infof("current result: total %d, available %d, unavailable %d", result.Count, result.Available, result.Unavailable)
		//}()

		pipe := s.redis.Client.Pipeline()

		cmds := make(map[string]*baseredis.StringCmd)

		for _, key := range buffer {
			// wrong key
			if key == "early-dropper:refresher-node" {
				continue
			}

			cmds[key] = pipe.Get(ctx, key)
		}

		_, err := pipe.Exec(ctx)

		if err != nil && !errors.Is(err, baseredis.Nil) {
			return err
		}

		freeSeatAvailableTotal := 0
		freeSeatCheck := make([]string, 0)

		for key, cmd := range cmds {
			val, resultErr := cmd.Result()
			if resultErr != nil {
				l.Sugar().Warnf("failed to get key %s: %v", key, err)
				continue
			}

			if strings.Contains(key, "numbered") {
				// numbered set
				if val == string(entity.SeatStatus__Available) {
					result.Available += 1
				} else {
					result.Unavailable += 1
				}

				result.Count += 1
			} else {
				// free seat
				availableCount, parseErr := strconv.ParseInt(val, 10, 32)

				if parseErr != nil {
					l.Sugar().Warnf("cannot parse value for key %s with value %s", key, val)
					continue
				}

				result.Available += int(availableCount)
				freeSeatAvailableTotal += int(availableCount)
				freeSeatCheck = append(freeSeatCheck, key)
			}
		}

		buffer = make([]string, 0)

		if len(freeSeatCheck) == 0 {
			return nil
		}

		pipe = s.redis.Client.Pipeline()
		cmds = make(map[string]*baseredis.StringCmd)

		for _, key := range freeSeatCheck {
			debugKey := fmt.Sprintf("debug:%s", key)
			cmds[debugKey] = pipe.Get(ctx, debugKey)
		}

		_, err = pipe.Exec(ctx)

		if err != nil && !errors.Is(err, baseredis.Nil) {
			return err
		}

		freeSeatCount := 0

		for key, cmd := range cmds {
			val, resultErr := cmd.Result()
			if resultErr != nil {
				l.Sugar().Warnf("failed to get key %s: %v", key, err)
				continue
			}

			initialCount, parseErr := strconv.ParseInt(val, 10, 32)

			if parseErr != nil {
				l.Sugar().Warnf("cannot parse value for key %s with value %s", key, val)
				continue
			}

			freeSeatCount += int(initialCount)
		}

		result.Count += freeSeatCount
		result.Unavailable += freeSeatCount - freeSeatAvailableTotal

		return nil
	}

	for _, key := range keys {
		buffer = append(buffer, key)

		if len(buffer) >= bufferCount {
			batchErr := batchCheck()

			if batchErr != nil {
				l.Sugar().Error("batch check error", zap.Error(batchErr))
			}
		}
	}

	if len(buffer) > 0 {
		batchErr := batchCheck()

		if batchErr != nil {
			l.Sugar().Error("batch check error", zap.Error(batchErr))
		}
	}

	return &result, nil
}
