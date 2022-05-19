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
		OutJson(c, OutData{Code: -1, Success: false, Msg: err.Error()})
		return
	}

	s.svc.AddIpblack(context.TODO(), arg.ShopId, arg.IP)
	OutJson(c, OutData{Code: 200, Success: true, Result: nil})
}

// IpblackDel ip从黑名单删除
func (s *Router) IpblackDel(c *gin.Context) {
	var arg struct {
		IP     string `json:"ip"`
		ShopId string `json:"shop_id"`
	}
	if err := c.BindJSON(&arg); err != nil {
		OutJson(c, OutData{Code: -1, Success: false, Msg: err.Error()})
		return
	}

	s.svc.DelIpblack(context.TODO(), arg.ShopId, arg.IP)
	OutJson(c, OutData{Code: 200, Success: true, Result: nil})
}

// listIpblack 查看所有ip
func (s *Router) IpblackList(c *gin.Context) {
	var arg struct {
		ShopId string `json:"shop_id"`
	}
	if err := c.BindJSON(&arg); err != nil {
		OutJson(c, OutData{Code: -1, Success: false, Msg: err.Error()})
		return
	}

	// 查询在线人数
	page := NewPage(c)
	dst, total, err := s.svc.ListIpblack(arg.ShopId, "-inf", "+inf", int64(page.Page), int64(page.Limit))
	if err != nil {
		OutJson(c, OutData{Code: -1, Success: false, Msg: err.Error()})
		return
	}
	page.Total = total

	OutPageJson(c, dst, page)
}
