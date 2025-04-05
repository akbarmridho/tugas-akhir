package availability

import (
	"context"
	"fmt"
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

	// Scan for all matching keys
	var cursor uint64
	var keys []string

	for {
		var batch []string

		scanCmd := r.redis.Client.Scan(ctx, cursor, pattern, 100)

		batch, cursor = scanCmd.Val()

		if scanCmd.Err() != nil {
			return nil, scanCmd.Err()
		}

		keys = append(keys, batch...)

		if cursor == 0 {
			break
		}
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

	valuesCmd := r.redis.Client.MGet(ctx, keys...)
	if valuesCmd.Err() != nil {
		return nil, valuesCmd.Err()
	}

	values := valuesCmd.Val()
	keyToValue := make(map[string]string)

	for i, key := range keys {
		if values[i] != nil {
			if strValue, ok := values[i].(string); ok {
				keyToValue[key] = strValue
			}
		}
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
