package business

import (
	"encoding/json"
	"fmt"
	"goim-demo/internal/logic/business/model"
	"goim-demo/internal/logic/business/util"
)

// GetShop 获取后台商户
func (s *Business) GetShop(nickname string) (*model.Shop, error) {
	body, err := s.dao.GetShop(nickname)
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
func (s *Business) AddShop(mid, nickname, face, password string) error {
	dst := model.Shop{
		Mid:      mid,
		Nickname: nickname,
		Face:     face,
		Password: password,
	}

	bytes := util.JsonMarshal(dst)
	return s.dao.AddShop(nickname, bytes)
}
