package main

import (
	"context"
	baseredis "github.com/redis/go-redis/v9"
	_ "go.uber.org/automaxprocs"
	"go.uber.org/zap"
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

func seedPayloadByScenario(scenario string) seeder.SeederPayload {
	switch scenario {
	default:
		return seeder.SeederPayload{
			DayCount: 2,
			SeatedCategories: []seeder.CategoryPayload{
				{Name: "Grandstand Row A-E", Price: 750000, AreaCount: 1, SeatPerArea: 100},
				{Name: "Grandstand Row F-M", Price: 500000, AreaCount: 1, SeatPerArea: 250},
			},
			FreeStandingCategories: []seeder.CategoryPayload{
				{Name: "Front Stage Pit", Price: 600000, AreaCount: 1, SeatPerArea: 300},
				{Name: "General Lawn Area", Price: 300000, AreaCount: 2, SeatPerArea: 1000},
			},
		}

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

	err = caseSeeder.Seed(ctx, seedPayloadByScenario(c.TestScenario))
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
