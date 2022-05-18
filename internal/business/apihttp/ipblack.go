package apihttp

import (
	"net/http"
)

// addIpblack ip添加至黑名单
func (h *Router) addIpblack(w http.ResponseWriter, r *http.Request) {
	var (
		shopId string
		ip     string
	)
	if r.Method == "POST" {
		ip = r.FormValue("ip")
		shopId = r.FormValue("shop_id")
	}
	if ip == "" || shopId == "" {
		OutJson(w, OutData{Code: -1, Success: false, Msg: "参数ip不能为空"})
		return
	}
	svc.AddIpblack(r.Context(), shopId, ip)
	OutJson(w, OutData{Code: 200, Success: true, Result: nil})
}

// addIpblack ip从黑名单删除
func (h *Router) delIpblack(w http.ResponseWriter, r *http.Request) {
	var (
		shopId string
		ip     string
	)
	if r.Method == "POST" {
		ip = r.FormValue("ip")
		shopId = r.FormValue("shop_id")
	}
	if ip == "" || shopId == "" {
		OutJson(w, OutData{Code: -1, Success: false, Msg: "参数ip不能为空"})
		return
	}
	svc.DelIpblack(r.Context(), shopId, ip)
	OutJson(w, OutData{Code: 200, Success: true, Result: nil})
}

// listIpblack 查看所有ip
func (h *Router) listIpblack(w http.ResponseWriter, r *http.Request) {
	var (
		shopId string
	)

	if r.Method == "POST" {
		shopId = r.FormValue("shop_id")
	}
	if shopId == "" {
		OutJson(w, OutData{Code: -1, Success: false, Msg: "参数不能为空"})
		return
	}

	// 查询在线人数
	page := NewPage(r)
	dst, total, err := svc.ListIpblack(shopId, "-inf", "+inf", int64(page.Page), int64(page.Limit))
	if err != nil {
		OutJson(w, OutData{Code: -1, Success: false, Msg: err.Error()})
		return
	}
	page.Total = total

	OutPageJson(w, dst, page)
}
