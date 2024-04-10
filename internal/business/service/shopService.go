package service

import (
	"context"
	"encoding/json"
	"fmt"
	"goim-example/internal/business/model"
	"goim-example/internal/business/util"
	"strconv"
)

// ShopFindOne 获取后台商户
func (s *Service) ShopFindOne(ctx context.Context, nickname string) (*model.Shop, error) {
	body, err := s.dao.ShopFindOne(ctx, nickname)
	if err != nil {
		return nil, err
	}

	item := model.Shop{}
	if err := json.Unmarshal(body, &item); err != nil {
		return nil, fmt.Errorf("json.Unmarshal expected ")
	}
	return &item, nil
}

// ShopCreate 添加后台商户
func (s *Service) ShopCreate(ctx context.Context, nickname, face, password string) (*model.Shop, error) {
	Mid := util.SFlake.GetID()
	smid := strconv.FormatInt(Mid, 10)

	req := &model.Shop{
		Mid:      smid,
		Nickname: nickname,
		Face:     face,
		Password: password,
	}
	bytes, err := json.Marshal(req)
	if err != nil {
		panic(err.Error())
	}
	err = s.dao.ShopCreate(ctx, nickname, string(bytes))
	return req, err
}

// ShopByUsers 查询某商户下的所有用户
func (s *Service) ShopByUsers(ctx context.Context, shopId, min, max string, page, limit int64) ([]string, int64, error) {
	return s.dao.ShopByUsers(ctx, shopId, min, max, page, limit)
}

// ShopAppendUserId 临时用户id放入 商户列表
func (s *Service) ShopAppendUserId(ctx context.Context, shopId, mid string) error {
	return s.dao.ShopAppendUserId(ctx, shopId, mid)
}
