package sanity

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/allegro/bigcache"
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

const availabilityCacheKey = "availability-keys"
const dropperCacheKey = "dropper-keys"

func (s *RedisCheck) GetAvailability(ctx context.Context) (*AvailabilityCheck, error) {
	l := logger.FromCtx(ctx)

	pattern := fmt.Sprintf("%s:*:*:*:*", availability.AvailabilityPrefix)

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

		if setCacheErr := s.cache.Cache.Set(availabilityCacheKey, raw); setCacheErr != nil {
			logger.FromCtx(ctx).Error("Cannot set cache keys", zap.Error(setCacheErr))
			return setCacheErr
		}

		return nil
	}

	cache, cacheErr := s.cache.Cache.Get(availabilityCacheKey)

	var getKeysError error

	if cacheErr != nil {
		if !errors.Is(cacheErr, bigcache.ErrEntryNotFound) {
			logger.FromCtx(ctx).Error("Cannot get keys from cache", zap.Error(cacheErr))
		} else {
			getKeysError = getKeys()
		}
	} else {
		marshallErr := json.Unmarshal(cache, &keys)

		if marshallErr != nil {
			logger.FromCtx(ctx).Error("Cannot unmashall cached keys")

			getKeysError = marshallErr
		}
	}

	if getKeysError != nil {
		return nil, getKeysError
	}

	// Group keys by their area identifiers
	areaMap := make(map[string][]string)
	for _, key := range keys {
		parts := strings.Split(key, ":")
		if len(parts) != 5 {
			l.Sugar().Warnf("parts count not 5 for key %s", key)
			continue
		}

		// Create area identifier (saleID:packageID:areaID)
		areaIdentifier := fmt.Sprintf("%s:%s:%s", parts[1], parts[2], parts[3])
		areaMap[areaIdentifier] = append(areaMap[areaIdentifier], key)
	}

	// Get all values at once if there are keys
	if len(keys) == 0 {
		return nil, entity.AreaAvailabilityNotFoundError
	}

	keyToValue := make(map[string]string)

	pipe := s.redis.Client.Pipeline()

	cmds := make(map[string]*baseredis.StringCmd)

	for _, key := range keys {
		cmds[key] = pipe.Get(ctx, key)
	}

	_, err := pipe.Exec(ctx)

	if err != nil && !errors.Is(err, baseredis.Nil) {
		return nil, err
	}

	for key, cmd := range cmds {
		val, resultErr := cmd.Result()
		if resultErr != nil {
			l.Sugar().Warnf("failed to get key %s: %v", key, err)
			continue
		}
		keyToValue[key] = val
	}

	// Build the result
	result := AvailabilityCheck{
		Count:       0,
		Available:   0,
		Unavailable: 0,
	}

	for areaIdentifier, _ := range areaMap {
		parts := strings.Split(areaIdentifier, ":")
		if len(parts) != 3 {
			l.Sugar().Warnf("parts count not 3 for key %s", areaIdentifier)
			continue
		}

		saleID, _ := strconv.ParseInt(parts[0], 10, 64)
		packageID, _ := strconv.ParseInt(parts[1], 10, 64)
		areaID, _ := strconv.ParseInt(parts[2], 10, 64)

		// Construct keys for total and available
		totalKey := fmt.Sprintf("%s:%d:%d:%d:total", availability.AvailabilityPrefix, saleID, packageID, areaID)
		availableKey := fmt.Sprintf("%s:%d:%d:%d:available", availability.AvailabilityPrefix, saleID, packageID, areaID)

		// Get values
		if totalStr, ok := keyToValue[totalKey]; ok {
			totalSeats, err := strconv.ParseInt(totalStr, 10, 32)
			if err == nil {
				result.Count += int(totalSeats)
			}
		}

		if availableStr, ok := keyToValue[availableKey]; ok {
			availableSeats, err := strconv.ParseInt(availableStr, 10, 32)
			if err == nil {
				result.Available += int(availableSeats)
			}
		}
	}

	result.Unavailable = result.Count - result.Available

	return &result, nil
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

		if setCacheErr := s.cache.Cache.Set(dropperCacheKey, raw); setCacheErr != nil {
			logger.FromCtx(ctx).Error("Cannot set cache keys", zap.Error(setCacheErr))
			return setCacheErr
		}

		return nil
	}

	cache, cacheErr := s.cache.Cache.Get(dropperCacheKey)

	var getKeysError error

	if cacheErr != nil {
		if !errors.Is(cacheErr, bigcache.ErrEntryNotFound) {
			logger.FromCtx(ctx).Error("Cannot get keys from cache", zap.Error(cacheErr))
		} else {
			getKeysError = getKeys()
		}
	} else {
		marshallErr := json.Unmarshal(cache, &keys)

		if marshallErr != nil {
			l.Error("Cannot unmashall cached keys")

			getKeysError = marshallErr
		}

		if len(keys) == 0 {
			l.Warn("got empty keys length from cache")
			getKeysError = getKeys()
		}
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
