package router

import (
	"fmt"
	"goim-example/internal/logic/business/model"
	"goim-example/internal/logic/business/util"
	utilModel "goim-example/internal/logic/model"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// MsgList 查历史消息
func (s *Router) MsgList(c *gin.Context) {
	var arg struct {
		RoomId string `json:"room_id"`
	}
	if err := c.BindJSON(&arg); err != nil {
		s.OutJson(c, -1, err.Error(), nil)
		return
	}

	page := NewPage(c)
	dst, total, err := s.svc.MsgList(c.Request.Context(), arg.RoomId, "-inf", "+inf", int64(page.Page), int64(page.Limit))
	if err != nil {
		s.OutJson(c, -1, err.Error(), nil)
		return
	}
	page.Total = total

	s.OutPageJson(c, dst, page)
}

// MsgClear 数据清理
func (s *Router) MsgClear(c *gin.Context) {
	s.svc.MsgClear(c.Request.Context())
	s.OutJson(c, 200, "success", nil)
}

// apiPush 数据推送
func (s *Router) MsgPush(c *gin.Context) {
	var arg struct {
		RoomId string `json:"room_id"`
		Typ    string `json:"type"`
		Msg    string `json:"msg"`
		Mid    string `json:"mid"`
		ShopId string `json:"shop_id"`
	}
	if err := c.BindJSON(&arg); err != nil {
		s.OutJson(c, -1, err.Error(), nil)
		return
	}
	if arg.RoomId == "" || arg.Typ == "" || arg.Msg == "" || arg.Mid == "" || arg.ShopId == "" {
		s.OutJson(c, -1, "参数room_id type msg user_id shop_id不能为空", nil)
		return
	}
	// 处理特殊字符
	msg := strings.Replace(arg.Msg, "\r\n", "\\r\\n", -1)
	msg = strings.Replace(arg.Msg, "\r", "\\r", -1)
	msg = strings.Replace(arg.Msg, "\n", "\\n", -1)

	msgId := time.Now().UnixNano() // 消息唯一id 为了方便临时example采用该方案， 后期线上可以用雪花算法
	body := fmt.Sprintf(`{"mid":%s, "shop_id":%s, "type":"%s", "msg":"%s", "room_id":"%s", "dateline":%d, "id":"%d"}`,
		arg.Mid, arg.ShopId, arg.Typ, msg, arg.RoomId, time.Now().Unix(), msgId)

	// 消息持久化
	err := s.svc.MsgPush(c.Request.Context(), arg.RoomId, msgId, body)
	if err != nil {
		s.OutJson(c, -1, err.Error(), nil)
		return
	}

	typ, room, _ := utilModel.DecodeRoomKey(arg.RoomId)

	// 推送
	if err := s.l.PushRoom(c, model.OpMessage, typ, room, util.S2B(body)); err != nil {
		s.OutJson(c, -1, err.Error(), nil)
		return
	}
	s.OutJson(c, 200, "success", nil)
}
