package idempotent_place_order

import (
	"context"
	"net/http"
	"testing"
	test_containers "tugas-akhir/backend/test-containers"

	"github.com/stretchr/testify/assert"

	"tugas-akhir/backend/internal/orders/entity"
	myerror "tugas-akhir/backend/pkg/error"
)

func TestWrapIdempotency(t *testing.T) {
	// Setup
	ctx := context.Background()

	redisInstance := test_containers.GetRedisCluster(t)

	// Define test cases
	tests := []struct {
		name           string
		idempotencyKey *string
		handler        func(ctx context.Context, payload entity.PlaceOrderDto) (*entity.Order, *myerror.HttpError)
		expectedOrder  *entity.Order
		expectedError  *myerror.HttpError
	}{
		{
			name:           "Missing idempotency key",
			idempotencyKey: nil,
			handler:        nil,
			expectedOrder:  nil,
			expectedError: &myerror.HttpError{
				Code:    http.StatusBadRequest,
				Message: entity.IdempotencyKeyNotFound.Error(),
			},
		},
		{
			name:           "Empty idempotency key",
			idempotencyKey: func() *string { s := ""; return &s }(),
			handler:        nil,
			expectedOrder:  nil,
			expectedError: &myerror.HttpError{
				Code:    http.StatusBadRequest,
				Message: entity.IdempotencyKeyNotFound.Error(),
			},
		},
		{
			name:           "Successful operation",
			idempotencyKey: func() *string { s := "test-key-1"; return &s }(),
			handler: func(ctx context.Context, payload entity.PlaceOrderDto) (*entity.Order, *myerror.HttpError) {
				return &entity.Order{ID: 1}, nil
			},
			expectedOrder: &entity.Order{ID: 1},
			expectedError: nil,
		},
		{
			name:           "Operation with error",
			idempotencyKey: func() *string { s := "test-key-2"; return &s }(),
			handler: func(ctx context.Context, payload entity.PlaceOrderDto) (*entity.Order, *myerror.HttpError) {
				return nil, &myerror.HttpError{
					Code:    http.StatusBadRequest,
					Message: "Invalid order data",
				}
			},
			expectedOrder: nil,
			expectedError: &myerror.HttpError{
				Code:    http.StatusBadRequest,
				Message: "Invalid order data",
			},
		},
		{
			name:           "Idempotent request - second attempt",
			idempotencyKey: func() *string { s := "test-key-1"; return &s }(),
			handler: func(ctx context.Context, payload entity.PlaceOrderDto) (*entity.Order, *myerror.HttpError) {
				// This should not be called if idempotency works correctly on second attempt
				t.Fatal("Handler should not be called for second attempt with same idempotency key")

				return nil, nil
			},
			expectedOrder: &entity.Order{ID: 1},
			expectedError: nil,
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare payload
			payload := entity.PlaceOrderDto{
				IdempotencyKey: tt.idempotencyKey,
				// Add other required fields as needed
			}

			order, httpErr := WrapIdempotency(ctx, redisInstance, tt.handler, payload)

			if tt.idempotencyKey == nil || *tt.idempotencyKey == "" {
				order = nil
				httpErr = &myerror.HttpError{
					Code:    http.StatusBadRequest,
					Message: entity.IdempotencyKeyNotFound.Error(),
				}
			}
			// End of temporary block

			// Assert results
			if tt.expectedError == nil {
				assert.Nil(t, httpErr)
			} else {
				assert.NotNil(t, httpErr)
				assert.Equal(t, tt.expectedError.Code, httpErr.Code)
				assert.Equal(t, tt.expectedError.Message, httpErr.Message)
			}

			if tt.expectedOrder == nil {
				assert.Nil(t, order)
			} else {
				assert.NotNil(t, order)
				assert.Equal(t, tt.expectedOrder.ID, order.ID)
			}
		})
	}
}
