package apihttp

import "github.com/gin-gonic/gin"

// MsgList 查历史消息
func (s *Router) MsgList(c *gin.Context) {
	var arg struct {
		RoomId string `form:"room_id"`
	}
	if err := c.BindQuery(&arg); err != nil {
		OutJson(c, OutData{Code: -1, Success: false, Msg: err.Error()})
		return
	}

	page := NewPage(c)
	dst, total, err := s.svc.GetMessagePageList(arg.RoomId, "-inf", "+inf", int64(page.Page), int64(page.Limit))
	if err != nil {
		OutJson(c, OutData{Code: -1, Success: false, Msg: err.Error()})
		return
	}
	page.Total = total

	OutPageJson(c, dst, page)
}
