package service

import (
	"context"
	"goim-example/internal/logic/business/dao"
	"goim-example/internal/logic/conf"
	"sync"
)

// Service struct
type Service struct {
	c   *conf.Config
	dao *dao.Dao
}

var (
	once sync.Once
	svc  *Service = nil
)

// New init
func New(c *conf.Config) (s *Service) {
	if svc != nil {
		return svc
	}

	once.Do(func() {
		svc = &Service{
			c:   c,
			dao: dao.New(c),
		}
	})

	return svc
}

// Close Service.
func (s *Service) Close() {
	s.dao.Close()
}

// Ping check server ok.
func (s *Service) Ping(c context.Context) (err error) {
	err = s.dao.Ping(c)
	return
}
