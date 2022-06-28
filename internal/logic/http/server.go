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
	r := apihttp.New(s.logic, s.logic.Business)
	s.engine.Use(r.CorsMiddleware)
	group := s.engine.Group("/api")
	{
		group.POST("/user/login", r.Login)
		group.POST("/user/register", r.Register)
		group.GET("/user/create", r.CreateUser)
		authorized := group.Group("")
		authorized.Use(r.VerifyMiddleware)
		{
			authorized.POST("/user/list", r.ListUser)

			authorized.POST("/msg/push", r.PushMsg)
			authorized.POST("/msg/list", r.ListMsg)
			authorized.POST("/msg/clear", r.ClearMsg)

			authorized.POST("/ipblack/add", r.AddIpblack)
			authorized.POST("/ipblack/del", r.DelIpblack)
			authorized.POST("/ipblack/list", r.ListIpblack)
		}
	}
}

// Close close the server.
func (s *Server) Close() {

}
