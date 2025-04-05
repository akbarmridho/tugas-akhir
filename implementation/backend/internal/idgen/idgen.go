package idgen

import "github.com/sony/sonyflake"

type Idgen struct {
	sonyflake *sonyflake.Sonyflake
}

func NewIdgen() *Idgen {
	sony := sonyflake.NewSonyflake(sonyflake.Settings{})

	return &Idgen{
		sonyflake: sony,
	}
}

func (i *Idgen) Next() (int64, error) {
	raw, err := i.sonyflake.NextID()

	if err != nil {
		return 0, err
	}

	return int64(raw), nil
}
