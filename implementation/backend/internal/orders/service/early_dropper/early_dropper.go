package early_dropper

import (
	"context"
	"errors"
	"fmt"
	errors2 "github.com/pkg/errors"
	baseredis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"strconv"
	"time"
	"tugas-akhir/backend/infrastructure/config"
	"tugas-akhir/backend/infrastructure/redis"
	"tugas-akhir/backend/internal/bookings/repository/booked_seats"
	entity2 "tugas-akhir/backend/internal/events/entity"
	"tugas-akhir/backend/internal/orders/entity"
	"tugas-akhir/backend/pkg/logger"
)

const DropperRedisPrefix = "early-dropper:"
const referesherRedisKey = "refresher-node"

func numberedSeatKey(areaID int64, seatID int64) string {
	return fmt.Sprintf("%s{area-%d}status:numbered:%d", DropperRedisPrefix, areaID, seatID)
}

func freeStandingKey(areaID int64) string {
	return fmt.Sprintf("%s{area-%d}status:free-standing:%d", DropperRedisPrefix, areaID, areaID)
}

type EarlyDropper struct {
	ctx                  context.Context
	cancelFunc           context.CancelFunc
	config               *config.Config
	redis                *redis.Redis
	bookedSeatRepository booked_seats.BookedSeatRepository
}

func NewFCEarlyDropper(
	config *config.Config,
	redis *redis.Redis,
	bookedSeatRepository booked_seats.BookedSeatRepository,
) *EarlyDropper {
	ctx, cancel := context.WithCancel(context.Background())
	return &EarlyDropper{
		ctx:                  ctx,
		cancelFunc:           cancel,
		config:               config,
		redis:                redis,
		bookedSeatRepository: bookedSeatRepository,
	}
}

func (s *EarlyDropper) tryAcquireRefresher() (bool, error) {
	result, err := s.redis.GetOrSetWithEx(s.ctx, DropperRedisPrefix+referesherRedisKey, s.config.PodName, 15*time.Minute)

	if err != nil {
		return false, err
	}

	return result == s.config.PodName, nil
}

func (s *EarlyDropper) refreshData(returnOnError bool) error {
	l := logger.FromCtx(s.ctx)

	shouldGo, err := s.tryAcquireRefresher()

	if err != nil {
		l.Sugar().Error(err)
		if returnOnError {
			return err
		}
		return nil
	}

	if !shouldGo {
		l.Info("skipping refresh early dropper because not instance with lock")
		if returnOnError {
			return fmt.Errorf("skipping refresh early dropper because not instance with lock")
		}
		return nil
	}

	// refresh data
	data, iter, err := s.bookedSeatRepository.IterSeats(s.ctx)

	if err != nil {
		l.Sugar().Error(err)
		if returnOnError {
			return err
		}
		return nil
	}

	defer iter.Close(s.ctx)

	countSet := 0

	freeStandingAvailability := make(map[string]int)
	numberedBuffer := make([]entity2.TicketSeat, 0)
	numberedBufferBatchSize := 100

	sendNumberedBatch := func() error {
		values := make(map[string]string)

		for _, seat := range numberedBuffer {
			values[numberedSeatKey(seat.TicketArea.ID, seat.ID)] = string(seat.Status)
		}

		pipe := s.redis.Client.Pipeline()
		for k, v := range values {
			countSet++
			pipe.Set(s.ctx, k, v, 0)
		}

		numberedBuffer = make([]entity2.TicketSeat, 0)

		if _, err := pipe.Exec(s.ctx); err != nil {
			l.Sugar().Error(err)
			if returnOnError {
				return err
			}
		}

		return nil
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
				err = sendNumberedBatch()
				if returnOnError && err != nil {
					return err
				}
			}
		}
	}

	if err := iter.Error(); err != nil {
		l.Sugar().Error(err)
		if returnOnError {
			return err
		}
	}

	if len(numberedBuffer) > 0 {
		err = sendNumberedBatch()
		if returnOnError && err != nil {
			return err
		}
	}

	pipe := s.redis.Client.Pipeline()
	for k, v := range freeStandingAvailability {
		countSet++
		pipe.Set(s.ctx, k, v, 0)
		// add original count here
		pipe.Set(s.ctx, fmt.Sprintf("debug:%s", k), v, 0)
	}

	if _, err := pipe.Exec(s.ctx); err != nil {
		l.Sugar().Error(err)
		if returnOnError && err != nil {
			return err
		}
	}

	l.Info("completed refreshing early dropper data", zap.Int("countSet", countSet))
	return nil
}

func (s *EarlyDropper) Stop() error {
	s.cancelFunc()
	return nil
}

func (s *EarlyDropper) RunSync(ctx context.Context) error {
	return s.refreshData(true)
}

func (s *EarlyDropper) Run(ctx context.Context) error {
	return s.refreshData(false)
}

func (s *EarlyDropper) TryAcquireLock(ctx context.Context, payload entity.PlaceOrderDto) (*LockReleaser, error) {
	if payload.TicketAreaID == nil {
		return nil, fmt.Errorf("ticket area id must not be nil")
	}

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
		keysToWatch = append(keysToWatch, numberedSeatKey(*payload.TicketAreaID, seatID))
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
				numberedKeysToCheck = append(numberedKeysToCheck, numberedSeatKey(*payload.TicketAreaID, seatID))
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

				for k, result := range results {
					status, ok := result.(string)
					if !ok || status != string(entity2.SeatStatus__Available) {
						return errors2.WithMessagef(entity.DropperSeatNotAvailable, "seat %d is not available", numberedSeatIDs[k])
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
					key := numberedSeatKey(*payload.TicketAreaID, seatID)
					pipe.Set(ctx, key, string(entity2.SeatStatus__OnHold), 0)
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
					key := numberedSeatKey(*payload.TicketAreaID, seatID)
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
	return nil, errors2.WithStack(entity.CannotAcquireLock)
}

func (s *EarlyDropper) FinalizeLock(ctx context.Context, items []entity.OrderItem, status entity2.SeatStatus) error {
	if status == entity2.SeatStatus__OnHold {
		return fmt.Errorf("final seat status cannot be on hold")
	}

	// operation success
	if status == entity2.SeatStatus__Sold {
		// only update seat status for numbered seat
		pipe := s.redis.Client.TxPipeline()

		for _, item := range items {
			if item.TicketSeat.TicketArea.Type == entity2.AreaType__NumberedSeating {
				key := numberedSeatKey(item.TicketSeat.TicketAreaID, item.TicketSeatID)
				pipe.Set(ctx, key, string(entity2.SeatStatus__Sold), 0)
			}
		}

		_, err := pipe.Exec(ctx)
		if err != nil {
			return err
		}

		return nil
	}

	// operation fail
	if status == entity2.SeatStatus__Available {
		// only update seat status for numbered seat
		pipe := s.redis.Client.TxPipeline()

		for _, item := range items {
			if item.TicketSeat.TicketArea.Type == entity2.AreaType__NumberedSeating {
				key := numberedSeatKey(item.TicketSeat.TicketAreaID, item.TicketSeatID)
				pipe.Set(ctx, key, string(entity2.SeatStatus__Available), 0)
			} else {
				key := freeStandingKey(item.TicketSeat.TicketAreaID)
				pipe.IncrBy(ctx, key, int64(1))
			}
		}

		_, err := pipe.Exec(ctx)
		if err != nil {
			return err
		}

		return nil
	}

	return fmt.Errorf("code should be unreachable")
}
