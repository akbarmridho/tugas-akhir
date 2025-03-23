package entity

import "time"

type Event struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Location    string    `json:"location"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`

	// relations
	TicketSales []TicketSale `json:"ticketSales,omitempty"`
}

type TicketCategory struct {
	ID        int64     `json:"id"`
	Name      int64     `json:"name"`
	EventID   int64     `json:"eventId"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type TicketSale struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	SaleBeginAt time.Time `json:"saleBeginAt"`
	SaleEndAt   time.Time `json:"saleEndAt"`
	EventID     int64     `json:"eventId"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`

	// Relations
	TicketPackages []TicketPackage `json:"ticketPackages"`
}

type TicketPackage struct {
	ID               int64     `json:"id"`
	Price            int32     `json:"price"`
	TicketCategoryID int64     `json:"ticketCategoryId"`
	TicketSaleID     int64     `json:"ticketSaleId"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`

	// Relations
	TicketCategory TicketCategory `json:"ticketCategory"`
	TicketAreas    []TicketArea   `json:"ticketAreas"`
}

type TicketArea struct {
	ID              int64     `json:"id"`
	Type            AreaType  `json:"type"`
	TicketPackageID int64     `json:"ticketPackageID"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`

	// Relations
	TicketSeats []TicketSeat `json:"ticketSeats"`
}

type TicketSeat struct {
	ID           int64      `json:"id"`
	SeatNumber   string     `json:"seatNumber"`
	Status       SeatStatus `json:"status"`
	TicketAreaID int64      `json:"ticketAreaId"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`

	// Relations
	TicketArea *TicketArea `json:"ticketArea,omitempty"`
}

type AreaAvailability struct {
	TicketPackageID int64 `json:"ticketPackageId"`
	TicketAreaID    int64 `json:"ticketAreaId"`
	TotalSeats      int32 `json:"totalSeats"`
	AvailableSeats  int32 `json:"availableSeats"`
}

type GetEventDto struct {
	ID int64 `param:"eventId"`
}

type GetAvailabilityDto struct {
	TicketSaleID int64 `param:"ticketSaleId"`
}

type GetSeatsDto struct {
	TicketAreaID int64 `param:"ticketAreaId"`
}
