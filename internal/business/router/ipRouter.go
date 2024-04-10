package router

import (
	"github.com/gin-gonic/gin"
)

// IpCreate ip添加至黑名单
func (s *Router) IpCreate(c *gin.Context) {
	var arg struct {
		IP     string `json:"ip"`
		ShopId string `json:"shop_id"`
	}
	if err := c.BindJSON(&arg); err != nil {
		s.OutJson(c, -1, err.Error(), nil)
		return
	}

	s.svc.IpCreate(c.Request.Context(), arg.ShopId, arg.IP)
	s.OutJson(c, 200, "success", nil)
}

// DelIp ip从黑名单删除
func (s *Router) IpRemove(c *gin.Context) {
	var arg struct {
		IP     string `json:"ip"`
		ShopId string `json:"shop_id"`
	}
	if err := c.BindJSON(&arg); err != nil {
		s.OutJson(c, -1, err.Error(), nil)
		return
	}

	s.svc.IpRemove(c.Request.Context(), arg.ShopId, arg.IP)
	s.OutJson(c, 200, "success", nil)
}

// IpList 查看所有ip
func (s *Router) IpList(c *gin.Context) {
	var arg struct {
		ShopId string `json:"shop_id"`
	}
	if err := c.BindJSON(&arg); err != nil {
		s.OutJson(c, -1, err.Error(), nil)
		return
	}

	// 查询在线人数
	page := NewPage(c)
	dst, total, err := s.svc.IpList(c.Request.Context(), arg.ShopId, "-inf", "+inf", int64(page.Page), int64(page.Limit))
	if err != nil {
		s.OutJson(c, -1, err.Error(), nil)
		return
	}
	page.Total = total

	s.OutPageJson(c, dst, page)
}
