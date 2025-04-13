package entity

import "errors"

var OrderPlacementInternalError = errors.New("internal order placement configuration")

var OrderFetchInternalError = errors.New("internal order fetch configuration")

var OrderNotFoundError = errors.New("order not found")

var TicketSaleNotFoundError = errors.New("ticket sale not found")

var WebhookInternalError = errors.New("webhook internal error")

var TicketSaleNotStartedError = errors.New("ticket sale is not yet started")

var TicketSaleEndedError = errors.New("ticket sale is ended")

var IdempotencyKeyNotFound = errors.New("idempotency key not found")

var PlaceOrderTimeoutError = errors.New("place order timed out")

var PlaceOrderCancelled = errors.New("place order cancelled")

var DropperSeatNotAvailable = errors.New("seat not available")

var DropperInternalError = errors.New("dropper internal error")

var LockAlreadyReleased = errors.New("lock already released")

var CannotAcquireLock = errors.New("cannot acquire lock")

var OrderSameAreaID = errors.New("orders must belongs to the same ticket area id")
