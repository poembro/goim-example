package router

import (
	"context"
	"encoding/json"
	"goim-example/internal/logic/business/model"
	"goim-example/internal/logic/business/util"

	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// ShopLogin 登录 (后台)
func (s *Router) ShopLogin(c *gin.Context) {
	var arg struct {
		Nickname string `json:"nickname"`
		Password string `json:"password"`
	}
	if err := c.BindJSON(&arg); err != nil {
		s.OutJson(c, -1, err.Error(), nil)
		return
	}
	if arg.Nickname == "" || arg.Password == "" {
		s.OutJson(c, -1, "参数nickname or password不能为空", nil)
		return
	}
	item, err := s.svc.ShopFindOne(arg.Nickname)
	if err != nil || item == nil || item.Mid == "" {
		s.OutJson(c, -1, "未注册", nil)
		return
	}

	if item.Password != arg.Password {
		s.OutJson(c, -1, "密码错误", nil)
		return
	}
	dst := s.svc.UserCreate(item, "0.0.0.0", c.GetHeader("referer"), c.GetHeader("user-agent"))
	s.OutJson(c, 200, "success", dst)
}

// UserRegister 注册 (后台) 为了演示,临时采用redis存储
func (s *Router) ShopRegister(c *gin.Context) {
	var arg struct {
		Nickname string `json:"nickname"`
		Password string `json:"password"`
	}

	if err := c.BindJSON(&arg); err != nil {
		s.OutJson(c, -1, err.Error(), nil)
		return
	}

	item, _ := s.svc.ShopFindOne(arg.Nickname)
	if item != nil {
		s.OutJson(c, -1, "已经被注册", nil)
		return
	}

	face := "https://img.wxcha.com/m00/86/59/7c6242363084072b82b6957cacc335c7.jpg"
	item, err := s.svc.ShopCreate(arg.Nickname, face, arg.Password)
	if err != nil {
		s.OutJson(c, -1, err.Error(), nil)
		return
	}
	s.OutJson(c, 200, "success", item)
}

// ShopList 查看所有与自己聊天的用户
func (s *Router) ShopList(c *gin.Context) {
	var arg struct {
		ShopId string `json:"shop_id"`
		Typ    string `json:"typ"`
	}
	if err := c.BindJSON(&arg); err != nil {
		s.OutJson(c, -1, err.Error(), nil)
		return
	}

	var (
		idsTmp []string
		total  int64
		err    error
	)

	// 查询在线人数
	page := NewPage(c)
	if arg.Typ == "offline" {
		idsTmp, total, err = s.svc.ShopByUsers(arg.ShopId,
			"-inf", "+inf", int64(page.Page), int64(page.Limit))
	} else {
		max := strconv.FormatInt(time.Now().UnixNano(), 10)
		min := strconv.FormatInt(time.Now().Add(-time.Hour*1).UnixNano(), 10)

		idsTmp, total, err = s.svc.ShopByUsers(arg.ShopId,
			min, max, int64(page.Page), int64(page.Limit))
	}
	if err != nil {
		s.OutJson(c, -1, err.Error(), nil)
		return
	}
	page.Total = total

	userIds, err := s.svc.UserFinds(idsTmp)

	// 查询已读/未读
	onlineTmp := make([]*model.User, 0)
	offlineTmp := make([]*model.User, 0)
	for deviceId, v := range userIds {
		if v == "" {
			continue
		}
		item := new(model.User)
		json.Unmarshal(util.S2B(v), item)

		tmpUid := strconv.FormatInt(int64(item.Mid), 10)
		if arg.ShopId == tmpUid {
			continue // 不要展示商户自己
		}
		// 获取消息已读偏移
		index, _ := s.svc.MsgAckMapping(deviceId, item.RoomID)   // 获取消息已读偏移
		count, err := s.svc.MsgCount(item.RoomID, index, "+inf") // 拿到偏移去统计未读
		if err != nil {
			continue
		}

		lastMessage, err := s.svc.MsgListByTop(item.RoomID, 0, 0) // 取回消息
		if err != nil {
			continue
		}

		item.Unread = count
		item.LastMessage = lastMessage

		item.IsOnline = s.logic.IsOnline(context.TODO(), []string{deviceId})
		// 在线的用户先暂存起来
		if item.IsOnline {
			onlineTmp = append(onlineTmp, item)
			continue
		}

		offlineTmp = append(offlineTmp, item)
	}

	onlineTmp = append(onlineTmp, offlineTmp...) //合并离线与在线
	s.OutPageJson(c, onlineTmp, page)
}
