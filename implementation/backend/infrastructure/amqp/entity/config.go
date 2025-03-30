package entity

import "time"

type QueueConfig struct {
	Name       string
	Durable    bool
	AutoDelete bool
	Exclusive  bool
	NoWait     bool
	Timeout    *time.Duration
}

type ExchangeConfig struct {
	Name       string
	Kind       string // direct, topic, or fanout
	Durable    bool
	AutoDelete bool
	NoWait     bool
	Internal   bool
}

type ConsumeConfig struct {
	PrefetchCount int
	PrefetchSize  int
	AutoAck       bool
	RoutingKeys   []string
}
