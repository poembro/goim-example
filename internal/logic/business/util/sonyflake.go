package util

import (
	"github.com/sony/sonyflake"
)

var (
	SFlake *SnowFlake
)

// SnowFlake SnowFlake算法结构体
type SnowFlake struct {
	sFlake *sonyflake.Sonyflake
}

func (s *SnowFlake) GetID() (uint64, error) {
	return s.sFlake.NextID()
}

func init() {
	SFlake = NewSnowFlake()
}

// 模拟获取本机的机器ID
func getMachineID() (mID uint16, err error) {
	mID = 10
	return
}

func NewSnowFlake() *SnowFlake {
	st := sonyflake.Settings{
		MachineID: getMachineID, // machineID是个回调函数
	}

	return &SnowFlake{
		sFlake: sonyflake.NewSonyflake(st),
	}
}
