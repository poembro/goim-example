package apihttp

import (
	"net/http"
)

// addIpblack ip添加至黑名单
func (s *Router) addIpblack(c *gin.Context) {
	var (
		shopId string
		ip     string
	)
	if r.Method == "POST" {
		ip = r.FormValue("ip")
		shopId = r.FormValue("shop_id")
	}
	if ip == "" || shopId == "" {
		OutJson(c, OutData{Code: -1, Success: false, Msg: "参数ip不能为空"})
		return
	}
	svc.AddIpblack(r.Context(), shopId, ip)
	OutJson(c, OutData{Code: 200, Success: true, Result: nil})
}

// addIpblack ip从黑名单删除
func (s *Router) delIpblack(c *gin.Context) {
	var (
		shopId string
		ip     string
	)
	if r.Method == "POST" {
		ip = r.FormValue("ip")
		shopId = r.FormValue("shop_id")
	}
	if ip == "" || shopId == "" {
		OutJson(c, OutData{Code: -1, Success: false, Msg: "参数ip不能为空"})
		return
	}
	svc.DelIpblack(r.Context(), shopId, ip)
	OutJson(c, OutData{Code: 200, Success: true, Result: nil})
}

// listIpblack 查看所有ip
func (s *Router) listIpblack(c *gin.Context) {
	var (
		shopId string
	)

	if r.Method == "POST" {
		shopId = r.FormValue("shop_id")
	}
	if shopId == "" {
		OutJson(c, OutData{Code: -1, Success: false, Msg: "参数不能为空"})
		return
	}

	// 查询在线人数
	page := NewPage(r)
	dst, total, err := svc.ListIpblack(shopId, "-inf", "+inf", int64(page.Page), int64(page.Limit))
	if err != nil {
		OutJson(c, OutData{Code: -1, Success: false, Msg: err.Error()})
		return
	}
	page.Total = total

	OutPageJson(w, dst, page)
}
