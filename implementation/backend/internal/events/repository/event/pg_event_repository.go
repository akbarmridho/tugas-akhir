package event

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
	"tugas-akhir/backend/infrastructure/memcache"
	"tugas-akhir/backend/infrastructure/postgres"
	"tugas-akhir/backend/internal/events/entity"
	"tugas-akhir/backend/pkg/logger"
)

type PGEventRepository struct {
	db    *postgres.Postgres
	cache *memcache.Memcache
}

func NewPGEventRepository(
	db *postgres.Postgres,
	cache *memcache.Memcache,
) *PGEventRepository {
	return &PGEventRepository{
		db:    db,
		cache: cache,
	}
}

const EventsCacheKey = "events"

func eventCacheKey(id int64) string {
	return fmt.Sprintf("%s:%d", "events:", id)
}

func (r *PGEventRepository) GetEvents(ctx context.Context) ([]entity.Event, error) {
	result := make([]entity.Event, 0)

	cache, cacheErr := r.cache.Cache.Get(EventsCacheKey)

	if cacheErr != nil {
		logger.FromCtx(ctx).Error("Cannot get events from cache")
	} else {
		marshallErr := json.Unmarshal(cache, &result)

		if marshallErr == nil {
			return result, nil
		} else {
			logger.FromCtx(ctx).Error("Cannot unmashall cached events")
		}
	}

	query := `SELECT * FROM events`

	err := pgxscan.Select(
		ctx,
		r.db.GetExecutor(ctx),
		&result,
		query,
	)

	if err != nil {
		return nil, err
	}

	raw, err := json.Marshal(result)

	if err != nil {
		logger.FromCtx(ctx).Error("Cannot marshall events")
	}

	if setCacheErr := r.cache.Cache.Set(EventsCacheKey, raw); setCacheErr != nil {
		logger.FromCtx(ctx).Error("Cannot set cache events", zap.Error(setCacheErr))
	}

	return result, nil
}

func (r *PGEventRepository) GetEvent(ctx context.Context, payload entity.GetEventDto) (*entity.Event, error) {
	var event entity.Event

	cache, cacheErr := r.cache.Cache.Get(eventCacheKey(payload.ID))

	if cacheErr != nil {
		logger.FromCtx(ctx).Sugar().Errorf("Cannot get event %d from cache", payload.ID)
	} else {
		marshallErr := json.Unmarshal(cache, &event)

		if marshallErr == nil {
			return &event, nil
		} else {
			logger.FromCtx(ctx).Sugar().Errorf("Cannot unmashall cached event %d", payload.ID)
		}
	}

	query := `
	WITH event_data AS (
		SELECT
			e.id, e.name, e.location, e.description, e.created_at, e.updated_at
		FROM events e
		WHERE e.id = $1
	),
	ticket_sales_data AS (
		SELECT
			ts.id, ts.name, ts.sale_begin_at, ts.sale_end_at, ts.event_id,
			ts.created_at, ts.updated_at
		FROM ticket_sales ts
		WHERE ts.event_id = $1
	),
	ticket_packages_data AS (
		SELECT
			tp.id, tp.price, tp.ticket_category_id, tp.ticket_sale_id,
			tp.created_at, tp.updated_at
		FROM ticket_packages tp
		JOIN ticket_sales ts ON tp.ticket_sale_id = ts.id
		WHERE ts.event_id = $1
	),
	ticket_categories_data AS (
		SELECT
			tc.id, tc.name, tc.event_id, tc.created_at, tc.updated_at
		FROM ticket_categories tc
		WHERE tc.event_id = $1
	),
	ticket_areas_data AS (
		SELECT
			ta.id, ta.type, ta.ticket_package_id, ta.created_at, ta.updated_at
		FROM ticket_areas ta
		JOIN ticket_packages tp ON ta.ticket_package_id = tp.id
		JOIN ticket_sales ts ON tp.ticket_sale_id = ts.id
		WHERE ts.event_id = $1
	)
	SELECT
		json_build_object(
			'id', e.id,
			'name', e.name,
			'location', e.location,
			'description', e.description,
			'createdAt', e.created_at,
			'updatedAt', e.updated_at,
			'ticketSales', COALESCE(
				(
					SELECT json_agg(
						json_build_object(
							'id', ts.id,
							'name', ts.name,
							'saleBeginAt', ts.sale_begin_at,
							'saleEndAt', ts.sale_end_at,
							'eventId', ts.event_id,
							'createdAt', ts.created_at,
							'updatedAt', ts.updated_at,
							'ticketPackages', COALESCE(
								(
									SELECT json_agg(
										json_build_object(
											'id', tp.id,
											'price', tp.price,
											'ticketCategoryId', tp.ticket_category_id,
											'ticketSaleId', tp.ticket_sale_id,
											'createdAt', tp.created_at,
											'updatedAt', tp.updated_at,
											'ticketCategory', (
												SELECT json_build_object(
													'id', tc.id,
													'name', tc.name,
													'eventId', tc.event_id,
													'createdAt', tc.created_at,
													'updatedAt', tc.updated_at
												)
												FROM ticket_categories_data tc
												WHERE tc.id = tp.ticket_category_id
											),
											'ticketAreas', COALESCE(
												(
													SELECT json_agg(
														json_build_object(
															'id', ta.id,
															'type', ta.type,
															'ticketPackageId', ta.ticket_package_id,
															'createdAt', ta.created_at,
															'updatedAt', ta.updated_at
														)
													)
													FROM ticket_areas_data ta
													WHERE ta.ticket_package_id = tp.id
												),
												'[]'::json
											)
										)
									)
									FROM ticket_packages_data tp
									WHERE tp.ticket_sale_id = ts.id
								),
								'[]'::json
							)
						)
					)
					FROM ticket_sales_data ts
				),
				'[]'::json
			)
		) as event_json
	FROM event_data e;
    `

	var eventJSON json.RawMessage
	err := r.db.GetExecutor(ctx).QueryRow(ctx, query, payload.ID).Scan(&eventJSON)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entity.EventNotFoundError
		}
		return nil, err
	}

	if setCacheErr := r.cache.Cache.Set(eventCacheKey(payload.ID), eventJSON); setCacheErr != nil {
		logger.FromCtx(ctx).Error("Cannot set cache events")
	}

	if err := json.Unmarshal(eventJSON, &event); err != nil {
		return nil, err
	}

	return &event, nil
}
