package apihttp

import (
	"fmt"
	"goim-demo/internal/logic/business/util"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)


func buildUserMap(r *http.Request, userId, shopId, shopName, shopFace string) map[string]string {
	platform := "web"

	// 客服聊天场景
	l := len(userId)
	nickname := fmt.Sprintf("user%s", userId[l-6:l])
	deviceId := svc.BuildDeviceId(platform, userId)
	token, err := util.GetToken(userId, deviceId, nickname)
	if err != nil {
		return nil
	}
	dst := map[string]string{
		"user_id":     userId,                                        // 用户大多是临时过来咨询,所以这里采用随机唯一
		"nickname":    nickname,                                      // 随机昵称
		"face":        "http://img.touxiangwu.com/2020/3/uq6Bja.jpg", // 随机头像
		"device_id":   deviceId,                                      // 多个平台达到的效果不一样
		"room_id":     svc.BuildDeviceId(userId, deviceId),           //房间号唯一否则消息串房间(暂时以用户id为房间号)
		"shop_id":     shopId,                                        // 登录该后台的手机号
		"shop_name":   shopName,                                      // 客服昵称
		"shop_face":   shopFace,                                      // 客服头像
		"platform":    platform,
		"suburl":      "ws://localhost:7923/ws",                                // 订阅地址
		"pushurl":     "http://localhost:8090/open/push?&platform=" + platform, // 发布地址
		"remote_addr": GetAddr(r),
		"referer":     r.Referer(),
		"user_agent":  r.UserAgent(),
		"created_at":  util.FormatTime(time.Now()),
		"token":       token,
	}
	return dst
}

// UserCreate 创建用户
func (s *Router) UserCreate(c *gin.Context) {
	var arg struct {
		ShopId   int32    `form:"shop_id"` 
	}
	if err := c.BindQuery(&arg); err != nil {
		OutJson(c, OutData{Code: -1, Success: false, Msg: err.Error()})
		return
	}
	if arg.ShopId == "" {
		OutJson(c, OutData{Code: -1, Success: false, Msg: "参数不能为空"})
		return
	}
	dst, err := s.svc.UserCreate(arg.ShopId)
	// 客服聊天场景
	OutJson(c, OutData{Code: 200, Success: true, Msg: "success", Result: dst})
}

// apiLogin 登录 (后台)
func (s *Router) apiLogin(c *gin.Context) {
	var (
		nickname string
		password string
	)

	if r.Method == "POST" {
		nickname = r.FormValue("nickname")
		password = r.FormValue("password")
	}

	if nickname == "" || password == "" {
		OutJson(c, OutData{Code: -1, Success: false, Msg: "参数nickname or password不能为空"})
		return
	}

	shop, err := svc.GetShop(nickname)
	if err != nil {
		OutJson(c, OutData{Code: -1, Success: false, Msg: "未注册"})
		return
	}

	if shop.Password != password {
		OutJson(c, OutData{Code: -1, Success: false, Msg: "密码错误"})
		return
	}
	dst := buildUserMap(r, shop.UserId, shop.UserId, shop.Nickname, shop.Face)
	OutJson(c, OutData{Code: 200, Success: true, Result: dst})
}

// apiRegister 注册 (后台) 为了演示,临时采用redis存储
func (s *Router) apiRegister(c *gin.Context) {
	var (
		nickname string
		password string
	)

	if r.Method == "POST" {
		nickname = r.FormValue("nickname")
		password = r.FormValue("password")
	}
	if nickname == "" || password == "" {
		OutJson(c, OutData{Code: -1, Success: false, Msg: "参数nickname or password不能为空"})
		return
	}

	sID, err := util.SFlake.GetID()
	if err != nil {
		OutJson(c, OutData{Code: -1, Success: false, Msg: err.Error()})
		return
	}

	face := "https://img.wxcha.com/m00/86/59/7c6242363084072b82b6957cacc335c7.jpg"
	svc.AddShop(strconv.FormatUint(sID, 10), nickname, face, password)

	OutJson(c, OutData{Code: 200, Success: true, Msg: "success", Result: "xxx"})
	return
}
