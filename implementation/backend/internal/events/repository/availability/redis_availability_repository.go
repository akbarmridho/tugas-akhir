package availability

import (
	"context"
	"errors"
	"fmt"
	baseredis "github.com/redis/go-redis/v9"
	"strconv"
	"strings"
	"tugas-akhir/backend/infrastructure/redis"
	"tugas-akhir/backend/internal/events/entity"
	"tugas-akhir/backend/pkg/logger"
)

const AvailabilityPrefix = "redis-availability"

func CacheKey(ticketSaleID int64) string {
	return fmt.Sprintf("%s:sale:%d", AvailabilityPrefix, ticketSaleID)
}

func GetTotalSeatsField(data entity.AreaAvailability) string {
	return fmt.Sprintf("%d:%d:total", data.TicketPackageID, data.TicketAreaID)
}

func GetAvailableSeatsField(data entity.AreaAvailability) string {
	return fmt.Sprintf("%d:%d:available", data.TicketPackageID, data.TicketAreaID)
}

type RedisAvailabilityRepository struct {
	redis *redis.Redis
}

func NewRedisAvailabilityRepository(
	redis *redis.Redis,
) *RedisAvailabilityRepository {
	return &RedisAvailabilityRepository{
		redis: redis,
	}
}

func (r *RedisAvailabilityRepository) GetAvailability(ctx context.Context, payload entity.GetAvailabilityDto) ([]entity.AreaAvailability, error) {
	l := logger.FromCtx(ctx)
	key := CacheKey(payload.TicketSaleID)

	// Fetch all fields and values from the hash in one operation.
	fields, err := r.redis.Client.HGetAll(ctx, key).Result()
	if err != nil {
		// This handles cases where the command fails for reasons other than the key not existing.
		if errors.Is(err, baseredis.Nil) {
			return nil, entity.AreaAvailabilityNotFoundError
		}
		l.Sugar().Errorf("failed to execute HGetAll for key %s: %v", key, err)
		return nil, err
	}

	if len(fields) == 0 {
		return nil, entity.AreaAvailabilityNotFoundError
	}

	// Use a map to aggregate total and available seats for each unique area.
	// The key is "packageID:areaID".
	areaMap := make(map[string]*entity.AreaAvailability)

	for field, value := range fields {
		parts := strings.Split(field, ":")
		if len(parts) != 3 {
			l.Sugar().Warnf("invalid field format in hash %s: %s", key, field)
			continue
		}

		packageID, pErr := strconv.ParseInt(parts[0], 10, 64)
		areaID, aErr := strconv.ParseInt(parts[1], 10, 64)
		fieldType := parts[2]

		if pErr != nil || aErr != nil {
			l.Sugar().Warnf("could not parse packageID or areaID from field: %s", field)
			continue
		}

		areaIdentifier := fmt.Sprintf("%d:%d", packageID, areaID)

		// If we haven't seen this area yet, create a new struct for it.
		if _, ok := areaMap[areaIdentifier]; !ok {
			areaMap[areaIdentifier] = &entity.AreaAvailability{
				TicketSaleID:    payload.TicketSaleID,
				TicketPackageID: packageID,
				TicketAreaID:    areaID,
			}
		}

		// Parse the seat count value.
		seatCount, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			l.Sugar().Warnf("could not parse seat count '%s' for field: %s", value, field)
			continue
		}

		// Assign the seat count to the correct field in the struct.
		switch fieldType {
		case "total":
			areaMap[areaIdentifier].TotalSeats = int32(seatCount)
		case "available":
			areaMap[areaIdentifier].AvailableSeats = int32(seatCount)
		}
	}

	// Convert the map of pointers to a slice of structs.
	result := make([]entity.AreaAvailability, 0, len(areaMap))
	for _, area := range areaMap {
		result = append(result, *area)
	}

	return result, nil
}
