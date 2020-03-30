package utils

import (
	"github.com/sony/sonyflake"
)

var (
	snow = sonyflake.NewSonyflake(sonyflake.Settings{})
)

// SnowID returns random id
func SnowID() uint64 {
	id, _ := snow.NextID()
	return id % 20000000
}
