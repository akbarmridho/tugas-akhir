package main

import (
	"context"
	baseredis "github.com/redis/go-redis/v9"
	_ "go.uber.org/automaxprocs"
	"go.uber.org/zap"
	"math"
	"os"
	"tugas-akhir/backend/infrastructure/config"
	"tugas-akhir/backend/infrastructure/postgres"
	"tugas-akhir/backend/infrastructure/redis"
	"tugas-akhir/backend/internal/bookings/repository/booked_seats"
	"tugas-akhir/backend/internal/bookings/service"
	"tugas-akhir/backend/internal/events/service/redis_availability_seeder"
	"tugas-akhir/backend/internal/orders/service/early_dropper"
	"tugas-akhir/backend/internal/seeder"
	"tugas-akhir/backend/pkg/logger"
)

func getScaledCategories(scale int) seeder.SeederPayload {
	return seeder.SeederPayload{
		DayCount: 0,
		SeatedCategories: []seeder.CategoryPayload{
			{Name: "Lower - Platinum East 1", Price: 3000000, AreaCount: 1, SeatPerArea: int(math.Floor(float64(2000 / scale)))},
			{Name: "Lower - Platinum East 2", Price: 3000000, AreaCount: 1, SeatPerArea: int(math.Floor(float64(2000 / scale)))},
			{Name: "Lower - Platinum West 1", Price: 3000000, AreaCount: 1, SeatPerArea: int(math.Floor(float64(2000 / scale)))},
			{Name: "Lower - Platinum West 2", Price: 3000000, AreaCount: 1, SeatPerArea: int(math.Floor(float64(2000 / scale)))},
			{Name: "Lower - Gold East 1", Price: 2500000, AreaCount: 1, SeatPerArea: int(math.Floor(float64(1750 / scale)))},
			{Name: "Lower - Gold East 2", Price: 2500000, AreaCount: 1, SeatPerArea: int(math.Floor(float64(1750 / scale)))},
			{Name: "Lower - Gold West 1", Price: 2500000, AreaCount: 1, SeatPerArea: int(math.Floor(float64(1750 / scale)))},
			{Name: "Lower - Gold West 2", Price: 2500000, AreaCount: 1, SeatPerArea: int(math.Floor(float64(1750 / scale)))},
			{Name: "Lower - Silver North", Price: 2000000, AreaCount: 5, SeatPerArea: int(math.Floor(float64(1000 / scale)))},
			{Name: "Lower - Silver South", Price: 2000000, AreaCount: 5, SeatPerArea: int(math.Floor(float64(1000 / scale)))},
			{Name: "Upper - Bronze West", Price: 1750000, AreaCount: 10, SeatPerArea: int(math.Floor(float64(1050 / scale)))},
			{Name: "Upper - Bronze East", Price: 1750000, AreaCount: 10, SeatPerArea: int(math.Floor(float64(1050 / scale)))},
			{Name: "Upper - Bronze North", Price: 1500000, AreaCount: 7, SeatPerArea: int(math.Floor(float64(1000 / scale)))},
			{Name: "Upper - Bronze South", Price: 1500000, AreaCount: 7, SeatPerArea: int(math.Floor(float64(1000 / scale)))},
		},
		FreeStandingCategories: []seeder.CategoryPayload{
			{Name: "VIP", Price: 4000000, AreaCount: 1, SeatPerArea: int(math.Floor(float64(4000 / scale)))},
			{Name: "Zone A", Price: 3250000, AreaCount: 1, SeatPerArea: int(math.Floor(float64(8000 / scale)))},
			{Name: "Zone B", Price: 2500000, AreaCount: 1, SeatPerArea: int(math.Floor(float64(8000 / scale)))},
		},
	}
}

