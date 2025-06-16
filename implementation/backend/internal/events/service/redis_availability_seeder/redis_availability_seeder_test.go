package redis_availability_seeder

import (
	"context"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"tugas-akhir/backend/infrastructure/config"
	"tugas-akhir/backend/internal/events/entity"
	"tugas-akhir/backend/internal/events/repository/availability"
	"tugas-akhir/backend/internal/seeder"
	test_containers "tugas-akhir/backend/test-containers"
)

func TestRedisAvailabilitySeeder(t *testing.T) {
	for _, variant := range test_containers.RelationalDBVariants {
		t.Run(string(variant), func(t *testing.T) {
			ctx := context.Background()

			db := seeder.GetConnAndSchema(t, variant)
			seeder.SeedSchema(t, ctx, db)

			redisClient := test_containers.GetRedisCluster(t)

			cfg := &config.Config{
				PodName: "test-pod-1",
			}

			// Clean up redis before each test run to ensure isolation
			redisClient.Client.FlushDB(ctx)

			t.Run("Seeder loads availability data into Redis using Hashes", func(t *testing.T) {
				// Create the seeder
				redisSeeder := NewRedisAvailabilitySeeder(cfg, redisClient, db)

				// Run the seeder
				err := redisSeeder.Run(ctx)
				require.NoError(t, err)

				// Verify data was loaded into Redis correctly
				data, iter, err := redisSeeder.iterAvailability()
				require.NoError(t, err)
				defer iter.Close(ctx)

				// Structure to hold expected values: map[hashKey]map[field]value
				expectedValues := make(map[string]map[string]string)
				for iter.Next(ctx) {
					item := data[iter.ValueIndex()]
					key := availability.CacheKey(item.TicketSaleID)

					if _, ok := expectedValues[key]; !ok {
						expectedValues[key] = make(map[string]string)
					}
					expectedValues[key][availability.GetTotalSeatsField(item)] = strconv.Itoa(int(item.TotalSeats))
					expectedValues[key][availability.GetAvailableSeatsField(item)] = strconv.Itoa(int(item.AvailableSeats))
				}
				require.NoError(t, iter.Error())

				// Verify each hash in Redis
				for key, fields := range expectedValues {
					for field, expectedValue := range fields {
						actualValue, err := redisClient.Client.HGet(ctx, key, field).Result()
						require.NoError(t, err, "failed to HGet key %s, field %s", key, field)
						assert.Equal(t, expectedValue, actualValue, "Redis value for key %s, field %s should match", key, field)
					}
				}
			})

			t.Run("ApplyAvailability decrements available seats in hash", func(t *testing.T) {
				redisSeeder := NewRedisAvailabilitySeeder(cfg, redisClient, db)
				// Ensure data is seeded first
				require.NoError(t, redisSeeder.Run(ctx))

				data, iter, err := redisSeeder.iterAvailability()
				require.NoError(t, err)
				defer iter.Close(ctx)

				require.True(t, iter.Next(ctx), "iterator should have at least one item")
				testItem := data[iter.ValueIndex()]

				key := availability.CacheKey(testItem.TicketSaleID)
				field := availability.GetAvailableSeatsField(testItem)

				initialValueStr, err := redisClient.Client.HGet(ctx, key, field).Result()
				require.NoError(t, err)
				initialValue, err := strconv.Atoi(initialValueStr)
				require.NoError(t, err)

				err = redisSeeder.ApplyAvailability(ctx, []entity.AreaAvailability{testItem})
				require.NoError(t, err)

				newValueStr, err := redisClient.Client.HGet(ctx, key, field).Result()
				require.NoError(t, err)
				newValue, err := strconv.Atoi(newValueStr)
				require.NoError(t, err)

				assert.Equal(t, initialValue-1, newValue, "Available seats should decrease by 1")
			})

			t.Run("RevertAvailability increments available seats in hash", func(t *testing.T) {
				redisSeeder := NewRedisAvailabilitySeeder(cfg, redisClient, db)
				// Ensure data is seeded first
				require.NoError(t, redisSeeder.Run(ctx))

				data, iter, err := redisSeeder.iterAvailability()
				require.NoError(t, err)
				defer iter.Close(ctx)

				require.True(t, iter.Next(ctx), "iterator should have at least one item")
				testItem := data[iter.ValueIndex()]

				key := availability.CacheKey(testItem.TicketSaleID)
				field := availability.GetAvailableSeatsField(testItem)

				initialValueStr, err := redisClient.Client.HGet(ctx, key, field).Result()
				require.NoError(t, err)
				initialValue, err := strconv.Atoi(initialValueStr)
				require.NoError(t, err)

				err = redisSeeder.RevertAvailability(ctx, []entity.AreaAvailability{testItem})
				require.NoError(t, err)

				newValueStr, err := redisClient.Client.HGet(ctx, key, field).Result()
				require.NoError(t, err)
				newValue, err := strconv.Atoi(newValueStr)
				require.NoError(t, err)

				assert.Equal(t, initialValue+1, newValue, "Available seats should increase by 1")
			})

			t.Run("Lock mechanism allows only one instance to refresh data", func(t *testing.T) {
				seeder1 := NewRedisAvailabilitySeeder(cfg, redisClient, db)
				cfg2 := &config.Config{PodName: "test-pod-2"}
				seeder2 := NewRedisAvailabilitySeeder(cfg2, redisClient, db)

				acquired1, err := seeder1.tryAcquireSeeder()
				require.NoError(t, err)
				assert.True(t, acquired1, "First seeder should acquire the lock")

				acquired2, err := seeder2.tryAcquireSeeder()
				require.NoError(t, err)
				assert.False(t, acquired2, "Second seeder should not acquire the lock")

				err = redisClient.Client.Del(ctx, seederRedisKey).Err()
				require.NoError(t, err)

				acquired2, err = seeder2.tryAcquireSeeder()
				require.NoError(t, err)
				assert.True(t, acquired2, "Second seeder should acquire the lock after expiration")
			})

			t.Run("Seeder handles multiple areas in batch with hashes", func(t *testing.T) {
				redisSeeder := NewRedisAvailabilitySeeder(cfg, redisClient, db)
				// Ensure data is seeded first
				require.NoError(t, redisSeeder.Run(ctx))

				data, iter, err := redisSeeder.iterAvailability()
				require.NoError(t, err)
				defer iter.Close(ctx)

				var testItems []entity.AreaAvailability
				for iter.Next(ctx) && len(testItems) < 5 {
					testItems = append(testItems, data[iter.ValueIndex()])
				}
				require.NoError(t, iter.Error())
				require.GreaterOrEqual(t, len(testItems), 2, "Need at least 2 items for batch test")

				type itemState struct {
					key   string
					field string
					value int
				}
				initialStates := make(map[string]itemState)

				for _, item := range testItems {
					key := availability.CacheKey(item.TicketSaleID)
					field := availability.GetAvailableSeatsField(item)
					valStr, err := redisClient.Client.HGet(ctx, key, field).Result()
					require.NoError(t, err)
					val, err := strconv.Atoi(valStr)
					require.NoError(t, err)
					initialStates[key+field] = itemState{key: key, field: field, value: val}
				}

				err = redisSeeder.ApplyAvailability(ctx, testItems)
				require.NoError(t, err)

				for _, state := range initialStates {
					newValStr, err := redisClient.Client.HGet(ctx, state.key, state.field).Result()
					require.NoError(t, err)
					newVal, err := strconv.Atoi(newValStr)
					require.NoError(t, err)
					assert.Equal(t, state.value-1, newVal, "Available seats should decrease by 1 for key %s, field %s", state.key, state.field)
				}

				err = redisSeeder.RevertAvailability(ctx, testItems)
				require.NoError(t, err)

				for _, state := range initialStates {
					newValStr, err := redisClient.Client.HGet(ctx, state.key, state.field).Result()
					require.NoError(t, err)
					newVal, err := strconv.Atoi(newValStr)
					require.NoError(t, err)
					assert.Equal(t, state.value, newVal, "Available seats should return to initial value for key %s, field %s", state.key, state.field)
				}
			})
		})
	}
}
