package router

import (

	//"goim-example/pkg/logger"

	"github.com/gin-gonic/gin"
)

// UserCreate 创建用户
func (s *Router) UserCreate(c *gin.Context) {
	var arg struct {
		ShopId string `form:"shop_id"`
	}
	if err := c.BindQuery(&arg); err != nil {
		s.OutJson(c, -1, err.Error(), nil)
		return
	}
	if arg.ShopId == "" {
		s.OutJson(c, -1, "参数不能为空", nil)
		return
	}
	//判断客服是否存在
	shop, err := s.svc.ShopFindOne(arg.ShopId) // ShopId 就是商户昵称
	if err != nil || shop == nil || shop.Mid == "" {
		s.OutJson(c, -1, "参数错误", nil)
		return
	}
	item := s.svc.UserCreate(shop, c.ClientIP(), c.GetHeader("referer"), c.GetHeader("user-agent"))

	// 客服聊天场景
	s.OutJson(c, 200, "success", item)
}
