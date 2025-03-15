package entity

import "time"

type Invoice struct {
	ID         int64         `json:"id"`
	Status     InvoiceStatus `json:"status"`
	Amount     int32         `json:"amount"`
	ExternalID *string       `json:"externalId"`
	OrderID    int64         `json:"orderId"`
	CreatedAt  time.Time     `json:"createdAt"`
	UpdatedAt  time.Time     `json:"updatedAt"`
}
