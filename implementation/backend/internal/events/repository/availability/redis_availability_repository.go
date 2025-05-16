package availability

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
	"tugas-akhir/backend/pkg/logger"
)

const AvailabilityPrefix = "redis-availability"

func cacheKey(pattern string) string {
	return fmt.Sprintf("%s:key:%s", AvailabilityPrefix, pattern)
}

func GetTotalSeatsKey(data entity.AreaAvailability) string {
	return fmt.Sprintf("%s:%d:%d:%d:total", AvailabilityPrefix, data.TicketSaleID, data.TicketPackageID, data.TicketAreaID)
}

func GetAvailableSeats(data entity.AreaAvailability) string {
	return fmt.Sprintf("%s:%d:%d:%d:available", AvailabilityPrefix, data.TicketSaleID, data.TicketPackageID, data.TicketAreaID)
}

type RedisAvailabilityRepository struct {
	redis *redis.Redis
	cache *memcache.Memcache
}

func NewRedisAvailabilityRepository(
	redis *redis.Redis,
	cache *memcache.Memcache,
) *RedisAvailabilityRepository {
	return &RedisAvailabilityRepository{
		redis: redis,
		cache: cache,
	}
}

func (r *RedisAvailabilityRepository) GetAvailability(ctx context.Context, payload entity.GetAvailabilityDto) ([]entity.AreaAvailability, error) {
	l := logger.FromCtx(ctx)

	pattern := fmt.Sprintf("%s:%d:*:*:*", AvailabilityPrefix, payload.TicketSaleID)

	keys := make([]string, 0)

	getKeys := func() error {
		var mu sync.Mutex

		// ForEachMaster will loop over every master node in the cluster.
		err := r.redis.Client.ForEachMaster(ctx, func(ctx context.Context, client *baseredis.Client) error {
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

		if setCacheErr := r.cache.Cache.Set(cacheKey(pattern), raw); setCacheErr != nil {
			logger.FromCtx(ctx).Error("Cannot set cache keys", zap.Error(setCacheErr))
			return setCacheErr
		}

		return nil
	}

	cache, cacheErr := r.cache.Cache.Get(cacheKey(pattern))

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

	pipe := r.redis.Client.Pipeline()

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
	result := make([]entity.AreaAvailability, 0, len(areaMap))

	for areaIdentifier, _ := range areaMap {
		parts := strings.Split(areaIdentifier, ":")
		if len(parts) != 3 {
			l.Sugar().Warnf("parts count not 3 for key %s", areaIdentifier)
			continue
		}

		saleID, _ := strconv.ParseInt(parts[0], 10, 64)
		packageID, _ := strconv.ParseInt(parts[1], 10, 64)
		areaID, _ := strconv.ParseInt(parts[2], 10, 64)

		area := entity.AreaAvailability{
			TicketSaleID:    saleID,
			TicketPackageID: packageID,
			TicketAreaID:    areaID,
		}

		// Construct keys for total and available
		totalKey := fmt.Sprintf("%s:%d:%d:%d:total", AvailabilityPrefix, saleID, packageID, areaID)
		availableKey := fmt.Sprintf("%s:%d:%d:%d:available", AvailabilityPrefix, saleID, packageID, areaID)

		// Get values
		if totalStr, ok := keyToValue[totalKey]; ok {
			totalSeats, err := strconv.ParseInt(totalStr, 10, 32)
			if err == nil {
				area.TotalSeats = int32(totalSeats)
			}
		}

		if availableStr, ok := keyToValue[availableKey]; ok {
			availableSeats, err := strconv.ParseInt(availableStr, 10, 32)
			if err == nil {
				area.AvailableSeats = int32(availableSeats)
			}
		}

		result = append(result, area)
	}

	return result, nil
}
