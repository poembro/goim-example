package apihttp

import (
	"context"

	"github.com/gin-gonic/gin"
)

func (s *Router) OnlineTop(c *gin.Context) {
	var arg struct {
		Type  string `form:"type" binding:"required"`
		Limit int    `form:"limit" binding:"required"`
	}
	if err := c.BindQuery(&arg); err != nil {
		s.OutJson(c, -1, err.Error(), nil)
		return
	}
	res, err := s.logic.OnlineTop(c, arg.Type, arg.Limit)
	if err != nil {
		s.OutJson(c, -1, err.Error(), nil)
		return
	}
	s.OutJson(c, 200, "success", res)
}

func (s *Router) OnlineRoom(c *gin.Context) {
	var arg struct {
		Type  string   `form:"type" binding:"required"`
		Rooms []string `form:"rooms" binding:"required"`
	}
	if err := c.BindQuery(&arg); err != nil {
		s.OutJson(c, -1, err.Error(), nil)
		return
	}
	res, err := s.logic.OnlineRoom(c, arg.Type, arg.Rooms)
	if err != nil {
		s.OutJson(c, -1, err.Error(), nil)
		return
	}
	s.OutJson(c, 200, "success", res)
}

func (s *Router) OnlineTotal(c *gin.Context) {
	ipCount, connCount := s.logic.OnlineTotal(context.TODO())
	res := map[string]interface{}{
		"ip_count":   ipCount,
		"conn_count": connCount,
	}
	s.OutJson(c, 200, "success", res)
}
