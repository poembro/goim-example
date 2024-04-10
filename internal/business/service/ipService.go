package service

import (
	"context"
)

// IpRemove ip从黑名单删除
func (s *Service) IpRemove(ctx context.Context, shopId string, ip string) error {
	s.dao.IpRemove(ctx, shopId, ip)
	return nil
}

// IpCreate ip添加至黑名单
func (s *Service) IpCreate(ctx context.Context, shopId string, ip string) error {
	s.dao.IpCreate(ctx, shopId, ip)
	return nil
}

// IpList 查询某商户下的用户
func (s *Service) IpList(ctx context.Context, shopId, min, max string, page, limit int64) ([]string, int64, error) {
	return s.dao.IpList(ctx, shopId, min, max, page, limit)
}
