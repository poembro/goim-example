package apihttp

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// apiClearData 数据清理
func (s *Router) apiClearData(c *gin.Context) {
	s.svc.ClearData(context.TODO())
	OutJson(c, OutData{Code: 200, Success: true})
}

// apiPush 数据推送
func (s *Router) apiPush(c *gin.Context) {
	var (
		msg   string
		msgId int64
	)
	var arg struct {
		RoomId string `form:"room_id"`
		Typ    string `form:"typ"`
		Msg    string `form:"msg"`
		UserId string `form:"user_id"`
		ShopId string `form:"shop_id"`
	}
	if err := c.BindQuery(&arg); err != nil {
		OutJson(c, OutData{Code: -1, Success: false, Msg: err.Error()})
		return
	}
	if arg.RoomId == "" || arg.Typ == "" || arg.Msg == "" || arg.UserId == "" || arg.ShopId == "" {
		OutJson(c, OutData{Code: -1, Success: false, Msg: "参数room_id type msg user_id shop_id不能为空"})
		return
	}

	msg = strings.Replace(arg.Msg, "\r\n", "\\r\\n", -1)
	msg = strings.Replace(arg.Msg, "\r", "\\r", -1)
	msg = strings.Replace(arg.Msg, "\n", "\\n", -1)
	msgId = time.Now().UnixNano() // 消息唯一id 为了方便临时demo采用该方案， 后期线上可以用雪花算法
	body := fmt.Sprintf(`{"user_id":%s, "shop_id":%s, "type":"%s", "msg":"%s", "room_id":"%s", "dateline":%d, "id":"%d"}`,
		arg.UserId, arg.ShopId, arg.Typ, msg, arg.RoomId, time.Now().Unix(), msgId)

	// 消息持久化
	err := s.svc.AddMessageList(arg.RoomId, msgId, body)
	if err != nil {
		OutJson(c, OutData{Code: -1, Success: false, Msg: err.Error()})
		return
	}
	OutJson(c, OutData{Code: 200, Success: true, Msg: "success"})
}
