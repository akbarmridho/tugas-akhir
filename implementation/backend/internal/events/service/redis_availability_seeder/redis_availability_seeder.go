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

	totalSet := 0

	toSet := make(map[string]int32)
	batchSize := 200

	sendBatch := func() error {
		pipe := s.redis.Client.Pipeline()
		for k, v := range toSet {
			totalSet++
			pipe.Set(s.ctx, k, v, 0)
		}

		if _, err := pipe.Exec(s.ctx); err != nil {
			l.Sugar().Error(err)
			if returnOnError {
				return err
			}
		}

		toSet = make(map[string]int32)
		return nil
	}

	for iter.Next(s.ctx) {
		availability := data[iter.ValueIndex()]

		toSet[availability2.GetTotalSeatsKey(availability)] = availability.TotalSeats
		toSet[availability2.GetAvailableSeats(availability)] = availability.AvailableSeats

		if len(toSet) >= batchSize {
			err = sendBatch()
			if returnOnError && err != nil {
				return err
			}
		}
	}

	if err := iter.Error(); err != nil {
		l.Sugar().Error(err)
		if returnOnError {
			return err
		}
	}

	if len(toSet) > 0 {
		err = sendBatch()
		if returnOnError && err != nil {
			return err
		}
	}

	l.Info("completed seeder availability data", zap.Int("sendCount", totalSet))
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
		pipe.Decr(ctx, availability2.GetAvailableSeats(item))
	}

	_, err := pipe.Exec(ctx)

	return err
}

func (s *RedisAvailabilitySeeder) RevertAvailability(ctx context.Context, items []entity2.AreaAvailability) error {
	pipe := s.redis.Client.TxPipeline()

	for _, item := range items {
		pipe.Incr(ctx, availability2.GetAvailableSeats(item))
	}

	_, err := pipe.Exec(ctx)

	return err
}
