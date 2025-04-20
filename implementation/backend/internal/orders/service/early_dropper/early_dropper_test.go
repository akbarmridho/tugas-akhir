package early_dropper

import (
	"strconv"
	"testing"
	"tugas-akhir/backend/internal/bookings/repository/booked_seats"
	"tugas-akhir/backend/internal/bookings/service"
	"tugas-akhir/backend/internal/seeder"
	test_containers "tugas-akhir/backend/test-containers"

	errors2 "github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"tugas-akhir/backend/infrastructure/config"
	entity2 "tugas-akhir/backend/internal/events/entity"
	"tugas-akhir/backend/internal/orders/entity"
)

func TestEarlyDropper_Seed(t *testing.T) {
	ctx := t.Context()
	db := seeder.GetConnAndSchema(t, test_containers.RelationalDBVariant__Postgres)
	seeder.SeedSchema(t, ctx, db)

	redisInstance := test_containers.GetRedisCluster(t)

	dropper := NewPGPEarlyDropper(ctx, &config.Config{PodName: "default"}, redisInstance, booked_seats.NewPGBookedSeatRepository(db, service.NewSerialNumberGenerator()))
	require.NotNil(t, dropper)
	require.NoError(t, dropper.Run())
}

func TestEarlyDropper_TryAcquireLock_NumberedSeat_Success(t *testing.T) {
	ctx := t.Context()
	redisInstance := test_containers.GetRedisCluster(t)

	dropper := NewPGPEarlyDropper(ctx, &config.Config{PodName: "default"}, redisInstance, nil)
	require.NotNil(t, dropper)

	idempotencyKey := "test-idem-num-success"

	areaID := int64(1)
	seatIDs := []int64{100, 101}

	payload := entity.PlaceOrderDto{
		IdempotencyKey: &idempotencyKey,
		Items:          []entity.OrderItemDto{},
		TicketAreaID:   &areaID,
	}

	for _, seatID := range seatIDs {
		key := numberedSeatKey(areaID, seatID)
		err := redisInstance.Client.Set(ctx, key, string(entity2.SeatStatus__Available), 0).Err()
		require.NoError(t, err)

		payload.Items = append(payload.Items, entity.OrderItemDto{
			TicketSeatID: &seatID,
		})
	}

	// Action
	releaser, err := dropper.TryAcquireLock(ctx, payload)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, releaser)

	// Check Redis state
	for _, seatID := range seatIDs {
		key := numberedSeatKey(areaID, seatID)
		status, err := redisInstance.Client.Get(ctx, key).Result()
		require.NoError(t, err)
		assert.Equal(t, string(entity2.SeatStatus__OnHold), status)
	}

	// Test releaser (failure case - reverts state)
	err = releaser.OnFailed()
	assert.NoError(t, err)

	// Check Redis state after release
	for _, seatID := range seatIDs {
		key := numberedSeatKey(areaID, seatID)
		status, err := redisInstance.Client.Get(ctx, key).Result()
		require.NoError(t, err)
		assert.Equal(t, string(entity2.SeatStatus__Available), status)
	}

	// Try releasing again
	err = releaser.OnFailed()
	assert.ErrorIs(t, err, entity.LockAlreadyReleased)
	err = releaser.OnSuccess() // Also already released
	assert.ErrorIs(t, err, entity.LockAlreadyReleased)
}

func TestEarlyDropper_TryAcquireLock_NumberedSeat_Failure_NotAvailable(t *testing.T) {
	ctx := t.Context()

	redisInstance := test_containers.GetRedisCluster(t)

	dropper := NewPGPEarlyDropper(ctx, &config.Config{PodName: "default"}, redisInstance, nil)
	require.NotNil(t, dropper)

	idempotencyKey := "test-idem-num-fail"

	areaID := int64(1)
	seatIDs := []int64{200, 201}

	payload := entity.PlaceOrderDto{
		IdempotencyKey: &idempotencyKey,
		Items:          []entity.OrderItemDto{},
		TicketAreaID:   &areaID,
	}

	for _, seatID := range seatIDs {
		key := numberedSeatKey(areaID, seatID)
		err := redisInstance.Client.Set(ctx, key, string(entity2.SeatStatus__Sold), 0).Err()
		require.NoError(t, err)

		payload.Items = append(payload.Items, entity.OrderItemDto{
			TicketSeatID: &seatID,
		})
	}

	// Action
	releaser, err := dropper.TryAcquireLock(ctx, payload)

	// Assert
	require.Error(t, err)
	require.Nil(t, releaser)
	assert.True(t, errors2.Is(err, entity.DropperSeatNotAvailable), "Expected DropperSeatNotAvailable error")

	// Check Redis state
	for _, seatID := range seatIDs {
		key := numberedSeatKey(areaID, seatID)
		status, err := redisInstance.Client.Get(ctx, key).Result()
		require.NoError(t, err)
		assert.Equal(t, string(entity2.SeatStatus__Sold), status)
	}
}

