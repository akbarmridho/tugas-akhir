package entity

import (
	"time"
	"tugas-akhir/backend/internal/events/entity"
	entity2 "tugas-akhir/backend/internal/orders/entity"
)

type IssuedTicket struct {
	ID           int64     `json:"id"`
	SerialNumber string    `json:"serialNumber"`
	HolderName   string    `json:"holderName"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	TicketSeatID int64     `json:"ticketSeatId"`
	OrderID      int64     `json:"orderId"`
	OrderItemID  int64     `json:"orderItemId"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`

	// Relations
	TicketSeat entity.TicketSeat `json:"ticketSeat" db:"ticket_seat"`
}

type SeatInfoDto struct {
	CategoryName string
	SeatType     entity.AreaType
	SeatNumber   string
}

type PublishIssuedTicketDto struct {
	EventName      string
	TicketSaleName string
	TicketAreaID   int64
	SeatInfos      []SeatInfoDto
	Items          []entity2.OrderItem
}

type GetIssuedTicketDto struct {
	CompositePK  string `param:"id"`
	OrderID      int64
	TicketAreaID int64
	UserID       *string `json:"userId"`
}
