package apihttp

import (
	"context"
	"encoding/json"
	"goim-demo/internal/logic/business/model"
	"goim-demo/internal/logic/business/util"

	//"goim-demo/pkg/logger"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// UserCreate 创建用户
func (s *Router) UserCreate(c *gin.Context) {
	var arg struct {
		ShopId string `form:"shop_id"`
	}
	if err := c.BindQuery(&arg); err != nil {
		OutJson(c, OutData{Code: -1, Success: false, Msg: err.Error()})
		return
	}
	if arg.ShopId == "" {
		OutJson(c, OutData{Code: -1, Success: false, Msg: "参数不能为空"})
		return
	}
	//判断客服是否存在
	shop, err := s.svc.GetShop(arg.ShopId) // ShopId 就是商户昵称
	if err != nil || shop == nil || shop.Mid == "" {
		OutJson(c, OutData{Code: -1, Success: false, Msg: "参数错误"})
		return
	}
	dst := s.svc.UserCreate(shop.Mid, shop.Nickname, shop.Face,
		c.ClientIP(), c.GetHeader("referer"), c.GetHeader("user-agent"))

	// 客服聊天场景
	OutJson(c, OutData{Code: 200, Success: true, Msg: "success", Result: dst})
}

// Login 登录 (后台)
func (s *Router) Login(c *gin.Context) {
	var arg struct {
		Nickname string `json:"nickname"`
		Password string `json:"password"`
	}
	if err := c.BindJSON(&arg); err != nil {
		OutJson(c, OutData{Code: -1, Success: false, Msg: err.Error()})
		return
	}
	if arg.Nickname == "" || arg.Password == "" {
		OutJson(c, OutData{Code: -1, Success: false, Msg: "参数nickname or password不能为空"})
		return
	}
	shop, err := s.svc.GetShop(arg.Nickname)
	if err != nil || shop == nil || shop.Mid == "" {
		OutJson(c, OutData{Code: -1, Success: false, Msg: "未注册"})
		return
	}

	if shop.Password != arg.Password {
		OutJson(c, OutData{Code: -1, Success: false, Msg: "密码错误"})
		return
	}
	dst := s.svc.ShopCreate(shop.Mid, shop.Nickname, shop.Face,
		c.ClientIP(), c.GetHeader("referer"), c.GetHeader("user-agent"))
	OutJson(c, OutData{Code: 200, Success: true, Result: dst})
}

// Register 注册 (后台) 为了演示,临时采用redis存储
func (s *Router) Register(c *gin.Context) {
	var arg struct {
		Nickname string `json:"nickname"`
		Password string `json:"password"`
	}
	if err := c.BindJSON(&arg); err != nil {
		OutJson(c, OutData{Code: -1, Success: false, Msg: err.Error()})
		return
	}
	shop, _ := s.svc.GetShop(arg.Nickname)
	if shop != nil {
		OutJson(c, OutData{Code: -1, Success: false, Msg: "已经被注册"})
		return
	}

	face := "https://img.wxcha.com/m00/86/59/7c6242363084072b82b6957cacc335c7.jpg"
	_, mid := s.svc.BuildMid()
	s.svc.AddShop(mid, arg.Nickname, face, arg.Password)

	OutJson(c, OutData{Code: 200, Success: true, Msg: "success", Result: "xxx"})
}

// UserList 查看所有与自己聊天的用户
func (s *Router) UserList(c *gin.Context) {
	var arg struct {
		ShopId string `json:"shop_id"`
		Typ    string `json:"typ"`
	}
	if err := c.BindJSON(&arg); err != nil {
		OutJson(c, OutData{Code: -1, Success: false, Msg: err.Error()})
		return
	}

	var (
		idsTmp []string
		total  int64
		err    error
	)

	ids := make([]int64, 0)
	// 查询在线人数
	page := NewPage(c)

	if arg.Typ == "offline" {
		idsTmp, total, err = s.svc.GetShopByUsers(arg.ShopId,
			"-inf", "+inf", int64(page.Page), int64(page.Limit))
	} else {
		max := strconv.FormatInt(time.Now().UnixNano(), 10)
		min := strconv.FormatInt(time.Now().Add(-time.Hour*1).UnixNano(), 10)

		idsTmp, total, err = s.svc.GetShopByUsers(arg.ShopId,
			min, max, int64(page.Page), int64(page.Limit))
	}
	if err != nil {
		OutJson(c, OutData{Code: -1, Success: false, Msg: err.Error()})
		return
	}
	page.Total = total

	for _, v := range idsTmp {
		id, _ := strconv.ParseInt(v, 10, 64)
		ids = append(ids, id)
	}
	userIds, err := s.svc.KeysByUserIds(ids)

	// 查询已读/未读
	onlineTmp := make([]*model.User, 0)
	offlineTmp := make([]*model.User, 0)
	for deviceId, v := range userIds {
		if v == "" {
			continue
		}
		user := new(model.User)
		json.Unmarshal(util.S2B(v), user)

		tmpUid := strconv.FormatInt(int64(user.Mid), 10)
		if arg.ShopId == tmpUid {
			continue // 不要展示商户自己
		}
		// 获取消息已读偏移
		index, _ := s.svc.GetMessageAckMapping(deviceId, user.RoomID) // 获取消息已读偏移

		count, err := s.svc.GetMessageCount(user.RoomID, index, "+inf") // 拿到偏移去统计未读
		if err != nil {
			continue
		}

		lastMessage, err := s.svc.GetMessageList(user.RoomID, 0, 0) // 取回消息
		if err != nil {
			continue
		}

		user.Unread = model.Int64(count)
		user.LastMessage = lastMessage

		user.IsOnline = s.logic.IsOnline(context.TODO(), []string{deviceId})
		// 在线的用户先暂存起来
		if user.IsOnline {
			onlineTmp = append(onlineTmp, user)
			continue
		}

		offlineTmp = append(offlineTmp, user)
	}

	onlineTmp = append(onlineTmp, offlineTmp...) //合并离线与在线
	OutPageJson(c, onlineTmp, page)
}
