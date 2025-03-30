package entity

import (
	"time"
	entity2 "tugas-akhir/backend/internal/events/entity"
	"tugas-akhir/backend/internal/payments/entity"
)

type Order struct {
	ID                int64       `json:"id"`
	Status            OrderStatus `json:"status"`
	FailReason        *string     `json:"failReason"`
	EventID           int64       `json:"eventId"`
	TicketSaleID      int64       `json:"ticketSaleId"`
	FirstTicketAreaID int64       `json:"firstTicketAreaId"`
	ExternalUserID    string      `json:"externalUserId"`
	CreatedAt         time.Time   `json:"createdAt"`
	UpdatedAt         time.Time   `json:"updatedAt"`

	// relations
	Items      []OrderItem         `json:"items"`
	Invoice    *entity.Invoice     `json:"invoice"`
	Event      *entity2.Event      `json:"event,omitempty"`
	TicketSale *entity2.TicketSale `json:"ticketSale,omitempty"`
}

type OrderItem struct {
	ID               int64     `json:"id"`
	CustomerName     string    `json:"customerName"`
	CustomerEmail    string    `json:"customerEmail"`
	Price            int64     `json:"price"`
	OrderID          int64     `json:"orderId"`
	TicketCategoryID int64     `json:"ticketCategoryId"`
	TicketSeatID     int64     `json:"ticketSeatId"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`

	// relation
	TicketSeat     *entity2.TicketSeat     `json:"ticketSeat,omitempty"`
	TicketCategory *entity2.TicketCategory `json:"ticketCategory,omitempty"`
}

type OrderItemDto struct {
	CustomerName     string `json:"customerName" validate:"required"`
	CustomerEmail    string `json:"customerEmail" validate:"required,email"`
	TicketSeatID     int64  `json:"ticketSeatId" validate:"required"`
	Price            *int32
	TicketCategoryID *int64
}

type PlaceOrderDto struct {
	UserID            *string
	IdempotencyKey    *string
	EventID           int64 `json:"eventId"`
	TicketSaleID      int64 `json:"ticketSaleId"`
	FirstTicketAreaID *int64
	Items             []OrderItemDto `json:"items" validate:"required,min=1,max=5"`
}

type GetOrderDto struct {
	OrderID      int64 `param:"id" validate:"required"`
	UserID       *string
	BypassUserID bool
}

type UpdateOrderStatusDto struct {
	OrderID    int64
	Status     OrderStatus
	FailReason *string
}
