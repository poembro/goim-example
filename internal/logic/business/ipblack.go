package business

import (
	"context"
)

// DelIpblack ip从黑名单删除
func (s *Business) DelIpblack(ctx context.Context, shopId string, ip string) error {
	s.dao.DelIpblack(shopId, ip)
	return nil
}

// AddIpblack ip添加至黑名单
func (s *Business) AddIpblack(ctx context.Context, shopId string, ip string) error {
	s.dao.AddIpblack(shopId, ip)
	return nil
}

// ListIpblack 查询某商户下的用户
func (s *Business) ListIpblack(shopId, min, max string, page, limit int64) ([]string, int64, error) {
	return s.dao.ListIpblack(shopId, min, max, page, limit)
}
