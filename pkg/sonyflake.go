package pkg

import (
	"errors"

	"github.com/sony/sonyflake"
)

// NewSonyFlake is the factory of sonyflake instance
func NewSonyFlake() (*sonyflake.Sonyflake, error) {
	var st sonyflake.Settings
	sf := sonyflake.NewSonyflake(st)
	if sf == nil {
		return nil, errors.New("sonyflake not created")
	}
	return sf, nil
}
