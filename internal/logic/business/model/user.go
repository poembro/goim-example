package model

type Auth struct {
	Mid      int64   `json:"mid"`
	Key      string  `json:"key"`
	RoomID   string  `json:"room_id"`
	Platform string  `json:"platform"`
	Accepts  []int32 `json:"accepts"`
	// Token string  `json:"token"` // 授权token 这里解析 调用第三方api
}

type User struct {
	Auth
	Nickname string `json:"nickname"`  // 用户昵称
	Face     string `json:"face"`      // 用户头像
	ShopId   string `json:"shop_id"`   // 商户id
	ShopName string `json:"shop_name"` // 商户昵称
	ShopFace string `json:"shop_face"` // 商户头像
	//Platform    string   `json:"platform"`             // 平台标识
	Suburl      string   `json:"suburl"`               // websocket 订阅推送地址
	Pushurl     string   `json:"pushurl"`              // http 推送地址
	IsOnline    bool     `json:"is_online"`            // 用户在线标识
	Unread      int64    `json:"unread"`               // 未读
	LastMessage []string `json:"last_message"`         // 最后一条消息
	Referer     string   `json:"referer"`              // 来源
	UserAgent   string   `json:"user_agent"`           // 用户标识
	RemoteAddr  string   `json:"remote_addr"`          // 客户端ip
	CreatedAt   string   `json:"created_at,omitempty"` // 用户创建时间
	Token       string   `json:"token"`                // token
}
