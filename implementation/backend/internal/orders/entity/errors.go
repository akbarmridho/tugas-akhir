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
