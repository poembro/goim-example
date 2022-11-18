package business

import (
	"context"
	"fmt"
	"goim-demo/internal/logic/business/model"
	"goim-demo/internal/logic/business/util"
	"strconv"
	"time"
)

// SignIn 长连接登录 (comet服务通过grpc发来的body参数)
// 方案一: body 是一个jwt token 值 去其他服务拿到对应的 头像昵称等信息
// 方案二: demo 中 body 是一个json 已经包含了头像昵称等信息
func (s *Business) SignIn(ctx context.Context, user *model.User, body []byte, connAddr string) error {
	//解析body  得到 deviceId, userId
	userId := int64(user.Mid)
	uidStr := strconv.FormatInt(userId, 10)
	deviceId := s.BuildDeviceId(user.Platform, uidStr)
	if user.Key != deviceId {
		return fmt.Errorf(uidStr + "应该是:" + deviceId + " 结果是:" + user.Key + " 登录认证错误,设备编号对不上!")
	}

	// 标记用户上线 并 存储用户信息
	s.dao.AddMapping(userId, deviceId, connAddr, string(body))
	//将用户归属到指定商户
	s.dao.AddUserByShop(user.ShopId, uidStr)
	return nil
}

// BuildDeviceId 构建 DeviceId
func (*Business) BuildDeviceId(platform string, userId string) string {
	return util.Md5(fmt.Sprintf("%s_%s", platform, userId))
}

// BuildDeviceId 构建 DeviceId
func (*Business) BuildMid() (uint64, string) {
	sID, err := util.SFlake.GetID()
	if err != nil {
		return 0, ""
	}

	return sID, strconv.FormatUint(sID, 10)
}

func (s *Business) CreateUser(shopId, shopName, shopFace, remoteAddr, referer, userAgent string) *model.User {
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
		Mid:      model.Int64(Mid),
		Key:      deviceId,
		RoomID:   fmt.Sprintf("%s://%s", model.RoomTyp, smid),
		Platform: platform,
		// 8000是频道 如: 客服聊天类型 弹幕信息类型 与某人聊天房间类型
		// 游戏大厅的通知 游戏匹配成功的通知 游戏房间的聊天
		// 如 在游戏房间刷怪、聊天的同时, 还能接收游戏大厅的广播通知消息)
		Accepts:    []int32{model.OpGlobal, model.OpMessage},
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

func (s *Business) ShopCreate(shopId, shopName, shopFace, remoteAddr, referer, userAgent string) *model.User {
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
		Mid:        model.Int64(sID),
		Key:        deviceId,
		RoomID:     fmt.Sprintf("%s://%s", model.RoomTyp, shopId),
		Platform:   platform,
		Accepts:    []int32{model.OpGlobal, model.OpMessage},
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
