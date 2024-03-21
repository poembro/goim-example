package service

import (
	"context"
)

// IpblackRemove ip从黑名单删除
func (s *Service) IpblackRemove(ctx context.Context, shopId string, ip string) error {
	s.dao.IpblackRemove(shopId, ip)
	return nil
}

// IpblackCreate ip添加至黑名单
func (s *Service) IpblackCreate(ctx context.Context, shopId string, ip string) error {
	s.dao.IpblackCreate(shopId, ip)
	return nil
}

// IpblackList 查询某商户下的用户
func (s *Service) IpblackList(shopId, min, max string, page, limit int64) ([]string, int64, error) {
	return s.dao.IpblackList(shopId, min, max, page, limit)
}
