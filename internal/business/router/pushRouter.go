package router

import (
	"context"
	"io"

	"github.com/gin-gonic/gin"
)

func (s *Router) PushKeys(c *gin.Context) {
	var arg struct {
		Op   int32    `form:"operation"`
		Keys []string `form:"keys"`
	}
	if err := c.BindQuery(&arg); err != nil {
		s.OutJson(c, -1, err.Error(), nil)
		return
	}
	// read message
	msg, err := io.ReadAll(c.Request.Body)
	if err != nil {
		s.OutJson(c, -1, err.Error(), nil)
		return
	}
	if err = s.logic.PushKeys(context.TODO(), arg.Op, arg.Keys, msg); err != nil {
		s.OutJson(c, -1, err.Error(), nil)
		return
	}
	s.OutJson(c, 200, "success", nil)
}

func (s *Router) PushMids(c *gin.Context) {
	var arg struct {
		Op   int32   `form:"operation"`
		Mids []int64 `form:"mids"`
	}
	if err := c.BindQuery(&arg); err != nil {
		s.OutJson(c, -1, err.Error(), nil)
		return
	}
	// read message
	msg, err := io.ReadAll(c.Request.Body)
	if err != nil {
		s.OutJson(c, -1, err.Error(), nil)
		return
	}
	if err = s.logic.PushMids(context.TODO(), arg.Op, arg.Mids, msg); err != nil {
		s.OutJson(c, -1, err.Error(), nil)
		return
	}
	s.OutJson(c, 200, "success", nil)
}

func (s *Router) PushRoom(c *gin.Context) {
	var arg struct {
		Op   int32  `form:"operation" binding:"required"`
		Type string `form:"type" binding:"required"`
		Room string `form:"room" binding:"required"`
	}
	if err := c.BindQuery(&arg); err != nil {
		s.OutJson(c, -1, err.Error(), nil)
		return
	}
	// read message
	msg, err := io.ReadAll(c.Request.Body)
	if err != nil {
		s.OutJson(c, -1, err.Error(), nil)
		return
	}

	if err = s.logic.PushRoom(c, arg.Op, arg.Type, arg.Room, msg); err != nil {
		s.OutJson(c, -1, err.Error(), nil)
		return
	}
	s.OutJson(c, 200, "success", nil)
}

func (s *Router) PushAll(c *gin.Context) {
	var arg struct {
		Op    int32 `form:"operation" binding:"required"`
		Speed int32 `form:"speed"`
	}
	if err := c.BindQuery(&arg); err != nil {
		s.OutJson(c, -1, err.Error(), nil)
		return
	}
	msg, err := io.ReadAll(c.Request.Body)
	if err != nil {
		s.OutJson(c, -1, err.Error(), nil)
		return
	}
	if err = s.logic.PushAll(c, arg.Op, arg.Speed, msg); err != nil {
		s.OutJson(c, -1, err.Error(), nil)
		return
	}
	s.OutJson(c, 200, "success", nil)
}
