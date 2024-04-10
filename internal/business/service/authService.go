package service

import (
	"context"
	"encoding/json"
	"fmt"
	"goim-example/internal/business/model"
	"goim-example/internal/business/util"
	utilModel "goim-example/internal/logic/model"
	"strconv"
	"time"

	log "github.com/golang/glog"
)

// AuthLogin 直接拿到客户端传来的明文 token
func (s *Service) AuthLogin(ctx context.Context, server, cookie string, token []byte) (model.Auth, error) {
	if cookie != "" { // 如果有cookie 则表示 有其他额外用户体系 解析cookie即可
		// return s.AuthLoginCookie(ctx, server, cookie, token)
	}
	var req model.Auth
	if err := json.Unmarshal(token, &req); err != nil {
		log.Errorf("json.Unmarshal(%s) error(%v)", token, err)
		return req, err
	}

	// 按 roomid 字段规则 拆分出 商户id 和 用户id
	shopId, midStr, _ := utilModel.DecodeRoomKey(req.RoomID)
	log.Errorf("---> req.RoomID (%s)  ::: shopId(%s) midStr(%s) ", req.RoomID, shopId, midStr)

	deviceId := s.BuildDeviceId(req.Platform, midStr)
	if req.Key != deviceId {
		return req, fmt.Errorf(midStr + ":" + deviceId + " 结果是:" + req.Key + " 登录认证错误,设备编号对不上!")
	}

	// 标记用户上线 并 存储用户信息
	s.dao.UserCreate(ctx, midStr, deviceId, server, string(token))

	// 将用户归属到指定商户
	s.ShopAppendUserId(ctx, shopId, midStr)
	return req, nil
}

// BuildDeviceId 构建 DeviceId
func (*Service) BuildDeviceId(platform string, userId string) string {
	key := fmt.Sprintf("%s_%s", platform, userId)

	body := util.Md5(key)

	log.Errorf("加密前 (%s)  加密后(%s)", key, body)

	return body
}

func (s *Service) UserCreate(shop *model.Shop, remoteAddr, referer, userAgent string) *model.User {
	platform := "web"
	var Mid int64
	var smid string
	if remoteAddr == "0.0.0.0" {
		Mid, _ = strconv.ParseInt(shop.Mid, 10, 64)
		smid = shop.Mid
	} else {
		Mid = util.SFlake.GetID()
		smid = strconv.FormatInt(Mid, 10)
	}

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
			RoomID:   utilModel.EncodeRoomKey(shop.Nickname, smid),
			Platform: platform,
			Accepts:  []int32{model.OpGlobal, model.OpMessage},
		},
		Nickname:   nickname,
		Face:       "http://img.touxiangwu.com/2020/3/uq6Bja.jpg", // 随机头像
		ShopId:     shop.Mid,
		ShopName:   shop.Nickname,
		ShopFace:   shop.Face,
		Suburl:     "ws://192.168.84.168:3102/sub", // 订阅地址
		Pushurl:    "/api/msg/push",                // 发布地址
		RemoteAddr: remoteAddr,
		Referer:    referer,
		UserAgent:  userAgent,
		CreatedAt:  util.FormatTime(time.Now()),
		Token:      token,
	}
}

// UserFinds 通过userId 获取用户信息
func (s *Service) UserFinds(ctx context.Context, userIds []string) (map[string]string, error) {
	return s.dao.UserFinds(ctx, userIds)
}
