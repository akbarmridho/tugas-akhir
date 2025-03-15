package entity

import "errors"

type InvoiceStatus string

const (
	OrderStatus__Pending InvoiceStatus = "pending"
	OrderStatus__Expired InvoiceStatus = "expired"
	OrderStatus__Paid    InvoiceStatus = "paid"
)

func (e *InvoiceStatus) Scan(value interface{}) error {
	var enumValue string
	switch val := value.(type) {
	case string:
		enumValue = val
	case []byte:
		enumValue = string(val)
	default:
		return errors.New("invalid scan value for InvoiceStatus enum. Enum value has to be of type string or []byte")
	}

	switch enumValue {
	case "pending":
		*e = OrderStatus__Pending
	case "expired":
		*e = OrderStatus__Expired
	case "paid":
		*e = OrderStatus__Paid
	default:
		return errors.New("invalid scan value '" + enumValue + "' for InvoiceStatus enum")
	}

	return nil
}

func (e *InvoiceStatus) Value() string {
	return string(*e)
}

func (e *InvoiceStatus) String() string {
	return string(*e)
}
