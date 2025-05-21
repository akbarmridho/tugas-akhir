package entity

import "tugas-akhir/backend/internal/events/entity"

type BookingRequestDto struct {
	SeatIDs       []int64
	TicketAreaIDs []int64
	TicketAreaID  int64
}

type UpdateSeatStatusDto struct {
	SeatIDs []int64
	Status  entity.SeatStatus
}
