package entity

import (
	"time"
	"tugas-akhir/backend/internal/events/entity"
)

type IssuedTicket struct {
	ID           int64     `json:"id"`
	SerialNumber string    `json:"serialNumber"`
	HolderName   string    `json:"holderName"`
	SeatID       int64     `json:"seatId"`
	OrderID      int64     `json:"orderId"`
	OrderItemID  int64     `json:"orderItemId"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`

	// Relations
	TicketSeat entity.TicketSeat `json:"ticketSeat"`
}
