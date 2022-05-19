package http

import (
	"goim-demo/internal/logic"
	"goim-demo/internal/logic/business/apihttp"
	"goim-demo/internal/logic/conf"

	"github.com/gin-gonic/gin"
)

// Server is http server.
type Server struct {
	engine *gin.Engine
	logic  *logic.Logic
}

// New new a http server.
func New(c *conf.HTTPServer, l *logic.Logic) *Server {
	engine := gin.New()
	engine.Use(loggerHandler, recoverHandler)
	go func() {
		if err := engine.Run(c.Addr); err != nil {
			panic(err)
		}
	}()
	s := &Server{
		engine: engine,
		logic:  l,
	}
	s.initRouter()
	s.initBusinessRouter() // 第三方业务
	return s
}

func (s *Server) initRouter() {
	group := s.engine.Group("/goim")
	group.POST("/push/keys", s.pushKeys)
	group.POST("/push/mids", s.pushMids)
	group.POST("/push/room", s.pushRoom)
	group.POST("/push/all", s.pushAll)
	group.GET("/online/top", s.onlineTop)
	group.GET("/online/room", s.onlineRoom)
	group.GET("/online/total", s.onlineTotal)
}

// initBusinessRouter 第三方业务
func (s *Server) initBusinessRouter() {
	r := apihttp.New(s.logic.Business)
	group := s.engine.Group("/api")
	{
		group.POST("/user/login", r.Login)
		group.POST("/user/register", r.Register)
		group.GET("/user/create", r.UserCreate)
		authorized := group.Group("")
		authorized.Use(r.CorsMiddleware, r.VerifyMiddleware)
		{
			authorized.POST("/user/list", r.UserList)

			authorized.POST("/msg/push", r.MsgPush)
			authorized.POST("/msg/list", r.MsgList)
			authorized.POST("/msg/clear", r.MsgClear)

			authorized.POST("/ipblack/add", r.IpblackAdd)
			authorized.POST("/ipblack/del", r.IpblackDel)
			authorized.POST("/ipblack/list", r.IpblackList)
		}
	}
}

// Close close the server.
func (s *Server) Close() {

}
