package redis_availability_seeder

import (
	"context"
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
	ctx    context.Context
	config *config.Config
	redis  *redis.Redis
	db     *postgres.Postgres
}

func NewRedisAvailabilitySeeder(
	ctx context.Context,
	config *config.Config,
	redis *redis.Redis,
	db *postgres.Postgres,
) *RedisAvailabilitySeeder {
	return &RedisAvailabilitySeeder{
		ctx:    ctx,
		config: config,
		redis:  redis,
		db:     db,
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

func (s *RedisAvailabilitySeeder) refreshData() {
	l := logger.FromCtx(s.ctx)

	shouldGo, err := s.tryAcquireSeeder()

	if err != nil {
		l.Sugar().Error(err)
		return
	}

	if !shouldGo {
		l.Info("skipping refresh redis availability because not instance with lock")
		return
	}

	// refresh data
	data, iter, err := s.iterAvailability()

	if err != nil {
		l.Sugar().Error(err)
		return
	}

	defer iter.Close(s.ctx)

	toSet := make(map[string]int32)
	batchSize := 200

	sendBatch := func() {
		pipe := s.redis.Client.Pipeline()
		for k, v := range toSet {
			pipe.Set(s.ctx, k, v, 0)
		}

		if _, err := pipe.Exec(s.ctx); err != nil {
			l.Sugar().Error(err)
		}

		toSet = make(map[string]int32)
	}

	for iter.Next(s.ctx) {
		availability := data[iter.ValueIndex()]

		toSet[availability2.GetTotalSeatsKey(availability)] = availability.TotalSeats
		toSet[availability2.GetAvailableSeats(availability)] = availability.AvailableSeats

		if len(toSet) >= batchSize {
			sendBatch()
		}
	}

	if err := iter.Error(); err != nil {
		l.Sugar().Error(err)
	}

	if len(toSet) > 0 {
		sendBatch()
	}

	l.Info("completed seeder availability data")
}

func (s *RedisAvailabilitySeeder) Stop() error {
	return nil
}

func (s *RedisAvailabilitySeeder) Run() error {
	s.refreshData()
	return nil
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
