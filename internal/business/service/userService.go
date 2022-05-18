package service

import (
	"context"
	"encoding/json"
	"fmt"
	"goim-demo/internal/logic/model"
	"goim-demo/pkg/util"
)

// Heartbeat 心跳
func (s *Service) Heartbeat(ctx context.Context, userId int64, deviceId, connAddr string) error {
	s.dao.ExpireMapping(userId, deviceId)
	return nil
}

// Offline 离线
func (s *Service) Offline(ctx context.Context, userId int64, deviceId, connAddr string) error {
	s.dao.DelMapping(userId, deviceId)
	return nil
}

// KeysByUserIds 通过userId 获取用户信息
func (s *Service) KeysByUserIds(userIds []int64) (map[string]string, error) {
	return s.dao.KeysByUserIds(userIds)
}

// GetShopByUsers 查询在线人数
func (s *Service) GetShopByUsers(shopId, min, max string, page, limit int64) ([]string, int64, error) {
	return s.dao.GetShopByUsers(shopId, min, max, page, limit)
}

// GetMessageAckMapping 读取某个用户的已读偏移
func (s *Service) GetMessageAckMapping(deviceId, roomId string) (string, error) {
	return s.dao.GetMessageAckMapping(deviceId, roomId)
}

// IsOnline 是否在线
func (s *Service) IsOnline(deviceId string) bool {
	return s.dao.IsOnline(deviceId)
}

// GetShop 获取后台商户
func (s *Service) GetShop(mobile string) (*model.Shop, error) {
	body, err := s.dao.GetShop(mobile)
	if err != nil {
		return nil, err
	}

	shop := new(model.Shop)
	if err := json.Unmarshal(body, shop); err != nil {
		return nil, fmt.Errorf("json.Unmarshal expected ")
	}
	return shop, nil
}

// AddShop 添加后台商户
func (s *Service) AddShop(userId, nickname, face, password string) error {
	dst := model.Shop{
		UserId:   userId,
		Nickname: nickname,
		Face:     face,
		Password: password,
	}

	bytes := util.JsonMarshal(dst)
	return s.dao.AddShop(nickname, bytes)
}
