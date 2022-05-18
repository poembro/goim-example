package model

type Shop struct {
	UserId   string `json:"user_id"`  // 用户id
	Nickname string `json:"nickname"` // 昵称
	Face     string `json:"face"`     // 头像
	Password string `json:"password"` // 密码
}
