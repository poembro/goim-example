package business

import (
	"context"
	"goim-demo/internal/logic/business/dao"
	"goim-demo/internal/logic/conf"
)

// Business struct
type Business struct {
	c   *conf.Config
	dao *dao.Dao
}

// New init
func New(c *conf.Config) (s *Business) {
	s = &Business{
		c:   c,
		dao: dao.New(c),
	}
	return
}

// Close Business.
func (s *Business) Close() {
	s.dao.Close()
}

// Ping check server ok.
func (s *Business) Ping(c context.Context) (err error) {
	err = s.dao.Ping(c)
	return
}
