package model

type Shop struct {
	Mid      string `json:"mid"`      // 用户id
	Nickname string `json:"nickname"` // 昵称
	Face     string `json:"face"`     // 头像
	Password string `json:"password"` // 密码
}
