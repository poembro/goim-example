package business

import (
	"goim-example/internal/logic"
	"goim-example/internal/logic/business/apihttp"
	"goim-example/internal/logic/business/service"
	"goim-example/internal/logic/conf"
	"net/http"

	"github.com/gin-gonic/gin"
)

func New(c *conf.Config, l *logic.Logic) *service.Service {
	// 初始化
	s := service.New(c)
	r := apihttp.New(c, l, s)

	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(r.LoggerHandler, r.RecoverHandler)

	// 消息推送模块
	group := engine.Group("/goim")
	group.POST("/push/keys", r.PushKeys)
	group.POST("/push/mids", r.PushMids)
	group.POST("/push/room", r.PushRoom)
	group.POST("/push/all", r.PushAll)
	group.GET("/online/top", r.OnlineTop)
	group.GET("/online/room", r.OnlineRoom)
	group.GET("/online/total", r.OnlineTotal)

	engine.StaticFS("/_/", http.Dir("./examples/javascript/"))

	// 业务模块
	engine.Use(r.CorsMiddleware)
	group2 := engine.Group("/api")
	{
		group2.POST("/user/login", r.Login)
		group2.POST("/user/register", r.Register)

		group2.GET("/user/create", r.CreateUser)
		authorized := group2.Group("")
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
	go func() {
		if err := engine.Run(c.HTTPServer.Addr); err != nil {
			panic(err)
		}
	}()

	return s
}
