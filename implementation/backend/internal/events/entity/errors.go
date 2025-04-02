package entity

import "errors"

var EventNotFoundError = errors.New("event not found")

var AreaAvailabilityNotFoundError = errors.New("area availability not found error")

var SeatNotFoundError = errors.New("seats not found error")

var InternalOrderConfigurationError = errors.New("internal order configuration error")

var PlaceOrderBadRequest = errors.New("seating configuration error")
