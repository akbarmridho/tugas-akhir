package seeder

import (
	"context"
	"fmt"
	"time"
	"tugas-akhir/backend/infrastructure/postgres"
	"tugas-akhir/backend/pkg/logger"

	"github.com/jackc/pgx/v5"
)

type CategoryPayload struct {
	Name        string
	Price       int
	AreaCount   int
	SeatPerArea int
}

type SeederPayload struct {
	DayCount               int
	SeatedCategories       []CategoryPayload
	FreeStandingCategories []CategoryPayload
}

type CaseSeeder struct {
	db *postgres.Postgres
}

func NewCaseSeeder(db *postgres.Postgres) *CaseSeeder {
	return &CaseSeeder{
		db: db,
	}
}

const seatBatchSize = 500

func (c *CaseSeeder) Seed(ctx context.Context, payload SeederPayload) (err error) {
	l := logger.FromCtx(ctx)

	// Start transaction
	tx, err := c.db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Ensure rollback on error or panic, commit on success
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback(ctx)
			panic(r)
		} else if err != nil {
			if rbErr := tx.Rollback(ctx); rbErr != nil {
				err = fmt.Errorf("seed error: %w; rollback error: %w", err, rbErr)
				l.Sugar().Error(err)
			} else {
				l.Sugar().Error(err)
			}
		} else {
			err = tx.Commit(ctx)
			if err != nil {
				err = fmt.Errorf("failed to commit transaction: %w", err)
				l.Sugar().Error(err)
			}
		}
	}()

	// --- 1. Create Event ---
	var eventID int64
	err = tx.QueryRow(ctx, `
		INSERT INTO events (name, location, description)
		VALUES ($1, $2, $3)
		RETURNING id
	`, "Music World Tour", "Jakarta International Stadium", "").Scan(&eventID)
	if err != nil {
		return fmt.Errorf("failed to insert event: %w", err)
	}
	l.Sugar().Infof("Created event with ID: %d", eventID)

	// --- 2. Create Ticket Categories ---
	categoryIDs := make(map[string]int64)
	allCategories := append([]CategoryPayload{}, payload.SeatedCategories...)
	allCategories = append(allCategories, payload.FreeStandingCategories...)

	for _, catPayload := range allCategories {
		var categoryID int64
		err = tx.QueryRow(ctx, `
			INSERT INTO ticket_categories (name, event_id)
			VALUES ($1, $2)
			RETURNING id
		`, catPayload.Name, eventID).Scan(&categoryID)
		if err != nil {
			return fmt.Errorf("failed to insert ticket category '%s': %w", catPayload.Name, err)
		}
		categoryIDs[catPayload.Name] = categoryID
		l.Sugar().Infof("Created category '%s' with ID: %d", catPayload.Name, categoryID)
	}

	// --- 3. Create Ticket Sales ---
	var ticketSaleID int64
	saleBeginAt := time.Now()
	saleEndAt := saleBeginAt.AddDate(0, 0, 7)

	for i := 1; i <= payload.DayCount; i++ {
		saleName := fmt.Sprintf("Day %d", i)

		err = tx.QueryRow(ctx, `
		INSERT INTO ticket_sales (name, sale_begin_at, sale_end_at, event_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, saleName, saleBeginAt, saleEndAt, eventID).Scan(&ticketSaleID)
		if err != nil {
			return fmt.Errorf("failed to insert ticket sale '%s': %w", saleName, err)
		}
		l.Sugar().Infof("Created ticket sale '%s' with ID: %d (Ends: %s)", saleName, ticketSaleID, saleEndAt.Format(time.RFC3339))

		// --- Prepare for batch seat insertion ---
		seatData := make([][]interface{}, 0, seatBatchSize)
		totalSeatsCreated := 0

		// --- 4. Create Ticket Packages, Areas, and Seats (Batched) ---

		// --- Handle Seated Categories ---
		fmt.Println("Processing Seated Categories...")
		for _, catPayload := range payload.SeatedCategories {
			categoryID, ok := categoryIDs[catPayload.Name]
			if !ok {
				return fmt.Errorf("internal error: category ID for '%s' not found", catPayload.Name)
			}

			// Create Ticket Package
			var ticketPackageID int64
			err = tx.QueryRow(ctx, `
			INSERT INTO ticket_packages (price, ticket_category_id, ticket_sale_id)
			VALUES ($1, $2, $3)
			RETURNING id
		`, catPayload.Price, categoryID, ticketSaleID).Scan(&ticketPackageID)
			if err != nil {
				return fmt.Errorf("failed to insert ticket package for seated category '%s': %w", catPayload.Name, err)
			}
			l.Sugar().Infof("Created package for seated category '%s' with ID: %d", catPayload.Name, ticketPackageID)

			// Create Ticket Areas and Seats for this package
			for areaIdx := 0; areaIdx < catPayload.AreaCount; areaIdx++ {
				// Create Ticket Area
				var ticketAreaID int64
				err = tx.QueryRow(ctx, `
				INSERT INTO ticket_areas (type, ticket_package_id)
				VALUES ($1, $2)
				RETURNING id
			`, "numbered-seating", ticketPackageID).Scan(&ticketAreaID)
				if err != nil {
					return fmt.Errorf("failed to insert numbered-seating area %d for package %d: %w", areaIdx+1, ticketPackageID, err)
				}
				l.Sugar().Infof("Created numbered-seating area with ID: %d", ticketAreaID)

				// Prepare seats for batch insert
				for seatIdx := 0; seatIdx < catPayload.SeatPerArea; seatIdx++ {
					// Example: Seat numbering like "R1-A1", "R1-A2" .. "R5-B100"
					seatNumber := fmt.Sprintf("S-A%d-%d", areaIdx+1, seatIdx+1)
					seatData = append(seatData, []interface{}{
						seatNumber,
						"available", // default status
						ticketAreaID,
					})

					// If batch is full, insert it
					if len(seatData) >= seatBatchSize {
						count, insertErr := insertSeatBatch(ctx, tx, seatData)
						if insertErr != nil {
							return fmt.Errorf("failed to batch insert seats (seated): %w", insertErr)
						}
						totalSeatsCreated += int(count)
						l.Sugar().Infof("Inserted batch of %d seats (Total: %d)", count, totalSeatsCreated)
						seatData = seatData[:0] // Clear the batch slice
					}
				} // End seat loop
			} // End area loop
		} // End seated category loop

		// --- Handle Free Standing Categories ---
		fmt.Println("Processing Free Standing Categories...")
		for _, catPayload := range payload.FreeStandingCategories {
			categoryID, ok := categoryIDs[catPayload.Name]
			if !ok {
				return fmt.Errorf("internal error: category ID for '%s' not found", catPayload.Name)
			}

			// Create Ticket Package
			var ticketPackageID int64
			err = tx.QueryRow(ctx, `
			INSERT INTO ticket_packages (price, ticket_category_id, ticket_sale_id)
			VALUES ($1, $2, $3)
			RETURNING id
		`, catPayload.Price, categoryID, ticketSaleID).Scan(&ticketPackageID)
			if err != nil {
				return fmt.Errorf("failed to insert ticket package for free-standing category '%s': %w", catPayload.Name, err)
			}
			l.Sugar().Infof("Created package for free-standing category '%s' with ID: %d", catPayload.Name, ticketPackageID)

			// Create Ticket Areas and Seats for this package
			for areaIdx := 0; areaIdx < catPayload.AreaCount; areaIdx++ {
				// Create Ticket Area
				var ticketAreaID int64
				err = tx.QueryRow(ctx, `
				INSERT INTO ticket_areas (type, ticket_package_id)
				VALUES ($1, $2)
				RETURNING id
			`, "free-standing", ticketPackageID).Scan(&ticketAreaID)
				if err != nil {
					return fmt.Errorf("failed to insert free-standing area %d for package %d: %w", areaIdx+1, ticketPackageID, err)
				}
				l.Sugar().Infof("Created free-standing area with ID: %d", ticketAreaID)

				// Prepare seats (conceptual spots) for batch insert
				for seatIdx := 0; seatIdx < catPayload.SeatPerArea; seatIdx++ {
					// Example: Seat numbering like "FS-A1-1", "FS-A1-2" ... "FS-A3-200"
					seatNumber := fmt.Sprintf("FS-A%d-%d", areaIdx+1, seatIdx+1) // Use a distinct naming scheme
					seatData = append(seatData, []interface{}{
						seatNumber,
						"available", // default status
						ticketAreaID,
					})

					// If batch is full, insert it
					if len(seatData) >= seatBatchSize {
						count, insertErr := insertSeatBatch(ctx, tx, seatData)
						if insertErr != nil {
							return fmt.Errorf("failed to batch insert seats (free-standing): %w", insertErr)
						}
						totalSeatsCreated += int(count)
						l.Sugar().Infof("Inserted batch of %d spots (Total: %d)", count, totalSeatsCreated)
						seatData = seatData[:0] // Clear the batch slice
					}
				} // End seat loop
			} // End area loop
		} // End free-standing category loop

		// --- 5. Insert any remaining seats in the last batch ---
		if len(seatData) > 0 {
			count, insertErr := insertSeatBatch(ctx, tx, seatData)
			if insertErr != nil {
				return fmt.Errorf("failed to batch insert remaining seats/spots: %w", insertErr)
			}
			totalSeatsCreated += int(count)
			l.Sugar().Infof("Inserted final batch of %d seats/spots (Grand Total: %d)", count, totalSeatsCreated)
		}

	}

	l.Info("Seeding completed successfully.")
	// Transaction commit/rollback is handled by the defer function

	return nil // err is nil if commit is successful
}

// Helper function for batch inserting seats using CopyFrom
func insertSeatBatch(ctx context.Context, tx pgx.Tx, seatData [][]interface{}) (int64, error) {
	if len(seatData) == 0 {
		return 0, nil
	}

	copyCount, err := tx.CopyFrom(
		ctx,
		pgx.Identifier{"ticket_seats"},
		[]string{"seat_number", "status", "ticket_area_id"}, // Column order must match seatData structure
		pgx.CopyFromRows(seatData),
	)

	if err != nil {
		return 0, fmt.Errorf("copyFrom failed: %w", err)
	}

	// Verify copyCount matches expected batch size
	if copyCount != int64(len(seatData)) {
		// This indicates a potential issue, treat as an error.
		return copyCount, fmt.Errorf("copyFrom reported inserting %d rows, but expected %d", copyCount, len(seatData))
	}

	return copyCount, nil
}
