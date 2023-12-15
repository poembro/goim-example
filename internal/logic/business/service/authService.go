package service

import (
	"context"
	"encoding/json"
	"fmt"
	"goim-example/internal/logic/business/model"
	"goim-example/internal/logic/business/util"
	"strconv"
	"time"

	log "github.com/golang/glog"
)

// AuthLogin 直接拿到客户端传来的明文 token
func (s *Service) AuthLogin(ctx context.Context, server, cookie string, token []byte) (model.Auth, error) {
	if cookie != "" { // 如果有cookie 则表示 有其他额外用户体系 解析cookie即可
		return s.AuthLoginCookie(ctx, server, cookie, token)
	}
	var req model.Auth
	if err := json.Unmarshal(token, &req); err != nil {
		log.Errorf("json.Unmarshal(%s) error(%v)", token, err)
		return req, err
	}

	//解析body  得到 deviceId, userId
	midStr := strconv.FormatInt(req.Mid, 10)
	deviceId := s.BuildDeviceId(req.Platform, midStr)
	if req.Key != deviceId {
		// return req, fmt.Errorf(midStr + ":" + deviceId + " 结果是:" + req.Key + " 登录认证错误,设备编号对不上!")
	}

	// 标记用户上线 并 存储用户信息
	s.dao.AddMapping(req.Mid, deviceId, server, string(token))
	// 将用户归属到指定商户
	s.dao.AddUserByShop(req.RoomID, midStr)
	return req, nil
}

// AuthLoginCookie 直接拿到客户端传来的明文 token
func (s *Service) AuthLoginCookie(ctx context.Context, server, cookie string, token []byte) (model.Auth, error) {
	var req model.Auth
	if err := json.Unmarshal(token, &req); err != nil {
		log.Errorf("json.Unmarshal(%s) error(%v)", token, err)
		return req, err
	}

	//解析body  得到 deviceId, userId
	midStr := strconv.FormatInt(req.Mid, 10)
	deviceId := s.BuildDeviceId(req.Platform, midStr)
	if req.Key != deviceId {
		return req, fmt.Errorf(midStr + "应该是:" + deviceId + " 结果是:" + req.Key + " 登录认证错误,设备编号对不上!")
	}

	// 标记用户上线 并 存储用户信息
	s.dao.AddMapping(req.Mid, deviceId, server, string(token))
	// 将用户归属到指定商户
	s.dao.AddUserByShop("8000", midStr)
	return req, nil
}

// BuildDeviceId 构建 DeviceId
func (*Service) BuildDeviceId(platform string, userId string) string {
	return util.Md5(fmt.Sprintf("%s_%s", platform, userId))
}

// BuildDeviceId 构建 DeviceId
func (*Service) BuildMid() (uint64, string) {
	sID := util.SFlake.GetID()

	return sID, strconv.FormatUint(sID, 10)
}

func (s *Service) CreateUser(shopId, shopName, shopFace, remoteAddr, referer, userAgent string) *model.User {
	platform := "web"
	Mid, smid := s.BuildMid()
	deviceId := s.BuildDeviceId(platform, smid)

	l := len(smid)
	nickname := fmt.Sprintf("user%s", smid[l-6:l])

	token, err := util.GetToken(smid, deviceId, nickname)
	if err != nil {
		panic(err)
	}

	return &model.User{
		Auth: model.Auth{
			Mid:      int64(Mid),
			Key:      deviceId,
			RoomID:   fmt.Sprintf("%s://%s", model.RoomTyp, smid),
			Platform: platform,
			// 8000是频道 如: 客服聊天类型 弹幕信息类型 与某人聊天房间类型
			// 游戏大厅的通知 游戏匹配成功的通知 游戏房间的聊天
			// 如 在游戏房间刷怪、聊天的同时, 还能接收游戏大厅的广播通知消息)
			Accepts: []int32{model.OpGlobal, model.OpMessage},
		},
		Nickname:   nickname,
		Face:       "http://img.touxiangwu.com/2020/3/uq6Bja.jpg", // 随机头像
		ShopId:     shopId,
		ShopName:   shopName,
		ShopFace:   shopFace,
		Suburl:     "ws://192.168.84.168:3102/sub",            // 订阅地址
		Pushurl:    "http://192.168.84.168:3111/api/msg/push", // 发布地址
		RemoteAddr: remoteAddr,
		Referer:    referer,
		UserAgent:  userAgent,
		CreatedAt:  util.FormatTime(time.Now()),
		Token:      token,
	}
}

func (s *Service) ShopCreate(shopId, shopName, shopFace, remoteAddr, referer, userAgent string) *model.User {
	platform := "web"
	userId := shopId
	sID, _ := strconv.ParseInt(shopId, 10, 64)

	deviceId := s.BuildDeviceId(platform, userId)
	//l := len(userId)   userId[l-6:l]
	nickname := fmt.Sprintf("user%s", userId)

	token, err := util.GetToken(userId, deviceId, nickname)
	if err != nil {
		panic(err)
	}
	return &model.User{
		Auth: model.Auth{
			Mid:      int64(sID),
			Key:      deviceId,
			RoomID:   fmt.Sprintf("%s://%s", model.RoomTyp, shopId),
			Platform: platform,
			// 8000是频道 如: 客服聊天类型 弹幕信息类型 与某人聊天房间类型
			// 游戏大厅的通知 游戏匹配成功的通知 游戏房间的聊天
			// 如 在游戏房间刷怪、聊天的同时, 还能接收游戏大厅的广播通知消息)
			Accepts: []int32{model.OpGlobal, model.OpMessage},
		},

		Nickname:   nickname,
		Face:       "http://img.touxiangwu.com/2020/3/uq6Bja.jpg", // 随机头像
		ShopId:     shopId,
		ShopName:   shopName,
		ShopFace:   shopFace,
		Suburl:     "ws://192.168.84.168:3102/sub",                                  // 订阅地址
		Pushurl:    "http://192.168.84.168:3111/api/msg/push?&platform=" + platform, // 发布地址
		RemoteAddr: remoteAddr,
		Referer:    referer,
		UserAgent:  userAgent,
		CreatedAt:  util.FormatTime(time.Now()),
		Token:      token,
	}
}
