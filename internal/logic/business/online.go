package business

import (
	"context"
)

// Heartbeat 心跳
func (s *Business) Heartbeat(ctx context.Context, userId int64, deviceId, connAddr string) error {
	s.dao.ExpireMapping(userId, deviceId)
	return nil
}

// Offline 离线
func (s *Business) Offline(ctx context.Context, userId int64, deviceId, connAddr string) error {
	s.dao.DelMapping(userId, deviceId)
	return nil
}

// KeysByUserIds 通过userId 获取用户信息
func (s *Business) KeysByUserIds(userIds []int64) (map[string]string, error) {
	return s.dao.KeysByUserIds(userIds)
}

// GetShopByUsers 查询在线人数
func (s *Business) GetShopByUsers(shopId, min, max string, page, limit int64) ([]string, int64, error) {
	return s.dao.GetShopByUsers(shopId, min, max, page, limit)
}

// GetMessageAckMapping 读取某个用户的已读偏移
func (s *Business) GetMessageAckMapping(deviceId, roomId string) (string, error) {
	return s.dao.GetMessageAckMapping(deviceId, roomId)
}

// IsOnline 是否在线
func (s *Business) IsOnline(deviceId string) bool {
	return s.dao.IsOnline(deviceId)
}
