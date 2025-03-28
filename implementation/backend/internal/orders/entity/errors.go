package entity

import "errors"

var OrderPlacementInternalError = errors.New("internal order placement configuration")

var OrderFetchInternalError = errors.New("internal order fetch configuration")

var OrderNotFoundError = errors.New("order not found")

var TicketSaleNotFoundError = errors.New("ticket sale not found")

var WebhookInternalError = errors.New("webhook internal error")
