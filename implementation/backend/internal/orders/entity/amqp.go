package entity

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"time"
	"tugas-akhir/backend/infrastructure/amqp/entity"
	myerror "tugas-akhir/backend/pkg/error"
)

var PlaceOrderTimeout = 130 * time.Second

var PlaceOrderQueue = entity.QueueConfig{
	Name:       "place_orders",
	Durable:    true,
	AutoDelete: false,
	NoWait:     false,
	Exclusive:  false,
	Timeout:    &PlaceOrderTimeout,
}

func NewPlaceOrderReplyQueue(identifier string) entity.QueueConfig {
	return entity.QueueConfig{
		Name:       fmt.Sprintf("place_orders_reply_%s", identifier),
		Durable:    false,
		AutoDelete: false,
		NoWait:     false,
		Exclusive:  false,
		Timeout:    &PlaceOrderTimeout,
	}
}

var PlaceOrderExchange = entity.ExchangeConfig{
	Name:       "place_order",
	Kind:       "topic",
	Durable:    true,
	AutoDelete: false,
	NoWait:     false,
	Internal:   false,
}

type PlaceOrderMessage struct {
	Data       PlaceOrderDto
	ReplyRoute string
}

func (m PlaceOrderMessage) ToMessage() (*entity.Message, error) {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)

	err := encoder.Encode(m)

	if err != nil {
		return nil, err
	}
	res := entity.Message{
		ContentType:  "application/json",
		Data:         buffer.Bytes(),
		TTL:          nil,
		RoutingKey:   "orders",
		Type:         nil,
		IsPersistent: true,
		LogDelivery:  false,
	}

	return &res, nil
}

type PlaceOrderReplyMessage struct {
	Order          *Order
	HttpErr        *myerror.HttpError
	ReplyRoute     string
	IdempotencyKey string
}

func (m PlaceOrderReplyMessage) ToMessage() (*entity.Message, error) {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)

	err := encoder.Encode(m)

	if err != nil {
		return nil, err
	}
	res := entity.Message{
		ContentType:  "application/json",
		Data:         buffer.Bytes(),
		TTL:          nil,
		RoutingKey:   m.ReplyRoute,
		Type:         nil,
		IsPersistent: true,
		LogDelivery:  false,
	}

	return &res, nil
}
