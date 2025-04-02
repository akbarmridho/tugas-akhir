package early_dropper

import (
	"context"
	"fmt"
	"time"
	"tugas-akhir/backend/infrastructure/config"
	"tugas-akhir/backend/infrastructure/redis"
	"tugas-akhir/backend/internal/bookings/repository/booking"
	entity2 "tugas-akhir/backend/internal/events/entity"
	"tugas-akhir/backend/internal/orders/entity"
	"tugas-akhir/backend/pkg/logger"
)

const redisPrefix = "early-dropper:"
const referesherRedisKey = "refresher-node"

func numberedSeatKey(seatID int64) string {
	return fmt.Sprintf("%sstatus:numbered:%d", redisPrefix, seatID)
}

func freeStandingKey(areaID int64) string {
	return fmt.Sprintf("%sstatus:free-standing:%d", redisPrefix, areaID)
}

type EarlyDropper struct {
	ctx               context.Context
	config            *config.Config
	redis             *redis.Redis
	bookingRepository booking.BookingRepository
}

func NewPGPEarlyDropper(
	ctx context.Context,
	config *config.Config,
	redis *redis.Redis,
	bookingRepository booking.BookingRepository,
) *EarlyDropper {
	return &EarlyDropper{
		ctx:               ctx,
		config:            config,
		redis:             redis,
		bookingRepository: bookingRepository,
	}
}

func (s *EarlyDropper) tryAcquireRefresher() (bool, error) {
	result, err := s.redis.GetOrSetWithEx(s.ctx, redisPrefix+referesherRedisKey, s.config.PodName, 3*time.Hour)

	if err == nil {
		return false, err
	}

	return result == s.config.PodName, nil
}

func (s *EarlyDropper) refreshData() {
	l := logger.FromCtx(s.ctx)

	shouldGo, err := s.tryAcquireRefresher()

	if err != nil {
		l.Sugar().Error(err)
		return
	}

	if !shouldGo {
		l.Info("skipping refresh early dropper because not instance with lock")
		return
	}

	// refresh data
	data, iter, err := s.bookingRepository.IterSeats(s.ctx)

	if err != nil {
		l.Sugar().Error(err)
		return
	}

	defer iter.Close(s.ctx)

	freeStandingAvailability := make(map[string]int)
	numberedBuffer := make([]entity2.TicketSeat, 0)
	numberedBufferBatchSize := 100

	sendNumberedBatch := func() {
		values := make(map[string]string)

		for _, seat := range numberedBuffer {
			values[numberedSeatKey(seat.ID)] = string(seat.Status)
		}

		if err := s.redis.Client.MSet(s.ctx, values).Err(); err != nil {
			l.Sugar().Error(err)
		}
	}

	for iter.Next(s.ctx) {
		seat := data[iter.ValueIndex()]

		if seat.TicketArea.Type == entity2.AreaType__FreeStanding {
			if seat.Status == entity2.SeatStatus__Available {
				key := freeStandingKey(seat.TicketAreaID)

				_, ok := freeStandingAvailability[key]

				if ok {
					freeStandingAvailability[key]++
				} else {
					freeStandingAvailability[key] = 1
				}
			}
		} else {
			// numbered
			numberedBuffer = append(numberedBuffer, seat)

			if len(numberedBuffer) >= numberedBufferBatchSize {
				sendNumberedBatch()
			}
		}
	}

	if err := iter.Error(); err != nil {
		l.Sugar().Error(err)
	}

	if len(numberedBuffer) > 0 {
		sendNumberedBatch()
	}

	if err := s.redis.Client.MSet(s.ctx, freeStandingAvailability).Err(); err != nil {
		l.Sugar().Error(err)
	}

	l.Info("completed refreshing early dropper data")
}

func (s *EarlyDropper) Stop() error {
	return nil
}

func (s *EarlyDropper) Run() error {
	s.refreshData()
	return nil
}

func (s *EarlyDropper) TryAcquireLock(ctx context.Context, payload entity.PlaceOrderDto) error {
	// todo try to acquire locks
	// todo think about the expirations
	// todo releaser on success and on failed
	return nil
}
