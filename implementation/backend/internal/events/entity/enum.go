package entity

import "errors"

type AreaType string

const (
	AreaType__NumberedSeating AreaType = "numbered-seating"
	AreaType__FreeStanding    AreaType = "free-standing"
)

func (e *AreaType) Scan(value interface{}) error {
	var enumValue string
	switch val := value.(type) {
	case string:
		enumValue = val
	case []byte:
		enumValue = string(val)
	default:
		return errors.New("invalid scan value for AreaType enum. Enum value has to be of type string or []byte")
	}

	switch enumValue {
	case "numbered-seating":
		*e = AreaType__NumberedSeating
	case "free-standing":
		*e = AreaType__FreeStanding
	default:
		return errors.New("invalid scan value '" + enumValue + "' for AreaType enum")
	}

	return nil
}

func (e *AreaType) Value() string {
	return string(*e)
}

func (e *AreaType) String() string {
	return string(*e)
}

type SeatStatus string

const (
	SeatStatus__Available SeatStatus = "available"
	SeatStatus__OnHold    SeatStatus = "on-hold"
	SeatStatus__Sold      SeatStatus = "sold"
)

func (e *SeatStatus) Scan(value interface{}) error {
	var enumValue string
	switch val := value.(type) {
	case string:
		enumValue = val
	case []byte:
		enumValue = string(val)
	default:
		return errors.New("invalid scan value for SeatStatus enum. Enum value has to be of type string or []byte")
	}

	switch enumValue {
	case "available":
		*e = SeatStatus__Available
	case "on-hold":
		*e = SeatStatus__OnHold
	case "sold":
		*e = SeatStatus__Sold
	default:
		return errors.New("invalid scan value '" + enumValue + "' for SeatStatus enum")
	}

	return nil
}

func (e *SeatStatus) Value() string {
	return string(*e)
}

func (e *SeatStatus) String() string {
	return string(*e)
}
