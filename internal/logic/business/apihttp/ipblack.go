package apihttp

import (
	"context"

	"github.com/gin-gonic/gin"
)

// IpblackAdd ip添加至黑名单
func (s *Router) IpblackAdd(c *gin.Context) {
	var arg struct {
		IP     string `json:"ip"`
		ShopId string `json:"shop_id"`
	}
	if err := c.BindJSON(&arg); err != nil {
		s.OutJson(c, -1, err.Error(), nil)
		return
	}

	s.svc.IpblackAdd(context.TODO(), arg.ShopId, arg.IP)
	s.OutJson(c, 200, "success", nil)
}

// DelIpblack ip从黑名单删除
func (s *Router) IpblackDel(c *gin.Context) {
	var arg struct {
		IP     string `json:"ip"`
		ShopId string `json:"shop_id"`
	}
	if err := c.BindJSON(&arg); err != nil {
		s.OutJson(c, -1, err.Error(), nil)
		return
	}

	s.svc.IpblackDel(context.TODO(), arg.ShopId, arg.IP)
	s.OutJson(c, 200, "success", nil)
}

// IpblackList 查看所有ip
func (s *Router) IpblackList(c *gin.Context) {
	var arg struct {
		ShopId string `json:"shop_id"`
	}
	if err := c.BindJSON(&arg); err != nil {
		s.OutJson(c, -1, err.Error(), nil)
		return
	}

	// 查询在线人数
	page := NewPage(c)
	dst, total, err := s.svc.IpblackList(arg.ShopId, "-inf", "+inf", int64(page.Page), int64(page.Limit))
	if err != nil {
		s.OutJson(c, -1, err.Error(), nil)
		return
	}
	page.Total = total

	s.OutPageJson(c, dst, page)
}
