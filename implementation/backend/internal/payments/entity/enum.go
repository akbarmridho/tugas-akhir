package entity

import "errors"

type InvoiceStatus string

const (
	InvoiceStatus__Pending InvoiceStatus = "pending"
	InvoiceStatus__Expired InvoiceStatus = "expired"
	InvoiceStatus__Failed  InvoiceStatus = "failed"
	InvoiceStatus__Paid    InvoiceStatus = "paid"
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
		*e = InvoiceStatus__Pending
	case "expired":
		*e = InvoiceStatus__Expired
	case "paid":
		*e = InvoiceStatus__Paid
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