func TestEarlyDropper_TryAcquireLock_FreeStanding_Success(t *testing.T) {
	ctx := t.Context()

	redisInstance := test_containers.GetRedisCluster(t)

	dropper := NewPGPEarlyDropper(ctx, &config.Config{PodName: "default"}, redisInstance, nil)
	require.NotNil(t, dropper)

	areaID := int64(1001)
	initialCount := 5
	requestCount := int64(2)
	key := freeStandingKey(areaID)
	idempotencyKey := "test-idem-free-success"

	err := redisInstance.Client.Set(ctx, key, strconv.Itoa(initialCount), 0).Err()
	require.NoError(t, err)

	payload := entity.PlaceOrderDto{
		IdempotencyKey: &idempotencyKey,
		Items: []entity.OrderItemDto{
			{TicketAreaID: areaID}, // Request 1
			{TicketAreaID: areaID}, // Request 2
		},
		TicketAreaID: &areaID,
	}

	// Action
	releaser, err := dropper.TryAcquireLock(ctx, payload)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, releaser)

	// Check Redis state
	countStr, err := redisInstance.Client.Get(ctx, key).Result()
	require.NoError(t, err)
	count, err := strconv.Atoi(countStr)
	require.NoError(t, err)
	assert.Equal(t, initialCount-int(requestCount), count)

	// Test releaser (success case - state persists for now in this implementation)
	err = releaser.OnSuccess()
	assert.NoError(t, err)

	// Check Redis state after success release (should be unchanged by this specific OnSuccess)
	countStr, err = redisInstance.Client.Get(ctx, key).Result()
	require.NoError(t, err)
	count, err = strconv.Atoi(countStr)
	require.NoError(t, err)
	assert.Equal(t, initialCount-int(requestCount), count)

	// Try releasing again
	err = releaser.OnFailed()
	assert.ErrorIs(t, err, entity.LockAlreadyReleased)
	err = releaser.OnSuccess() // Also already released
	assert.ErrorIs(t, err, entity.LockAlreadyReleased)
}

func TestEarlyDropper_TryAcquireLock_FreeStanding_Failure_Insufficient(t *testing.T) {
	ctx := t.Context()

	redisInstance := test_containers.GetRedisCluster(t)

	dropper := NewPGPEarlyDropper(ctx, &config.Config{PodName: "default"}, redisInstance, nil)
	require.NotNil(t, dropper)

	areaID := int64(1002)
	initialCount := 1
	key := freeStandingKey(areaID)
	idempotencyKey := "test-idem-free-fail"

	// Setup: Set initial availability
	err := redisInstance.Client.Set(ctx, key, strconv.Itoa(initialCount), 0).Err()
	require.NoError(t, err)

	payload := entity.PlaceOrderDto{
		IdempotencyKey: &idempotencyKey,
		Items: []entity.OrderItemDto{
			{TicketAreaID: areaID}, // Request 1
			{TicketAreaID: areaID}, // Request 2 (total 2, only 1 available)
		},
		TicketAreaID: &areaID,
	}

	// Action
	releaser, err := dropper.TryAcquireLock(ctx, payload)

	// Assert
	require.Error(t, err)
	require.Nil(t, releaser)
	assert.True(t, errors2.Is(err, entity.DropperSeatNotAvailable), "Expected DropperSeatNotAvailable error")

	// Check Redis state (should be unchanged)
	countStr, err := redisInstance.Client.Get(ctx, key).Result()
	require.NoError(t, err)
	count, err := strconv.Atoi(countStr)
	require.NoError(t, err)
	assert.Equal(t, initialCount, count)
}

