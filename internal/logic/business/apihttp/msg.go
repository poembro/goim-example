package apihttp

import (
	"context"
	"fmt"
	"goim-demo/internal/logic/business/model"
	"goim-demo/internal/logic/business/util"
	utilModel "goim-demo/internal/logic/model"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// ListMsg 查历史消息
func (s *Router) ListMsg(c *gin.Context) {
	var arg struct {
		RoomId string `json:"room_id"`
	}
	if err := c.BindJSON(&arg); err != nil {
		OutJson(c, OutData{Code: -1, Success: false, Msg: err.Error()})
		return
	}

	page := NewPage(c)
	dst, total, err := s.svc.GetMessagePageList(arg.RoomId, "-inf", "+inf", int64(page.Page), int64(page.Limit))
	if err != nil {
		OutJson(c, OutData{Code: -1, Success: false, Msg: err.Error()})
		return
	}
	page.Total = total

	OutPageJson(c, dst, page)
}

// apiClearData 数据清理
func (s *Router) ClearMsg(c *gin.Context) {
	s.svc.ClearMsg(context.TODO())
	OutJson(c, OutData{Code: 200, Success: true})
}

// apiPush 数据推送
func (s *Router) PushMsg(c *gin.Context) {
	var arg struct {
		RoomId string `json:"room_id"`
		Typ    string `json:"type"`
		Msg    string `json:"msg"`
		Mid    string `json:"mid"`
		ShopId string `json:"shop_id"`
	}
	if err := c.BindJSON(&arg); err != nil {
		OutJson(c, OutData{Code: -1, Success: false, Msg: err.Error()})
		return
	}
	if arg.RoomId == "" || arg.Typ == "" || arg.Msg == "" || arg.Mid == "" || arg.ShopId == "" {
		OutJson(c, OutData{Code: -1, Success: false, Msg: "参数room_id type msg user_id shop_id不能为空"})
		return
	}
	// 处理特殊字符
	msg := strings.Replace(arg.Msg, "\r\n", "\\r\\n", -1)
	msg = strings.Replace(arg.Msg, "\r", "\\r", -1)
	msg = strings.Replace(arg.Msg, "\n", "\\n", -1)

	msgId := time.Now().UnixNano() // 消息唯一id 为了方便临时demo采用该方案， 后期线上可以用雪花算法
	body := fmt.Sprintf(`{"mid":%s, "shop_id":%s, "type":"%s", "msg":"%s", "room_id":"%s", "dateline":%d, "id":"%d"}`,
		arg.Mid, arg.ShopId, arg.Typ, msg, arg.RoomId, time.Now().Unix(), msgId)

	// 消息持久化
	err := s.svc.AddMessageList(arg.RoomId, msgId, body)
	if err != nil {
		OutJson(c, OutData{Code: -1, Success: false, Msg: err.Error()})
		return
	}

	typ, room, _ := utilModel.DecodeRoomKey(arg.RoomId)

	// 推送
	if err := s.logic.PushRoom(c, model.OpMessage, typ, room, util.S2B(body)); err != nil {
		OutJson(c, OutData{Code: -1, Success: false, Msg: err.Error()})
		return
	}
	OutJson(c, OutData{Code: 200, Success: true, Msg: "success"})
}
