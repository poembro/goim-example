package business

import (
	"context"
	"fmt"
	"goim-demo/conf"
	"goim-demo/pkg/logger"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

var svr *Business

func init() {
	logger.Init()
	svr = New(conf.Conf)
	time.Sleep(time.Second)
}

func WithBusiness(f func(s *Business)) func() {
	return func() {
		Reset(func() {})
		f(svr)
	}
}

// 参考写法 https://blog.csdn.net/weixin_30337227/article/details/121316864
// goconvey -port 8081
func Test_Business(t *testing.T) {
	Convey("business test", t, WithBusiness(func(s *Business) {
		s.Ping(context.TODO())
		s.Close()
		Println("------->")
	}))
}

func TestBusiness_GetArticleMetas(t *testing.T) {
	Convey("sub TimeConf ", t, WithBusiness(func(s *Business) {
		md := s.BuildDeviceId("a", "b")
		//So(err, ShouldBeNil)
		fmt.Printf("%+v", md)
	}))
}
