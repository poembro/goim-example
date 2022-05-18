package apihttp

import (
	"encoding/json"
	"goim-demo/internal/logic/model"
	"goim-demo/pkg/logger"
	"goim-demo/internal/logic/business/util"
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"
)

// apiFindUserInfo 用户详细
func (s *Router) apiFindUserInfo(c *gin.Context) {
	var (
		userId string
		uInfo  model.User
	)

	if r.Method == "POST" {
		userId = r.FormValue("user_id")
	}
	if userId == "" {
		OutJson(c, OutData{Code: -1, Success: false, Msg: "参数user_id不能为空"})
		return
	}

	uid64, err := strconv.ParseInt(userId, 10, 64)
	userIds, err := svc.KeysByUserIds([]int64{uid64})

	// 查询已读/未读
	for _, v := range userIds {
		if v == "" {
			continue
		}
		logger.Logger.Debug("apiFindUserInfo", zap.Any("userJson", v))
		json.Unmarshal(util.S2B(v), &uInfo)
	}

	if err != nil {
		OutJson(c, OutData{Code: -1, Success: false, Msg: err.Error()})
		return
	}
	OutJson(c, OutData{Code: 200, Success: true, Result: uInfo})
}

// apiFindUserList 查看所有与自己聊天的用户
func (s *Router) apiFindUserList(c *gin.Context) {
	var (
		typ    string
		shopId string

		idsTmp []string
		total  int64
		err    error
	)

	if r.Method == "POST" {
		shopId = r.FormValue("shop_id")
		typ = r.FormValue("typ")
	}
	if shopId == "" {
		OutJson(c, OutData{Code: -1, Success: false, Msg: "参数shop_id不能为空"})
		return
	}
	ids := make([]int64, 0)
	// 查询在线人数
	page := NewPage(r)

	if typ == "offline" {
		idsTmp, total, err = svc.GetShopByUsers(shopId,
			"-inf", "+inf", int64(page.Page), int64(page.Limit))
	} else {
		max := strconv.FormatInt(time.Now().UnixNano(), 10)
		min := strconv.FormatInt(time.Now().Add(-time.Hour*1).UnixNano(), 10)

		idsTmp, total, err = svc.GetShopByUsers(shopId,
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
	userIds, err := svc.KeysByUserIds(ids)

	// 查询已读/未读
	onlineTmp := make([]*model.User, 0)
	offlineTmp := make([]*model.User, 0)
	for deviceId, v := range userIds {
		if v == "" {
			continue
		}
		user := new(model.User)
		json.Unmarshal(util.S2B(v), user)
		//logger.Logger.Debug("apiFindUserList", zap.Any("userJson", user))
		tmpUid := strconv.FormatInt(int64(user.UserId), 10)
		if shopId == tmpUid {
			continue // 不要展示商户自己
		}
		// 获取消息已读偏移
		index, _ := svc.GetMessageAckMapping(deviceId, user.RoomID) // 获取消息已读偏移

		count, err := svc.GetMessageCount(user.RoomID, index, "+inf") // 拿到偏移去统计未读
		if err != nil {
			logger.Logger.Debug("apiFindUserList", zap.String("desc", "拿到偏移去统计未读"), zap.String("err", err.Error()))
			continue
		}

		lastMessage, err := svc.GetMessageList(user.RoomID, 0, 0) // 取回消息
		if err != nil {
			logger.Logger.Debug("apiFindUserList", zap.String("desc", "取回最后一条消息"), zap.String("err", err.Error()))
			continue
		}

		user.Unread = model.Int64(count)
		user.LastMessage = lastMessage

		user.IsOnline = svc.IsOnline(deviceId)
		// 在线的用户先暂存起来
		if user.IsOnline {
			onlineTmp = append(onlineTmp, user)
			continue
		}

		offlineTmp = append(offlineTmp, user)
	}

	onlineTmp = append(onlineTmp, offlineTmp...) //合并离线与在线
	OutPageJson(w, onlineTmp, page)
}
