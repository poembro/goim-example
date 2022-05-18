package business

import (
	"context"
)

// ClearData 清理数据
func (s *Business) ClearData(ctx context.Context) error {
	s.dao.ClearData()
	return nil
}
