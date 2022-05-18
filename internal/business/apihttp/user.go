package apihttp

import (
	"fmt"
	"goim-demo/pkg/util"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func GetAddr(r *http.Request) string {
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	ip := strings.TrimSpace(strings.Split(xForwardedFor, ",")[0])
	if ip != "" {
		return ip
	}

	ip = strings.TrimSpace(r.Header.Get("X-Real-Ip"))
	if ip != "" {
		return ip
	}

	if ip, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr)); err == nil {
		return ip
	}

	return ""
}

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

// apiAddUser 创建客服
func (h *Router) apiAddUser(w http.ResponseWriter, r *http.Request) {
	var shopId string
	if r.Method == "POST" {
		shopId = r.FormValue("shop_id")
	}
	if shopId == "" {
		OutJson(w, OutData{Code: -1, Success: false, Msg: "参数不能为空"})
		return
	}
	//判断客服是否存在
	shop, err := svc.GetShop(shopId)
	if shop == nil || err != nil {
		OutJson(w, OutData{Code: -1, Success: false, Msg: "参数错误"})
		return
	}

	sID, err := util.SFlake.GetID()
	if err != nil {
		OutJson(w, OutData{Code: -1, Success: false, Msg: err.Error()})
		return
	}

	dst := buildUserMap(r, strconv.FormatUint(sID, 10), shop.UserId, shop.Nickname, shop.Face)
	// 客服聊天场景
	OutJson(w, OutData{Code: 200, Success: true, Msg: "success", Result: dst})
}

// apiLogin 登录 (后台)
func (h *Router) apiLogin(w http.ResponseWriter, r *http.Request) {
	var (
		nickname string
		password string
	)

	if r.Method == "POST" {
		nickname = r.FormValue("nickname")
		password = r.FormValue("password")
	}

	if nickname == "" || password == "" {
		OutJson(w, OutData{Code: -1, Success: false, Msg: "参数nickname or password不能为空"})
		return
	}

	shop, err := svc.GetShop(nickname)
	if err != nil {
		OutJson(w, OutData{Code: -1, Success: false, Msg: "未注册"})
		return
	}

	if shop.Password != password {
		OutJson(w, OutData{Code: -1, Success: false, Msg: "密码错误"})
		return
	}
	dst := buildUserMap(r, shop.UserId, shop.UserId, shop.Nickname, shop.Face)
	OutJson(w, OutData{Code: 200, Success: true, Result: dst})
}

// apiRegister 注册 (后台) 为了演示,临时采用redis存储
func (h *Router) apiRegister(w http.ResponseWriter, r *http.Request) {
	var (
		nickname string
		password string
	)

	if r.Method == "POST" {
		nickname = r.FormValue("nickname")
		password = r.FormValue("password")
	}
	if nickname == "" || password == "" {
		OutJson(w, OutData{Code: -1, Success: false, Msg: "参数nickname or password不能为空"})
		return
	}

	sID, err := util.SFlake.GetID()
	if err != nil {
		OutJson(w, OutData{Code: -1, Success: false, Msg: err.Error()})
		return
	}

	face := "https://img.wxcha.com/m00/86/59/7c6242363084072b82b6957cacc335c7.jpg"
	svc.AddShop(strconv.FormatUint(sID, 10), nickname, face, password)

	OutJson(w, OutData{Code: 200, Success: true, Msg: "success", Result: "xxx"})
	return
}
