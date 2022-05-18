package service

import (
	"context"
	"fmt"
	"goim-demo/conf"
	"goim-demo/pkg/logger"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

var svr *Service

func init() {
	logger.Init()
	svr = New(conf.Conf)
	time.Sleep(time.Second)
}

func WithService(f func(s *Service)) func() {
	return func() {
		Reset(func() {})
		f(svr)
	}
}

// 参考写法 https://blog.csdn.net/weixin_30337227/article/details/121316864
// goconvey -port 8081
func Test_Service(t *testing.T) {
	Convey("service test", t, WithService(func(s *Service) {
		s.Ping(context.TODO())
		s.Close()
		Println("------->")
	}))
}

func TestService_GetArticleMetas(t *testing.T) {
	Convey("sub TimeConf ", t, WithService(func(s *Service) {
		md := s.BuildDeviceId("a", "b")
		//So(err, ShouldBeNil)
		fmt.Printf("%+v", md)
	}))
}
