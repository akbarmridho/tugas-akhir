package early_dropper

import (
	"context"
	"errors"
	"fmt"
	errors2 "github.com/pkg/errors"
	baseredis "github.com/redis/go-redis/v9"
	"strconv"
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

func (s *EarlyDropper) TryAcquireLock(ctx context.Context, payload entity.PlaceOrderDto) (*LockReleaser, error) {
	l := logger.FromCtx(ctx)

	// Create maps to track operations we need to perform
	numberedSeatsToLock := make(map[int64]bool)
	freeStandingAreasToLock := make(map[int64]int)

	// Group seats by area for free standing and collect numbered seats
	for _, item := range payload.Items {
		if item.TicketSeatID != nil {
			numberedSeatsToLock[*item.TicketSeatID] = true
		} else {
			freeStandingAreasToLock[item.TicketAreaID]++
		}
	}

	// Collect all keys we need to watch
	var keysToWatch []string
	for seatID := range numberedSeatsToLock {
		keysToWatch = append(keysToWatch, numberedSeatKey(seatID))
	}
	for areaID := range freeStandingAreasToLock {
		keysToWatch = append(keysToWatch, freeStandingKey(areaID))
	}

	// Retry loop for optimistic locking
	var maxRetries = 5
	for i := 0; i < maxRetries; i++ {
		err := s.redis.Client.Watch(ctx, func(tx *baseredis.Tx) error {
			// Step 1: Check all numbered seats
			var numberedKeysToCheck []string
			var numberedSeatIDs []int64

			for seatID := range numberedSeatsToLock {
				numberedKeysToCheck = append(numberedKeysToCheck, numberedSeatKey(seatID))
				numberedSeatIDs = append(numberedSeatIDs, seatID)
			}

			var numberedResults *baseredis.SliceCmd
			if len(numberedKeysToCheck) > 0 {
				numberedResults = tx.MGet(ctx, numberedKeysToCheck...)
			}

			// Step 2: Check all free standing areas
			freeStandingChecks := make(map[int64]*baseredis.StringCmd)
			for areaID := range freeStandingAreasToLock {
				key := freeStandingKey(areaID)
				freeStandingChecks[areaID] = tx.Get(ctx, key)
			}

			// Execute all checks (implicit in the WATCH transaction)
			if len(numberedKeysToCheck) > 0 {
				results, err := numberedResults.Result()
				if err != nil && !errors.Is(err, baseredis.Nil) {
					return err
				}

				for _, result := range results {
					status, ok := result.(string)
					if !ok || status != string(entity2.SeatStatus__Available) {
						return errors2.WithMessagef(entity.DropperSeatNotAvailable, "seat %d is not available", numberedSeatIDs[i])
					}
				}
			}

			// Validate free standing areas
			for areaID, count := range freeStandingAreasToLock {
				availableStr, err := freeStandingChecks[areaID].Result()
				if errors.Is(err, baseredis.Nil) {
					return errors2.WithMessagef(entity.DropperInternalError, "free standing area %d not found", areaID)
				} else if err != nil {
					return err
				}

				available, err := strconv.Atoi(availableStr)
				if err != nil {
					return err
				}

				if available < count {
					return errors2.WithMessagef(entity.DropperSeatNotAvailable, "not enough seats available in area %d: requested %d, available %d", areaID, count, available)
				}
			}

			// If we get here, all checks passed - proceed with the transaction
			_, err := tx.TxPipelined(ctx, func(pipe baseredis.Pipeliner) error {
				// Lock numbered seats by changing their status to on-hold
				for seatID := range numberedSeatsToLock {
					key := numberedSeatKey(seatID)
					pipe.Set(ctx, key, string(entity2.SeatStatus__OnHold), 6*time.Hour) // bad but enough time for each test scenario to complete
				}

				// Decrement available counts for free standing areas
				for areaID, count := range freeStandingAreasToLock {
					key := freeStandingKey(areaID)
					pipe.DecrBy(ctx, key, int64(count))
				}
				return nil
			})
			return err
		}, keysToWatch...)

		if err == nil {
			// Success - create the releasers
			onSuccess := func() error {
				return nil
			}

			onFailure := func() error {
				pipe := s.redis.Client.TxPipeline()

				// Restore numbered seats to available
				for seatID := range numberedSeatsToLock {
					key := numberedSeatKey(seatID)
					pipe.Set(ctx, key, string(entity2.SeatStatus__Available), 0)
				}

				// Increment free standing area availability
				for areaID, count := range freeStandingAreasToLock {
					key := freeStandingKey(areaID)
					pipe.IncrBy(ctx, key, int64(count))
				}

				_, err := pipe.Exec(ctx)
				if err != nil {
					return err
				}

				return nil
			}

			return &LockReleaser{
				onSuccess: onSuccess,
				onFailed:  onFailure,
			}, nil
		}

		if errors.Is(err, baseredis.TxFailedErr) {
			// Conflict occurred, retry
			l.Sugar().Infof("Transaction conflict for %s, retrying (attempt %d/%d)", *payload.IdempotencyKey, i+1, maxRetries)
			continue
		}

		// Other error occurred
		return nil, err
	}

	// Max retries exceeded
	l.Sugar().Warnf("max retries exceeded for seat locking %s", *payload.IdempotencyKey)
	return nil, entity.CannotAcquireLock
}
