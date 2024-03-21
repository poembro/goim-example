package business

import (
	"goim-example/internal/logic"
	"goim-example/internal/logic/business/router"
	"goim-example/internal/logic/business/service"
	"goim-example/internal/logic/conf"
	"net/http"

	"github.com/gin-gonic/gin"
)

func New(c *conf.Config, l *logic.Logic) *service.Service {
	// 初始化
	s := service.New(c)
	r := router.New(c, l, s)

	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(r.LoggerHandler, r.RecoverHandler)
	engine.Use(r.CorsMiddleware)

	engine.StaticFS("/_/", http.Dir("./examples/javascript/"))
	engine.StaticFS("/front", http.Dir("./examples/front/"))
	engine.StaticFS("/admin", http.Dir("./examples/admin/"))

	// 消息推送模块
	group := engine.Group("/goim")
	group.POST("/push/keys", r.PushKeys)
	group.POST("/push/mids", r.PushMids)
	group.POST("/push/room", r.PushRoom)
	group.POST("/push/all", r.PushAll)
	group.GET("/online/top", r.OnlineTop)
	group.GET("/online/room", r.OnlineRoom)
	group.GET("/online/total", r.OnlineTotal)

	// 业务模块
	adminGroup := engine.Group("/api")
	{
		adminGroup.GET("/user/create", r.UserCreate)

		adminGroup.POST("/shop/login", r.ShopLogin)
		adminGroup.POST("/shop/register", r.ShopRegister)
		authorized := adminGroup.Group("")
		authorized.Use(r.VerifyMiddleware)
		{
			authorized.POST("/shop/list", r.ShopList)

			authorized.POST("/msg/push", r.MsgPush)
			authorized.POST("/msg/list", r.MsgList)
			authorized.POST("/msg/clear", r.MsgClear)

			authorized.POST("/ipblack/add", r.IpblackCreate)
			authorized.POST("/ipblack/del", r.IpblackRemove)
			authorized.POST("/ipblack/list", r.IpblackList)
		}
	}
	go func() {
		if err := engine.Run(c.HTTPServer.Addr); err != nil {
			panic(err)
		}
	}()

	return s
}