// seedPayloadByScenario
// List of scenario
// xx-y
// xx variant: sf (scale full), s2 (scale by 2), s3 (scale by 3), ...
// y variant: 1, 2, 3 (day count)
// Festival/ free seating area can hold 20.000 person.
// Lower seat can hold 25.000 person.
// Upper seat can hold 35.000 person.
// In GBK, lower seat divided into:
// - Platinum East 1, Platinum East 2, Platinum West 1, Platinum West 2 @2000 seat -> 1 area
// - Gold East 1, Gold East 2, Gold West 1, Gold West 2 @1750 seat -> 1 area
// - Silver North, Silver South @5000 seat -> 5 area
// Upper seat can divided into:
// - Bronze North, Bronze South @7000 seat -> 7 area
// - Bronze West, Bronze East @10500 seat -> 10 area
// Festival can be divided into:
// - VIP Total 4000 seat.
// - Zone A Total 8000 seat.
// - Zone B Total 8000 seat.
func seedPayloadByScenario(scenario string) *seeder.SeederPayload {
	switch scenario {
	// Per day 80.000 ticket With a total of 4 day
	case "sf-4":
		payload := getScaledCategories(1)
		payload.DayCount = 4
		return &payload
	case "sf-2":
		payload := getScaledCategories(1)
		payload.DayCount = 2
		return &payload
	case "sf-1":
		payload := getScaledCategories(1)
		payload.DayCount = 1
		return &payload
	case "s2-4":
		payload := getScaledCategories(2)
		payload.DayCount = 4
		return &payload
	case "s2-2":
		payload := getScaledCategories(2)
		payload.DayCount = 2
		return &payload
	case "s2-1":
		payload := getScaledCategories(2)
		payload.DayCount = 1
		return &payload
	case "s5-4":
		payload := getScaledCategories(5)
		payload.DayCount = 4
		return &payload
	case "s5-2":
		payload := getScaledCategories(5)
		payload.DayCount = 2
		return &payload
	case "s5-1":
		payload := getScaledCategories(5)
		payload.DayCount = 1
		return &payload
	case "s10-4":
		payload := getScaledCategories(10)
		payload.DayCount = 4
		return &payload
	case "s10-2":
		payload := getScaledCategories(10)
		payload.DayCount = 2
		return &payload
	case "s10-1":
		payload := getScaledCategories(10)
		payload.DayCount = 1
		return &payload
	default:
		return nil
	}
}

func main() {
	l := logger.GetInfo().Sugar()

	c, err := config.NewConfig()

	if err != nil {
		l.Error(err)
		os.Exit(1)
	}

	db, err := postgres.NewPostgres(c)
	if err != nil {
		l.Error(err)
		os.Exit(1)
	}

	redisInstance, err := redis.NewRedis(c)
	if err != nil {
		l.Error(err)
		os.Exit(1)
	}

	ctx := context.Background()

	err = redisInstance.Client.ForEachMaster(ctx, func(ctx context.Context, client *baseredis.Client) error {
		return client.FlushDB(ctx).Err()
	})
	if err != nil {
		l.Error("failed flushing redis cluster", zap.Error(err))
		os.Exit(1)
	}

	schemaManager := seeder.NewSchemaManager(db)

	err = schemaManager.SchemaDown(ctx)
	if err != nil {
		l.Error("failed running schema down", zap.Error(err))
		os.Exit(1)
	}

	err = schemaManager.SchemaUp(ctx)
	if err != nil {
		l.Error("failed running schema up", zap.Error(err))
		os.Exit(1)
	}

	if c.DBVariant == config.DBVariant__Citusdata {
		err = schemaManager.CitusSetup(ctx)
		if err != nil {
			l.Error("failed running citus setup", zap.Error(err))
			os.Exit(1)
		}
	}

	caseSeeder := seeder.NewCaseSeeder(db)

	seedPayload := seedPayloadByScenario(c.TestScenario)

	if seedPayload == nil {
		l.Error("Invalid scneario")
		os.Exit(1)
	}

	err = caseSeeder.Seed(ctx, *seedPayload)
	if err != nil {
		l.Error("failed running seed", zap.Error(err))
		os.Exit(1)
	}

	availabilitySeeder := redis_availability_seeder.NewRedisAvailabilitySeeder(ctx, c, redisInstance, db)
	err = availabilitySeeder.RunSync()
	if err != nil {
		l.Error("failed running availability seeder", zap.Error(err))
		os.Exit(1)
	}

	if c.FlowControlVariant == config.FlowControlVariant__DropperAsync {
		earlyDropper := early_dropper.NewFCEarlyDropper(ctx, c, redisInstance, booked_seats.NewPGBookedSeatRepository(db, service.NewSerialNumberGenerator()))
		err = earlyDropper.RunSync()
		if err != nil {
			l.Error("failed running early dropper seeder", zap.Error(err))
			os.Exit(1)
		}
	}

	l.Info("seeder success")
	os.Exit(0)
}
