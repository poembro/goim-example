package business

import (
	"context"
	"fmt"
	"goim-demo/internal/logic/business/model"
	"goim-demo/internal/logic/business/util"
	"strconv"
)



// SignIn 长连接登录
// 方案一: body 是一个jwt token 值 去其他服务拿到对应的 头像昵称等信息
// 方案二: demo 中 body 是一个json 已经包含了头像昵称等信息
func (s *Business) SignIn(ctx context.Context, user *model.User, body []byte, connAddr string) error {
	//解析body  得到 deviceId, userId
	userId = int64(user.Mid)
	uidStr := strconv.FormatInt(userId, 10)
	deviceId = s.BuildDeviceId(user.Platform, uidStr)
	if user.DeviceId != deviceId {
		return fmt.Errorf(uidStr + "应该是:" + deviceId + " 结果是:" + user.DeviceId + " 登录认证错误,设备编号对不上!")
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

func (*Business) buildUserMap(shopId, shopName, shopFace string)  *model.User {
	platform := "web"
	sID, err := util.SFlake.GetID()
	if err != nil { 
		return nil, fmt.Errorf("id错误")
	}

	userId := strconv.FormatUint(sID, 10)
	 
	l := len(userId)
	nickname := fmt.Sprintf("user%s", userId[l-6:l])
	deviceId := svc.BuildDeviceId(platform, userId)
	token, err := util.GetToken(userId, deviceId, nickname)
	if err != nil {
		return nil
	}
    return &model.User{
		Mid : model.Int64(sID),
		Key : deviceId,
		RoomID :"live://8000",
		Platform :platform,
		Accepts :[]int32{100},
		Nickname :nickname,
		Face : "http://img.touxiangwu.com/2020/3/uq6Bja.jpg", // 随机头像
		ShopId:shopId, 
		ShopName:shopName, 
		ShopFace:shopFace, 
		Suburl:      "ws://localhost:7923/ws",                                // 订阅地址
		Pushurl:     "http://localhost:8090/open/push?&platform=" + platform, // 发布地址
		RemoteAddr: "127.0.0.1",
		Referer:     "",
		UserAgent:  "",
		CreatedAt:  util.FormatTime(time.Now()), 
	}
}

func (s *Business) UserCreate(shopId string) (user *model.User, err error){
	//判断客服是否存在
	shop, err := s.GetShop(shopId)
	if shop == nil || err != nil {
		return nil, fmt.Errorf("参数错误")
	}
 
	dst := s.buildUserMap(shop.UserId, shop.Nickname, shop.Face)
    return dst, nil
}
