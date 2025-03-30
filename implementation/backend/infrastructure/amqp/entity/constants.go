package entity

import (
	"time"
)

const (
	// ReconnectDelay When reconnecting to the server after connection failure
	ReconnectDelay = 5 * time.Second

	// ReInitDelay When setting up the channel after a channel exception
	ReInitDelay = 2 * time.Second

	// ResendDelay When resending messages the server didn't confirm
	ResendDelay = 5 * time.Second
)
