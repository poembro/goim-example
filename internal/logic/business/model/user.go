package model

import (
	"encoding/json"
	"fmt"
	"goim-demo/internal/logic/business/util"
)

type Int64 int64

type User struct {
	Mid      Int64   `json:"mid,string"`
	Key      string  `json:"key"`     // 用户设备标识
	RoomID   string  `json:"room_id"` // 房间号
	Platform string  `json:"platform"`
	Accepts  []int32 `json:"accepts"`

	Nickname string `json:"nickname"`  // 用户昵称
	Face     string `json:"face"`      // 用户头像
	ShopId   string `json:"shop_id"`   // 商户id
	ShopName string `json:"shop_name"` // 商户昵称
	ShopFace string `json:"shop_face"` // 商户头像
	//Platform    string   `json:"platform"`             // 平台标识
	Suburl      string   `json:"suburl"`               // websocket 订阅推送地址
	Pushurl     string   `json:"pushurl"`              // http 推送地址
	IsOnline    bool     `json:"is_online"`            // 用户在线标识
	Unread      Int64    `json:"unread"`               // 未读
	LastMessage []string `json:"last_message"`         // 最后一条消息
	Referer     string   `json:"referer"`              // 来源
	UserAgent   string   `json:"user_agent"`           // 用户标识
	RemoteAddr  string   `json:"remote_addr"`          // 客户端ip
	CreatedAt   string   `json:"created_at,omitempty"` // 用户创建时间
	Token       string   `json:"token"`                // token
}

func (u *Int64) UnmarshalJSON(bs []byte) error {
	var i int64
	if err := json.Unmarshal(bs, &i); err == nil {
		*u = Int64(i)
		return nil
	}
	var s string
	if err := json.Unmarshal(bs, &s); err != nil {
		return fmt.Errorf("expected a string or an integer")
	}
	if err := json.Unmarshal(util.S2B(s), &i); err != nil {
		return err
	}
	*u = Int64(i)
	return nil
}
