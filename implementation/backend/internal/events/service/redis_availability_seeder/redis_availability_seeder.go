package redis_availability_seeder

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"time"
	"tugas-akhir/backend/infrastructure/config"
	"tugas-akhir/backend/infrastructure/postgres"
	"tugas-akhir/backend/infrastructure/redis"
	entity2 "tugas-akhir/backend/internal/events/entity"
	availability2 "tugas-akhir/backend/internal/events/repository/availability"
	"tugas-akhir/backend/pkg/cursor_iterator"
	"tugas-akhir/backend/pkg/logger"
)

const seederRedisKey = "redis-availability-seeder-node"

type RedisAvailabilitySeeder struct {
	ctx        context.Context
	cancelFunc context.CancelFunc
	config     *config.Config
	redis      *redis.Redis
	db         *postgres.Postgres
}

func NewRedisAvailabilitySeeder(
	config *config.Config,
	redis *redis.Redis,
	db *postgres.Postgres,
) *RedisAvailabilitySeeder {
	ctx, cancel := context.WithCancel(context.Background())
	return &RedisAvailabilitySeeder{
		ctx:        ctx,
		cancelFunc: cancel,
		config:     config,
		redis:      redis,
		db:         db,
	}
}

func (s *RedisAvailabilitySeeder) iterAvailability() ([]entity2.AreaAvailability, *cursor_iterator.CursorIterator, error) {
	query := `
	SELECT 
	    tp.ticket_sale_id AS ticket_sale_id,
		tp.id AS ticket_package_id,
		ta.id AS ticket_area_id,
		COUNT(ts.id) AS total_seats,
		COUNT(CASE WHEN ts.status = 'available' THEN 1 END) AS available_seats
	FROM 
		ticket_packages tp
	INNER JOIN 
		ticket_areas ta ON ta.ticket_package_id = tp.id
	INNER JOIN 
		ticket_seats ts ON ts.ticket_area_id = ta.id
	GROUP BY 
		tp.id, ta.id, tp.ticket_sale_id
    `

	result := make([]entity2.AreaAvailability, 500)

	iter, err := cursor_iterator.NewCursorIterator(s.db.Pool, result, query)

	if err != nil {
		return nil, nil, err
	}

	return result, iter, err
}

func (s *RedisAvailabilitySeeder) tryAcquireSeeder() (bool, error) {
	result, err := s.redis.GetOrSetWithEx(s.ctx, seederRedisKey, s.config.PodName, 3*time.Hour)
	if err != nil {
		return false, err
	}
	return result == s.config.PodName, nil
}

func (s *RedisAvailabilitySeeder) refreshData(returnOnError bool) error {
	l := logger.FromCtx(s.ctx)

	shouldGo, err := s.tryAcquireSeeder()

	if err != nil {
		l.Sugar().Error(err)
		if returnOnError {
			return err
		}
		return nil
	}

	if !shouldGo {
		l.Info("skipping refresh redis availability because not instance with lock")
		if returnOnError {
			return fmt.Errorf("skipping refresh redis availability because not instance with lock")
		}
		return nil
	}

	l.Info("refreshing redis availability")

	// refresh data
	data, iter, err := s.iterAvailability()

	if err != nil {
		l.Sugar().Error(err)
		if returnOnError {
			return err
		}
		return nil
	}

	defer iter.Close(s.ctx)

	// toSet now maps a hash key (for a ticket sale) to its fields and values.
	// map[hashKey] -> map[field]value
	toSet := make(map[string]map[string]interface{})
	batchSize := 200 // Number of fields to process before sending a batch
	totalFieldsSet := 0

	sendBatch := func() error {
		pipe := s.redis.Client.Pipeline()
		for key, fields := range toSet {
			// Use HSet to set multiple fields in the hash for the given key.
			pipe.HSet(s.ctx, key, fields)
			totalFieldsSet += len(fields)
		}

		if _, err := pipe.Exec(s.ctx); err != nil {
			l.Sugar().Error("error executing pipeline for HSet", zap.Error(err))
			if returnOnError {
				return err
			}
		}

		// Clear the map for the next batch.
		toSet = make(map[string]map[string]interface{})
		return nil
	}

	for iter.Next(s.ctx) {
		availability := data[iter.ValueIndex()]
		key := availability2.CacheKey(availability.TicketSaleID)

		// Initialize the inner map if it doesn't exist for the current key.
		if _, ok := toSet[key]; !ok {
			toSet[key] = make(map[string]interface{})
		}

		// Add the total and available seats as fields to the hash.
		toSet[key][availability2.GetTotalSeatsField(availability)] = availability.TotalSeats
		toSet[key][availability2.GetAvailableSeatsField(availability)] = availability.AvailableSeats

		// Check if the number of fields in the current hash key batch is large enough to send.
		if len(toSet[key]) >= batchSize {
			if err = sendBatch(); returnOnError && err != nil {
				return err
			}
		}
	}

	if err := iter.Error(); err != nil {
		l.Sugar().Error("error during cursor iteration", zap.Error(err))
		if returnOnError {
			return err
		}
	}

	// Send any remaining data that didn't fill a full batch.
	if len(toSet) > 0 {
		if err = sendBatch(); returnOnError && err != nil {
			return err
		}
	}

	l.Info("completed seeder availability data with hashes", zap.Int("totalFieldsSet", totalFieldsSet))
	return nil
}

func (s *RedisAvailabilitySeeder) Stop() error {
	s.cancelFunc()
	return nil
}

func (s *RedisAvailabilitySeeder) RunSync(ctx context.Context) error {
	return s.refreshData(true)
}

func (s *RedisAvailabilitySeeder) Run(ctx context.Context) error {
	return s.refreshData(false)
}

func (s *RedisAvailabilitySeeder) ApplyAvailability(ctx context.Context, items []entity2.AreaAvailability) error {
	pipe := s.redis.Client.TxPipeline()

	for _, item := range items {
		key := availability2.CacheKey(item.TicketSaleID)
		field := availability2.GetAvailableSeatsField(item)
		// Use HIncrBy to decrement the value of a field within the hash.
		pipe.HIncrBy(ctx, key, field, -1)
	}

	_, err := pipe.Exec(ctx)
	return err
}

func (s *RedisAvailabilitySeeder) RevertAvailability(ctx context.Context, items []entity2.AreaAvailability) error {
	pipe := s.redis.Client.TxPipeline()

	for _, item := range items {
		key := availability2.CacheKey(item.TicketSaleID)
		field := availability2.GetAvailableSeatsField(item)
		// Use HIncrBy to increment the value of a field within the hash.
		pipe.HIncrBy(ctx, key, field, 1)
	}

	_, err := pipe.Exec(ctx)
	return err
}
