package entity

import "errors"

type OrderStatus string

const (
	OrderStatus__WaitingForPayment OrderStatus = "waiting-for-payment"
	OrderStatus__Failed            OrderStatus = "failed"
	OrderStatus__Success           OrderStatus = "success"
)

func (e *OrderStatus) Scan(value interface{}) error {
	var enumValue string
	switch val := value.(type) {
	case string:
		enumValue = val
	case []byte:
		enumValue = string(val)
	default:
		return errors.New("invalid scan value for OrderStatus enum. Enum value has to be of type string or []byte")
	}

	switch enumValue {
	case "waiting-for-payment":
		*e = OrderStatus__WaitingForPayment
	case "success":
		*e = OrderStatus__Success
	case "failed":
		*e = OrderStatus__Failed
	default:
		return errors.New("invalid scan value '" + enumValue + "' for OrderStatus enum")
	}

	return nil
}

func (e *OrderStatus) Value() string {
	return string(*e)
}

func (e *OrderStatus) String() string {
	return string(*e)
}
