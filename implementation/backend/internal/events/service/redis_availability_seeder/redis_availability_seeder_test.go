package redis_availability_seeder

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
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

			t.Run("Seeder loads availability data into Redis", func(t *testing.T) {
				// Create the seeder
				redisSeeder := NewRedisAvailabilitySeeder(cfg, redisClient, db)

				// Run the seeder
				err := redisSeeder.Run(ctx)
				require.NoError(t, err)

				// Verify data was loaded into Redis correctly
				// First, get all the availability data from database to validate against
				data, iter, err := redisSeeder.iterAvailability()
				require.NoError(t, err)
				defer iter.Close(ctx)

				// Extract all expected values
				expectedValues := make(map[string]int32)
				for iter.Next(ctx) {
					availabilityEntity := data[iter.ValueIndex()]
					expectedValues[availability.GetTotalSeatsKey(availabilityEntity)] = availabilityEntity.TotalSeats
					expectedValues[availability.GetAvailableSeats(availabilityEntity)] = availabilityEntity.AvailableSeats
				}
				require.NoError(t, iter.Error())

				// Verify each key in Redis
				for key, expectedValue := range expectedValues {
					actualValue, err := redisClient.Client.Get(ctx, key).Int()
					require.NoError(t, err)
					assert.Equal(t, int(expectedValue), actualValue, "Redis value for %s should match expected", key)
				}
			})

			t.Run("ApplyAvailability decrements available seats", func(t *testing.T) {
				// Create the seeder
				redisSeeder := NewRedisAvailabilitySeeder(cfg, redisClient, db)

				// Get a sample of data to test with
				data, iter, err := redisSeeder.iterAvailability()
				require.NoError(t, err)
				defer iter.Close(ctx)

				require.True(t, iter.Next(ctx))
				testItem := data[iter.ValueIndex()]

				// Get initial value
				key := availability.GetAvailableSeats(testItem)
				initialValue, err := redisClient.Client.Get(ctx, key).Int()
				require.NoError(t, err)

				// Apply availability (decrease count)
				err = redisSeeder.ApplyAvailability(ctx, []entity.AreaAvailability{testItem})
				require.NoError(t, err)

				// Verify the count decreased by 1
				newValue, err := redisClient.Client.Get(ctx, key).Int()
				require.NoError(t, err)
				assert.Equal(t, initialValue-1, newValue, "Available seats should decrease by 1")
			})

			t.Run("RevertAvailability increments available seats", func(t *testing.T) {
				// Create the seeder
				redisSeeder := NewRedisAvailabilitySeeder(cfg, redisClient, db)

				// Get a sample of data to test with
				data, iter, err := redisSeeder.iterAvailability()
				require.NoError(t, err)
				defer iter.Close(ctx)

				require.True(t, iter.Next(ctx))
				testItem := data[iter.ValueIndex()]

				// Get initial value
				key := availability.GetAvailableSeats(testItem)
				initialValue, err := redisClient.Client.Get(ctx, key).Int()
				require.NoError(t, err)

				// Revert availability (increase count)
				err = redisSeeder.RevertAvailability(ctx, []entity.AreaAvailability{testItem})
				require.NoError(t, err)

				// Verify the count increased by 1
				newValue, err := redisClient.Client.Get(ctx, key).Int()
				require.NoError(t, err)
				assert.Equal(t, initialValue+1, newValue, "Available seats should increase by 1")
			})

			t.Run("Lock mechanism allows only one instance to refresh data", func(t *testing.T) {
				// Create two seeders with different pod names
				seeder1 := NewRedisAvailabilitySeeder(cfg, redisClient, db)

				cfg2 := &config.Config{
					PodName: "test-pod-2",
				}
				seeder2 := NewRedisAvailabilitySeeder(cfg2, redisClient, db)

				// First seeder should acquire the lock
				acquired1, err := seeder1.tryAcquireSeeder()
				require.NoError(t, err)
				assert.True(t, acquired1, "First seeder should acquire the lock")

				// Second seeder should not acquire the lock
				acquired2, err := seeder2.tryAcquireSeeder()
				require.NoError(t, err)
				assert.False(t, acquired2, "Second seeder should not acquire the lock")

				// Wait for lock to expire (mock by deleting the key)
				err = redisClient.Client.Del(ctx, seederRedisKey).Err()
				require.NoError(t, err)

				// Now second seeder should acquire the lock
				acquired2, err = seeder2.tryAcquireSeeder()
				require.NoError(t, err)
				assert.True(t, acquired2, "Second seeder should acquire the lock after expiration")
			})

			t.Run("Seeder handles multiple areas in batch", func(t *testing.T) {
				// Create seeder
				redisSeeder := NewRedisAvailabilitySeeder(cfg, redisClient, db)

				// Get multiple items to test batch operations
				data, iter, err := redisSeeder.iterAvailability()
				require.NoError(t, err)
				defer iter.Close(ctx)

				var testItems []entity.AreaAvailability
				for iter.Next(ctx) && len(testItems) < 5 {
					testItems = append(testItems, data[iter.ValueIndex()])
				}
				require.NoError(t, iter.Error())
				require.GreaterOrEqual(t, len(testItems), 2, "Need at least 2 items for batch test")

				// Track initial values
				initialValues := make(map[string]int)
				for _, item := range testItems {
					key := availability.GetAvailableSeats(item)
					val, err := redisClient.Client.Get(ctx, key).Int()
					require.NoError(t, err)
					initialValues[key] = val
				}

				// Apply availability changes in batch
				err = redisSeeder.ApplyAvailability(ctx, testItems)
				require.NoError(t, err)

				// Verify all values decreased by 1
				for _, item := range testItems {
					key := availability.GetAvailableSeats(item)
					newVal, err := redisClient.Client.Get(ctx, key).Int()
					require.NoError(t, err)
					assert.Equal(t, initialValues[key]-1, newVal, "Available seats should decrease by 1 for key %s", key)
				}

				// Revert availability changes in batch
				err = redisSeeder.RevertAvailability(ctx, testItems)
				require.NoError(t, err)

				// Verify all values returned to initial values
				for _, item := range testItems {
					key := availability.GetAvailableSeats(item)
					newVal, err := redisClient.Client.Get(ctx, key).Int()
					require.NoError(t, err)
					assert.Equal(t, initialValues[key], newVal, "Available seats should return to initial value for key %s", key)
				}
			})
		})
	}
}
