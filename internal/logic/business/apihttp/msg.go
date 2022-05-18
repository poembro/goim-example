package apihttp

import (
	"net/http"
)

// ListMsg 查历史消息
func (s *Router) ListMsg(c *gin.Context) {
	var (
		roomId string
	)

	if r.Method == "POST" {
		roomId = r.FormValue("room_id")
	}
	if roomId == "" {
		OutJson(c, OutData{Code: -1, Success: false, Msg: "参数不能为空"})
		return
	}

	page := NewPage(r)
	dst, total, err := svc.GetMessagePageList(roomId, "-inf", "+inf", int64(page.Page), int64(page.Limit))
	if err != nil {
		OutJson(c, OutData{Code: -1, Success: false, Msg: err.Error()})
		return
	}
	page.Total = total

	OutPageJson(w, dst, page)
}
