package idgen

import (
	"github.com/godruoyi/go-snowflake"
)

const bucket_size = 1000 * 60 * 60 * 24 * 4

func New(nodeId uint16) {
	snowflake.SetMachineID(nodeId)
}

func Next() int64 {
	return int64(snowflake.ID())
}

func GetBucket(id int64) int64 {
	return (id >> 20) / bucket_size
}
