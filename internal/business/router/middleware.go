package router

import (
	"fmt"
	"goim-example/internal/business/util"
	"net/http/httputil"
	"os"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/golang/glog"
)

func (s *Router) CorsMiddleware(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "*")
	c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, PATCH, DELETE")
	c.Header("Access-Control-Max-Age", "3600")
	c.Header("Access-Control-Expose-Headers", "*")
	c.Header("Access-Control-Allow-Credentials", "true")
	if c.Request.Method == "OPTIONS" {
		c.AbortWithStatus(200)
	}
	c.Next()
}

func (s *Router) VerifyMiddleware(c *gin.Context) {
	// 解析token
	var token string
	token = c.Query("token")
	if token == "" {
		token = c.GetHeader("token")
	}
	if token == "" {
		s.OutJson(c, -1, "参数token不能为空", nil)
		return
	}
	tokenInfo, err := util.DecryptToken(token)
	if tokenInfo == nil || err != nil {
		s.OutJson(c, -1, "参数token认证错误", nil)
		return
	}

	// c.Set("username", "value") 将用户信息保存到ctx,  方便后续handler读取

	// 去执行后续handler逻辑
	c.Next()
}

func (s *Router) LoggerHandler(c *gin.Context) {
	// Start timer
	start := time.Now()
	path := c.Request.URL.Path
	raw := c.Request.URL.RawQuery
	method := c.Request.Method

	// Process request
	c.Next()

	// Stop timer
	end := time.Now()
	latency := end.Sub(start)
	statusCode := c.Writer.Status()
	ecode := c.GetInt(contextErrCodeKey)
	clientIP := c.ClientIP()
	if raw != "" {
		path = path + "?" + raw
	}
	log.Infof("METHOD:%s | PATH:%s | CODE:%d | IP:%s | TIME:%d | ECODE:%d", method, path, statusCode, clientIP, latency/time.Millisecond, ecode)
}

func (s *Router) RecoverHandler(c *gin.Context) {
	defer func() {
		if err := recover(); err != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			httprequest, _ := httputil.DumpRequest(c.Request, false)
			pnc := fmt.Sprintf("[Recovery] %s panic recovered:\n%s\n%s\n%s", time.Now().Format("2006-01-02 15:04:05"), string(httprequest), err, buf)
			fmt.Fprintf(os.Stderr, pnc)
			log.Error(pnc)
			c.AbortWithStatus(500)
		}
	}()
	c.Next()
}
