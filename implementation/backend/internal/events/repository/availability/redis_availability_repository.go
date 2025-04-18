package availability

import (
	"context"
	"fmt"
	baseredis "github.com/redis/go-redis/v9"
	"strconv"
	"strings"
	"tugas-akhir/backend/infrastructure/redis"
	"tugas-akhir/backend/internal/events/entity"
	"tugas-akhir/backend/pkg/logger"
)

const prefix = "redis-availability"

func GetTotalSeatsKey(data entity.AreaAvailability) string {
	return fmt.Sprintf("%s:%d:%d:%d:total", prefix, data.TicketSaleID, data.TicketPackageID, data.TicketAreaID)
}

func GetAvailableSeats(data entity.AreaAvailability) string {
	return fmt.Sprintf("%s:%d:%d:%d:available", prefix, data.TicketSaleID, data.TicketPackageID, data.TicketAreaID)
}

type RedisAvailabilityRepository struct {
	redis *redis.Redis
}

func NewRedisAvailabilityRepository(redis *redis.Redis) *RedisAvailabilityRepository {
	return &RedisAvailabilityRepository{
		redis: redis,
	}
}

func (r *RedisAvailabilityRepository) GetAvailability(ctx context.Context, payload entity.GetAvailabilityDto) ([]entity.AreaAvailability, error) {
	l := logger.FromCtx(ctx)

	pattern := fmt.Sprintf("%s:%d:*:*:*", prefix, payload.TicketSaleID)

	var keys []string

	// ForEachMaster will loop over every master node in the cluster.
	err := r.redis.Client.ForEachMaster(ctx, func(ctx context.Context, client *baseredis.Client) error {
		var cursor uint64 = 0
		for {
			// Scan on the current node.
			ckeys, nextCursor, err := client.Scan(ctx, cursor, pattern, 100).Result()
			if err != nil {
				return err
			}

			keys = append(keys, ckeys...)

			if nextCursor == 0 {
				break
			}
			cursor = nextCursor
		}
		return nil
	})

	if err != nil {
		return nil, err
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
	for _, key := range keys {
		val, err := r.redis.Client.Get(ctx, key).Result()
		if err != nil {
			// Skip if key doesn't exist or other error
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
		totalKey := fmt.Sprintf("%s:%d:%d:%d:total", prefix, saleID, packageID, areaID)
		availableKey := fmt.Sprintf("%s:%d:%d:%d:available", prefix, saleID, packageID, areaID)

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
