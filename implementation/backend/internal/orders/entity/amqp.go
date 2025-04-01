package entity

import (
	"bytes"
	"encoding/gob"
	"time"
	"tugas-akhir/backend/infrastructure/amqp/entity"
	myerror "tugas-akhir/backend/pkg/error"
)

var PlaceOrderTimeout = 30 * time.Second

var PlaceOrderQueue = entity.QueueConfig{
	Name:       "place_orders",
	Durable:    true,
	AutoDelete: false,
	NoWait:     false,
	Exclusive:  false,
	Timeout:    &PlaceOrderTimeout,
}

var PlaceOrderExchange = entity.ExchangeConfig{
	Name:       "crawling",
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
		LogDelivery:  true,
	}

	return &res, nil
}

type PlaceOrderReplyMessage struct {
	Order      *Order
	HttpErr    *myerror.HttpError
	ReplyRoute string
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
		LogDelivery:  true,
	}

	return &res, nil
}