func TestEarlyDropper_FinalizeLock_Success_Sold(t *testing.T) {
	ctx := t.Context()

	redisInstance := test_containers.GetRedisCluster(t)

	dropper := NewPGPEarlyDropper(ctx, &config.Config{PodName: "default"}, redisInstance, nil)
	require.NotNil(t, dropper)

	seatID1 := int64(301)
	seatID2 := int64(302)
	seatedAreaID := int64(1)
	areaID := int64(401)
	freeCountBeforeFinalize := 5 // Assume count was decremented by TryAcquireLock

	numberedKey1 := numberedSeatKey(seatedAreaID, seatID1)
	numberedKey2 := numberedSeatKey(seatedAreaID, seatID2)
	freeKey := freeStandingKey(areaID)

	// Setup: Simulate state after successful TryAcquireLock
	err := redisInstance.Client.Set(ctx, numberedKey1, string(entity2.SeatStatus__OnHold), 0).Err()
	require.NoError(t, err)
	err = redisInstance.Client.Set(ctx, numberedKey2, string(entity2.SeatStatus__OnHold), 0).Err()
	require.NoError(t, err)
	err = redisInstance.Client.Set(ctx, freeKey, strconv.Itoa(freeCountBeforeFinalize), 0).Err()
	require.NoError(t, err)

	// Items involved in the order that succeeded
	items := []entity.OrderItem{
		{TicketSeatID: seatID1, TicketSeat: &entity2.TicketSeat{ID: seatID1, TicketAreaID: seatedAreaID, TicketArea: &entity2.TicketArea{ID: seatedAreaID, Type: entity2.AreaType__NumberedSeating}}},
		{TicketSeatID: seatID2, TicketSeat: &entity2.TicketSeat{ID: seatID2, TicketAreaID: seatedAreaID, TicketArea: &entity2.TicketArea{ID: seatedAreaID, Type: entity2.AreaType__NumberedSeating}}},
		{TicketSeat: &entity2.TicketSeat{TicketAreaID: areaID, TicketArea: &entity2.TicketArea{ID: areaID, Type: entity2.AreaType__FreeStanding}}}, // Assume 1 free standing seat was part of the order
	}

	// Action: Finalize as Sold
	err = dropper.FinalizeLock(ctx, items, entity2.SeatStatus__Sold)

	// Assert
	require.NoError(t, err)

	// Check Redis state
	status1, err := redisInstance.Client.Get(ctx, numberedKey1).Result()
	require.NoError(t, err)
	assert.Equal(t, string(entity2.SeatStatus__Sold), status1) // Finalized to Sold

	status2, err := redisInstance.Client.Get(ctx, numberedKey2).Result()
	require.NoError(t, err)
	assert.Equal(t, string(entity2.SeatStatus__Sold), status2) // Finalized to Sold

	// Free standing count should remain unchanged when finalizing to Sold
	countStr, err := redisInstance.Client.Get(ctx, freeKey).Result()
	require.NoError(t, err)
	count, err := strconv.Atoi(countStr)
	require.NoError(t, err)
	assert.Equal(t, freeCountBeforeFinalize, count)
}

func TestEarlyDropper_FinalizeLock_Failure_Available(t *testing.T) {
	ctx := t.Context()

	redisInstance := test_containers.GetRedisCluster(t)

	dropper := NewPGPEarlyDropper(ctx, &config.Config{PodName: "default"}, redisInstance, nil)
	require.NotNil(t, dropper)
	seatID1 := int64(303)
	areaID := int64(403)
	seatedAreaID := int64(1)
	freeCountBeforeFinalize := 5 // Assume count was decremented by 1 during TryAcquireLock
	expectedFreeCountAfterRevert := freeCountBeforeFinalize + 1

	numberedKey1 := numberedSeatKey(seatedAreaID, seatID1)
	freeKey := freeStandingKey(areaID)

	// Setup: Simulate state after successful TryAcquireLock
	err := redisInstance.Client.Set(ctx, numberedKey1, string(entity2.SeatStatus__OnHold), 0).Err()
	require.NoError(t, err)
	err = redisInstance.Client.Set(ctx, freeKey, strconv.Itoa(freeCountBeforeFinalize), 0).Err()
	require.NoError(t, err)

	// Items involved in the order that failed/was cancelled
	items := []entity.OrderItem{
		{TicketSeatID: seatID1, TicketSeat: &entity2.TicketSeat{ID: seatID1, TicketAreaID: seatedAreaID, TicketArea: &entity2.TicketArea{ID: seatedAreaID, Type: entity2.AreaType__NumberedSeating}}},
		{TicketSeat: &entity2.TicketSeat{TicketAreaID: areaID, TicketArea: &entity2.TicketArea{ID: areaID, Type: entity2.AreaType__FreeStanding}}}, // Assume 1 free standing seat was part of the order
	}

	// Action: Finalize as Available (revert)
	err = dropper.FinalizeLock(ctx, items, entity2.SeatStatus__Available)

	// Assert
	require.NoError(t, err)

	// Check Redis state
	status1, err := redisInstance.Client.Get(ctx, numberedKey1).Result()
	require.NoError(t, err)
	assert.Equal(t, string(entity2.SeatStatus__Available), status1) // Reverted to Available

	// Free standing count should be incremented back
	countStr, err := redisInstance.Client.Get(ctx, freeKey).Result()
	require.NoError(t, err)
	count, err := strconv.Atoi(countStr)
	require.NoError(t, err)
	assert.Equal(t, expectedFreeCountAfterRevert, count)
}
