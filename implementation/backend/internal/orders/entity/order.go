package entity

import (
	"time"
	"tugas-akhir/backend/internal/payments/entity"
)

type Order struct {
	ID                int64           `json:"id"`
	Status            OrderStatus     `json:"status"`
	FailReason        *string         `json:"failReason"`
	FirstTicketAreaID int64           `json:"firstTicketAreaId"`
	ExternalUserID    string          `json:"externalUserId"`
	CreatedAt         time.Time       `json:"createdAt"`
	UpdatedAt         time.Time       `json:"updatedAt"`
	Items             []OrderItem     `json:"items"`
	Invoice           *entity.Invoice `json:"invoice"`
}

type OrderItem struct {
	ID            int64     `json:"id"`
	CustomerName  string    `json:"customerName"`
	CustomerEmail string    `json:"customerEmail"`
	Price         int64     `json:"price"`
	OrderID       int64     `json:"orderId"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}
