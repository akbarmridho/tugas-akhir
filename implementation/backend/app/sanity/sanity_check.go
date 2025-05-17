package sanity

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
	"sync"
	"time"
	"tugas-akhir/backend/infrastructure/memcache"
	"tugas-akhir/backend/infrastructure/postgres"
	"tugas-akhir/backend/infrastructure/redis"
	"tugas-akhir/backend/pkg/logger"
)

type SanityCheck struct {
	pgCheck    *PGCheck
	redisCheck *RedisCheck
	quitChan   *chan struct{}

	// metrics
	dbAvailability    *prometheus.GaugeVec
	redisAvailability *prometheus.GaugeVec
	redisDropper      *prometheus.GaugeVec
	doubleOrder       prometheus.Gauge
}

func (s *SanityCheck) Run(ctx context.Context) {
	l := logger.FromCtx(ctx)

	l.Info("sanity check initialized")

	ticker := time.NewTicker(3 * time.Minute)

	quit := make(chan struct{})

	go func() {
		for {
			select {
			case <-ticker.C:
				s.Collect(ctx)
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

	s.quitChan = &quit

	// run immediately after started
	s.Collect(ctx)
}

func (s *SanityCheck) Stop() {
	if s.quitChan != nil {
		close(*s.quitChan)
	}
}

func (s *SanityCheck) push(gauge *prometheus.GaugeVec, check *AvailabilityCheck) {
	gauge.With(prometheus.Labels{
		"status": "total",
	}).Set(float64(check.Count))

	gauge.With(prometheus.Labels{
		"status": "available",
	}).Set(float64(check.Available))

	gauge.With(prometheus.Labels{
		"status": "unavailable",
	}).Set(float64(check.Unavailable))
}

func (s *SanityCheck) Collect(ctx context.Context) {
	l := logger.FromCtx(ctx)

	wg := sync.WaitGroup{}

	wg.Add(4)

	go func() {
		result, err := s.pgCheck.GetAvailability(ctx)

		defer wg.Done()

		if err != nil {
			l.Sugar().Error("cannot check availability from pg", zap.Error(err))
			return
		}

		s.push(s.dbAvailability, result)
	}()

	go func() {
		result, err := s.redisCheck.GetAvailability(ctx)

		defer wg.Done()

		if err != nil {
			l.Sugar().Error("cannot check availability from redis", zap.Error(err))
			return
		}

		s.push(s.redisAvailability, result)
	}()

	go func() {
		result, err := s.redisCheck.GetDropperAvailability(ctx)

		defer wg.Done()

		if err != nil {
			l.Sugar().Error("cannot check redis dropper", zap.Error(err))
			return
		}

		s.push(s.redisDropper, result)
	}()

	go func() {
		result, err := s.pgCheck.CheckDoubleOrder(ctx)

		defer wg.Done()

		if err != nil {
			l.Sugar().Error("cannot check double order", zap.Error(err))
			return
		}

		s.doubleOrder.Set(float64(result.Total))
	}()

	wg.Wait()
	l.Info("finished collecting sanity check")
}

func NewSanityCheck(
	db *postgres.Postgres,
	redis *redis.Redis,
) (*SanityCheck, error) {
	cache, err := memcache.NewMemcache()

	if err != nil {
		return nil, err
	}

	pgCheck := PGCheck{
		db: db,
	}

	redisCheck := RedisCheck{
		redis: redis,
		cache: cache,
	}

	dbAvailability := promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ticket_app_db_availability",
		Help: "Database Seat Availability",
	}, []string{"status"})

	redisAvailability := promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ticket_app_redis_availability",
		Help: "Redis Seat Availability",
	}, []string{"status"})

	redisDropper := promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ticket_app_redis_dropper",
		Help: "Redis Dropper Availability",
	}, []string{"status"})

	doubleOrder := promauto.NewGauge(prometheus.GaugeOpts{
		Name: "ticket_app_double_order",
		Help: "Seat double order check",
	})

	return &SanityCheck{
		pgCheck:           &pgCheck,
		redisCheck:        &redisCheck,
		dbAvailability:    dbAvailability,
		redisAvailability: redisAvailability,
		redisDropper:      redisDropper,
		doubleOrder:       doubleOrder,
	}, nil
}
