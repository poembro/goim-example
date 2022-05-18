package service

import (
	"context"
)

// ClearData 清理数据
func (s *Service) ClearData(ctx context.Context) error {
	s.dao.ClearData()
	return nil
}
