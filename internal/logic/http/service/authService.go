package service

import (
	"context"
	"encoding/json"
	"fmt"
	"goim-example/internal/logic/http/model"
	"goim-example/internal/logic/http/util"
	utilModel "goim-example/internal/logic/model"
	"strconv"
	"time"

	log "github.com/golang/glog"
)

func (s *Service) AuthLogin(ctx context.Context, server, cookie string, token []byte) (*model.Auth, error) {
	if cookie != "" { // 如果有cookie 则表示 有其他额外用户体系 解析cookie即可
		item, err := util.DecryptToken(cookie)
		if err != nil {
			return nil, fmt.Errorf("请重新授权登录")
		}

		return &model.Auth{
			Mid:     item.Mid,
			Key:     item.Key,
			RoomID:  utilModel.EncodeRoomKey(model.OpType, "1001"),
			Accepts: []int32{model.OpGlobal, model.OpMessage},
		}, nil
	}

	// AuthLogin 直接拿到客户端传来的明文 token  {"mid":123, "room_id":"live://1000", "platform":"web", "accepts":[1000,1001,1002]}
	var req model.Auth
	if err := json.Unmarshal(token, &req); err != nil {
		log.Errorf("json.Unmarshal(%s) error(%v)", token, err)
		return &req, err
	}

	_, shopId, _ := utilModel.DecodeRoomKey(req.RoomID)
	log.Errorf("---> req.RoomID (%s)  ::: shopId(%s) ", req.RoomID, shopId)

	// 标记用户上线 并 存储用户信息
	midStr := strconv.FormatInt(req.Mid, 10)
	s.dao.UserCreate(ctx, midStr, req.Key, server, string(token))

	// 将用户归属到指定商户
	s.ShopAppendUserId(ctx, shopId, midStr)
	return &req, nil
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

	midStr := strconv.FormatInt(Mid, 10)

	l := len(smid)
	nickname := fmt.Sprintf("user%s", smid[l-6:l])
	deviceId := fmt.Sprintf("%s:%s", platform, midStr)
	token, err := util.GetToken(Mid, deviceId, nickname)
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
