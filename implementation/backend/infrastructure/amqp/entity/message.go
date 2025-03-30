package entity

import "time"

type Message struct {
	ContentType  string
	Data         []byte
	TTL          *time.Duration
	RoutingKey   string
	Type         *string
	IsPersistent bool
	LogDelivery  bool
	Priority     *uint8
}
