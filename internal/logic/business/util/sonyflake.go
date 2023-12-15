package util

import (
	"math/rand"

	"github.com/sony/sonyflake"
)

var (
	SFlake *SnowFlake
)

func init() {
	n := rand.Intn(15)
	if n <= 0 {
		n++
	}
	SFlake = NewSnowFlake(uint16(n))
}

// SnowFlake SnowFlake算法结构体
type SnowFlake struct {
	sFlake *sonyflake.Sonyflake
}

func (s *SnowFlake) GetID() uint64 {
	n, err := s.sFlake.NextID()
	if err != nil {
		return 0
	}
	return n
}

func NewSnowFlake(id uint16) *SnowFlake {
	st := sonyflake.Settings{
		MachineID: func() (mID uint16, err error) { // 回调函数 返回当前机器编号
			mID = id
			return
		},
	}

	return &SnowFlake{
		sFlake: sonyflake.NewSonyflake(st),
	}
}
