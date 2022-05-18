package apihttp

import (
	"goim-demo/internal/logic/business"
	"goim-demo/internal/logic/business/util"
	 
	"github.com/gin-gonic/gin"
)

type Router struct {
	svc *business.Business
}

func New(s *business.Business) *Router {
	r := &Router{
		svc: s,
	}
	return r
}

func (s *Router) corsMiddleware(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "*")
	c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, PATCH, DELETE")
	c.Header("Access-Control-Max-Age", "3600")
	c.Header("Access-Control-Expose-Headers", "*")
	c.Header("Access-Control-Allow-Credentials", "true")
	if c.Request.Method == "OPTIONS" {
		c.AbortWithStatus(http.StatusNoContent)
	}
	c.Next()
}
 

func (s *Router) verifyMiddleware(c *gin.Context) {
	if r.Method == "OPTIONS" {
		return
	}
	// 解析token
	var token string
	token = c.Query("token")
	if token == "" {
		token = c.GetHeader("token")
	}
	if token == "" {
		OutJson(c, OutData{Code: -1, Success: false, Msg: "参数token不能为空"})
		return
	}
	tokenInfo, err := util.DecryptToken(token)
	if tokenInfo == nil || err != nil {
		OutJson(c, OutData{Code: -1, Success: false, Msg: "参数token认证错误"})
		return
	}

	// 去执行后续handler逻辑
	c.Next()
}

// Close close the server.
func (s *Router) Close() {
	h.svc.Close()
}
