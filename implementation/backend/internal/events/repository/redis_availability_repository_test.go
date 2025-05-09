package repository

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"tugas-akhir/backend/infrastructure/config"
	"tugas-akhir/backend/infrastructure/memcache"
	"tugas-akhir/backend/internal/events/entity"
	"tugas-akhir/backend/internal/events/repository/availability"
	"tugas-akhir/backend/internal/events/service/redis_availability_seeder"
	"tugas-akhir/backend/internal/seeder"
	test_containers "tugas-akhir/backend/test-containers"
)

// getIDsFromQuery is a helper to query one-column IDs and return a slice
func getIDsFromQuery(ctx context.Context, pool *pgxpool.Pool, query string, args ...interface{}) ([]int64, error) {
	rows, err := pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// TestRedisAvailabilityRepository_GetAvailability is an integration test for RedisAvailabilityRepository.
func TestRedisAvailabilityRepository_GetAvailability(t *testing.T) {
	for _, variant := range test_containers.RelationalDBVariants {
		t.Run(string(variant), func(t *testing.T) {
			ctx := context.Background()

			db := seeder.GetConnAndSchema(t, variant)
			seeder.SeedSchema(t, ctx, db)

			redisClient := test_containers.GetRedisCluster(t)

			cfg := &config.Config{
				PodName: "test-pod-1",
			}

			// Create the redis seeder
			redisSeeder := redis_availability_seeder.NewRedisAvailabilitySeeder(cfg, redisClient, db)

			// Run the seeder
			err := redisSeeder.Run(ctx)
			require.NoError(t, err)

			cache, cerr := memcache.NewMemcache()
			require.NoError(t, cerr)

			repo := availability.NewRedisAvailabilityRepository(redisClient, cache)

			// Query the necessary IDs from the database
			var (
				saleIDs         []int64
				packageIDs      []int64
				day1SeatedPkg1  int64
				day1SeatedPkg2  int64
				day1StandingPkg int64
				day1LawnPkg     int64
				day2SeatedPkg1  int64
				day2SeatedPkg2  int64
				day2StandingPkg int64
				day2LawnPkg     int64
			)

			// Get sale IDs (ordered by id, where first two correspond to day 1 and day 2)
			saleIDs, err = getIDsFromQuery(ctx, db.Pool, "SELECT id FROM ticket_sales ORDER BY id LIMIT 2")
			require.NoError(t, err)
			require.Len(t, saleIDs, 2, "expected two ticket sales")
			day1SaleID := saleIDs[0]
			day2SaleID := saleIDs[1]

			// Get package IDs for day 1
			packageIDs, err = getIDsFromQuery(ctx, db.Pool, `
		SELECT id FROM ticket_packages 
		WHERE ticket_sale_id = $1 
		ORDER BY id
	`, day1SaleID)
			require.NoError(t, err)
			require.GreaterOrEqual(t, len(packageIDs), 4, "expected at least 4 packages for day 1")
			day1SeatedPkg1 = packageIDs[0]  // Grandstand Row A-E
			day1SeatedPkg2 = packageIDs[1]  // Grandstand Row F-M
			day1StandingPkg = packageIDs[2] // Front Stage Pit
			day1LawnPkg = packageIDs[3]     // General Lawn Area

			// Get package IDs for day 2
			packageIDs, err = getIDsFromQuery(ctx, db.Pool, `
		SELECT id FROM ticket_packages 
		WHERE ticket_sale_id = $1 
		ORDER BY id
	`, day2SaleID)
			require.NoError(t, err)
			require.GreaterOrEqual(t, len(packageIDs), 4, "expected at least 4 packages for day 2")
			day2SeatedPkg1 = packageIDs[0]  // Grandstand Row A-E
			day2SeatedPkg2 = packageIDs[1]  // Grandstand Row F-M
			day2StandingPkg = packageIDs[2] // Front Stage Pit
			day2LawnPkg = packageIDs[3]     // General Lawn Area

			// Get area IDs for each package; we use a helper function for this.
			getAreaIDs := func(pkgID int64) []int64 {
				areas, err := getIDsFromQuery(ctx, db.Pool, `
			SELECT id FROM ticket_areas 
			WHERE ticket_package_id = $1 
			ORDER BY id
		`, pkgID)
				require.NoError(t, err)
				return areas
			}

			// Day 1 areas
			day1SeatedAreas1 := getAreaIDs(day1SeatedPkg1)   // Should be 1 area (e.g., 100 seats)
			day1SeatedAreas2 := getAreaIDs(day1SeatedPkg2)   // Should be 1 area (e.g., 250 seats)
			day1StandingAreas := getAreaIDs(day1StandingPkg) // Should be 1 area (e.g., 300 spots)
			day1LawnAreas := getAreaIDs(day1LawnPkg)         // Should be 2 areas (e.g., 1000 spots each)

			// Day 2 areas
			day2SeatedAreas1 := getAreaIDs(day2SeatedPkg1)   // Should be 1 area (e.g., 100 seats)
			day2SeatedAreas2 := getAreaIDs(day2SeatedPkg2)   // Should be 1 area (e.g., 250 seats)
			day2StandingAreas := getAreaIDs(day2StandingPkg) // Should be 1 area (e.g., 300 spots)
			day2LawnAreas := getAreaIDs(day2LawnPkg)         // Should be 2 areas (e.g., 1000 spots each)

			tests := []struct {
				name        string
				payload     entity.GetAvailabilityDto
				want        []entity.AreaAvailability
				expectError bool
			}{
				{
					name: "success - get availability for existing sale (Day 1)",
					payload: entity.GetAvailabilityDto{
						TicketSaleID: day1SaleID,
					},
					want: []entity.AreaAvailability{
						// Grandstand Row A-E (seated)
						{
							TicketSaleID:    day1SaleID,
							TicketPackageID: day1SeatedPkg1,
							TicketAreaID:    day1SeatedAreas1[0],
							TotalSeats:      100,
							AvailableSeats:  100,
						},
						// Grandstand Row F-M (seated)
						{
							TicketSaleID:    day1SaleID,
							TicketPackageID: day1SeatedPkg2,
							TicketAreaID:    day1SeatedAreas2[0],
							TotalSeats:      250,
							AvailableSeats:  250,
						},
						// Front Stage Pit (free-standing)
						{
							TicketSaleID:    day1SaleID,
							TicketPackageID: day1StandingPkg,
							TicketAreaID:    day1StandingAreas[0],
							TotalSeats:      300,
							AvailableSeats:  300,
						},
						// General Lawn Area (free-standing) - two areas expected
						{
							TicketSaleID:    day1SaleID,
							TicketPackageID: day1LawnPkg,
							TicketAreaID:    day1LawnAreas[0],
							TotalSeats:      1000,
							AvailableSeats:  1000,
						},
						{
							TicketSaleID:    day1SaleID,
							TicketPackageID: day1LawnPkg,
							TicketAreaID:    day1LawnAreas[1],
							TotalSeats:      1000,
							AvailableSeats:  1000,
						},
					},
					expectError: false,
				},
				{
					name: "success - get availability for existing sale (Day 1) - Cached",
					payload: entity.GetAvailabilityDto{
						TicketSaleID: day1SaleID,
					},
					want: []entity.AreaAvailability{
						// Grandstand Row A-E (seated)
						{
							TicketSaleID:    day1SaleID,
							TicketPackageID: day1SeatedPkg1,
							TicketAreaID:    day1SeatedAreas1[0],
							TotalSeats:      100,
							AvailableSeats:  100,
						},
						// Grandstand Row F-M (seated)
						{
							TicketSaleID:    day1SaleID,
							TicketPackageID: day1SeatedPkg2,
							TicketAreaID:    day1SeatedAreas2[0],
							TotalSeats:      250,
							AvailableSeats:  250,
						},
						// Front Stage Pit (free-standing)
						{
							TicketSaleID:    day1SaleID,
							TicketPackageID: day1StandingPkg,
							TicketAreaID:    day1StandingAreas[0],
							TotalSeats:      300,
							AvailableSeats:  300,
						},
						// General Lawn Area (free-standing) - two areas expected
						{
							TicketSaleID:    day1SaleID,
							TicketPackageID: day1LawnPkg,
							TicketAreaID:    day1LawnAreas[0],
							TotalSeats:      1000,
							AvailableSeats:  1000,
						},
						{
							TicketSaleID:    day1SaleID,
							TicketPackageID: day1LawnPkg,
							TicketAreaID:    day1LawnAreas[1],
							TotalSeats:      1000,
							AvailableSeats:  1000,
						},
					},
					expectError: false,
				},
				{
					name: "success - get availability for existing sale (Day 2)",
					payload: entity.GetAvailabilityDto{
						TicketSaleID: day2SaleID,
					},
					want: []entity.AreaAvailability{
						// Grandstand Row A-E (seated)
						{
							TicketSaleID:    day2SaleID,
							TicketPackageID: day2SeatedPkg1,
							TicketAreaID:    day2SeatedAreas1[0],
							TotalSeats:      100,
							AvailableSeats:  100,
						},
						// Grandstand Row F-M (seated)
						{
							TicketSaleID:    day2SaleID,
							TicketPackageID: day2SeatedPkg2,
							TicketAreaID:    day2SeatedAreas2[0],
							TotalSeats:      250,
							AvailableSeats:  250,
						},
						// Front Stage Pit (free-standing)
						{
							TicketSaleID:    day2SaleID,
							TicketPackageID: day2StandingPkg,
							TicketAreaID:    day2StandingAreas[0],
							TotalSeats:      300,
							AvailableSeats:  300,
						},
						// General Lawn Area (free-standing) - two areas expected
						{
							TicketSaleID:    day2SaleID,
							TicketPackageID: day2LawnPkg,
							TicketAreaID:    day2LawnAreas[0],
							TotalSeats:      1000,
							AvailableSeats:  1000,
						},
						{
							TicketSaleID:    day2SaleID,
							TicketPackageID: day2LawnPkg,
							TicketAreaID:    day2LawnAreas[1],
							TotalSeats:      1000,
							AvailableSeats:  1000,
						},
					},
					expectError: false,
				},
				{
					name: "success - get availability for existing sale (Day 2) - Cached",
					payload: entity.GetAvailabilityDto{
						TicketSaleID: day2SaleID,
					},
					want: []entity.AreaAvailability{
						// Grandstand Row A-E (seated)
						{
							TicketSaleID:    day2SaleID,
							TicketPackageID: day2SeatedPkg1,
							TicketAreaID:    day2SeatedAreas1[0],
							TotalSeats:      100,
							AvailableSeats:  100,
						},
						// Grandstand Row F-M (seated)
						{
							TicketSaleID:    day2SaleID,
							TicketPackageID: day2SeatedPkg2,
							TicketAreaID:    day2SeatedAreas2[0],
							TotalSeats:      250,
							AvailableSeats:  250,
						},
						// Front Stage Pit (free-standing)
						{
							TicketSaleID:    day2SaleID,
							TicketPackageID: day2StandingPkg,
							TicketAreaID:    day2StandingAreas[0],
							TotalSeats:      300,
							AvailableSeats:  300,
						},
						// General Lawn Area (free-standing) - two areas expected
						{
							TicketSaleID:    day2SaleID,
							TicketPackageID: day2LawnPkg,
							TicketAreaID:    day2LawnAreas[0],
							TotalSeats:      1000,
							AvailableSeats:  1000,
						},
						{
							TicketSaleID:    day2SaleID,
							TicketPackageID: day2LawnPkg,
							TicketAreaID:    day2LawnAreas[1],
							TotalSeats:      1000,
							AvailableSeats:  1000,
						},
					},
					expectError: false,
				},
				{
					name: "error - sale not found",
					payload: entity.GetAvailabilityDto{
						TicketSaleID: 999, // non-existent sale
					},
					want:        nil,
					expectError: true,
				},
			}

			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					ctx := context.Background()

					// --- EXECUTION PHASE ---
					got, err := repo.GetAvailability(ctx, tt.payload)

					// --- VALIDATION PHASE ---
					if tt.expectError {
						require.Error(t, err)
						if tt.name == "error - sale not found" {
							assert.Equal(t, entity.AreaAvailabilityNotFoundError, err)
						}
					} else {
						require.NoError(t, err)
						require.Equal(t, len(tt.want), len(got), "number of returned areas doesn't match")

						// For easier comparison, convert both slices to maps indexed by "packageID:areaID"
						gotMap := make(map[string]entity.AreaAvailability)
						for _, area := range got {
							key := fmt.Sprintf("%d:%d", area.TicketPackageID, area.TicketAreaID)
							gotMap[key] = area
						}

						for _, wantArea := range tt.want {
							key := fmt.Sprintf("%d:%d", wantArea.TicketPackageID, wantArea.TicketAreaID)
							found, ok := gotMap[key]
							require.True(t, ok, "expected area not found: %s", key)

							assert.Equal(t, wantArea.TotalSeats, found.TotalSeats, "total seats mismatch for area %s", key)
							assert.Equal(t, wantArea.AvailableSeats, found.AvailableSeats, "available seats mismatch for area %s", key)
						}
					}
				})
			}
		})
	}
}
