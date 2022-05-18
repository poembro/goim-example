package apihttp

import (
	"net/http"
)

// ListMsg 查历史消息
func (h *Router) ListMsg(w http.ResponseWriter, r *http.Request) {
	var (
		roomId string
	)

	if r.Method == "POST" {
		roomId = r.FormValue("room_id")
	}
	if roomId == "" {
		OutJson(w, OutData{Code: -1, Success: false, Msg: "参数不能为空"})
		return
	}

	page := NewPage(r)
	dst, total, err := svc.GetMessagePageList(roomId, "-inf", "+inf", int64(page.Page), int64(page.Limit))
	if err != nil {
		OutJson(w, OutData{Code: -1, Success: false, Msg: err.Error()})
		return
	}
	page.Total = total

	OutPageJson(w, dst, page)
}
