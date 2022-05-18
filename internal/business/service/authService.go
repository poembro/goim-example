package service

import (
	"context"
	"encoding/json"
	"fmt"
	"goim-demo/internal/logic/model"
	"goim-demo/pkg/util"
	"strconv"
)

// SignIn 长连接登录
// 方案一: body 是一个jwt token 值 去其他服务拿到对应的 头像昵称等信息
// 方案二: demo 中 body 是一个json 已经包含了头像昵称等信息
func (s *Service) SignIn(ctx context.Context, body []byte, connAddr string, clientAddr string) (string, int64, error) {
	var (
		user     model.User
		deviceId string
		userId   = int64(0)
	)
	//解析body  得到 deviceId, userId
	if err := json.Unmarshal(body, &user); err != nil {
		return deviceId, userId, fmt.Errorf("json.Unmarshal expected ")
	}
	userId = int64(user.UserId)
	uidStr := strconv.FormatInt(userId, 10)
	deviceId = s.BuildDeviceId(user.Platform, uidStr)
	if user.DeviceId != deviceId {
		return deviceId, userId, fmt.Errorf(uidStr + "应该是:" + deviceId + " 结果是:" + user.DeviceId + " 登录认证错误,设备编号对不上!")
	}

	// 标记用户上线 并 存储用户信息
	s.dao.AddMapping(userId, deviceId, connAddr, string(body))
	//将用户归属到指定商户
	s.dao.AddUserByShop(user.ShopId, uidStr)
	return deviceId, userId, nil
}

// BuildDeviceId 构建 DeviceId
func (*Service) BuildDeviceId(platform string, userId string) string {
	return util.Md5(fmt.Sprintf("%s_%s", platform, userId))
}
