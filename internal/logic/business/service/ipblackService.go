package service

import (
	"context"
)

// IpblackDel ip从黑名单删除
func (s *Service) IpblackDel(ctx context.Context, shopId string, ip string) error {
	s.dao.IpblackDel(shopId, ip)
	return nil
}

// IpblackAdd ip添加至黑名单
func (s *Service) IpblackAdd(ctx context.Context, shopId string, ip string) error {
	s.dao.IpblackAdd(shopId, ip)
	return nil
}

// IpblackList 查询某商户下的用户
func (s *Service) IpblackList(shopId, min, max string, page, limit int64) ([]string, int64, error) {
	return s.dao.IpblackList(shopId, min, max, page, limit)
}
