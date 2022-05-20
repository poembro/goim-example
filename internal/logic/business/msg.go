package business

import (
	"context"
	"encoding/json"
	"goim-demo/internal/logic/business/util"

	log "github.com/golang/glog"
)

// Sync 消息同步
func (s *Business) Sync(ctx context.Context, mid int64, body []byte) (int32, []string, []byte, error) {
	var arg struct {
		Op     int32  `json:"op"`
		Page   int64  `json:"id"`
		Key    string `json:"key"`
		RoomID string `json:"room_id"`
	}
	if err := json.Unmarshal(body, &arg); err != nil {
		log.Errorf("json.Unmarshal(%s) error(%v)", body, err)
		return 0, nil, nil, err
	}
	idx := (arg.Page - 1) * 50
	if idx < 0 {
		idx = 0
	}

	dst, err := s.dao.GetMessageList(arg.RoomID, idx, idx+50) // 取回最近50条消息
	if err != nil || len(dst) == 0 {
		return 0, nil, nil, err
	}
	max := len(dst)
	jsonStr := "["
	for i := max - 1; i >= 0; i-- {
		jsonStr += dst[i]
		if i == 0 {
			continue
		}
		jsonStr += ","
	}
	jsonStr = jsonStr + "]"
	return arg.Op, []string{arg.Key}, util.S2B(jsonStr), nil
}

// MessageACK 消息确认机制
func (s *Business) MessageACK(ctx context.Context, mid int64, body []byte) error {
	var params struct {
		ID     int64  `json:"id"`
		Key    string `json:"key"`
		RoomID string `json:"room_id"`
	}
	if err := json.Unmarshal(body, &params); err != nil {
		log.Errorf("json.Unmarshal(%s) error(%v)", body, err)
		return err
	}
	s.dao.AddMessageACKMapping(params.Key, params.RoomID, params.ID)
	return nil
}

// GetMessageCount 统计未读
func (s *Business) GetMessageCount(roomId, start, stop string) (int64, error) {
	return s.dao.GetMessageCount(roomId, start, stop)
}

// GetMessageList 取回消息
func (s *Business) GetMessageList(roomId string, start, stop int64) ([]string, error) {
	return s.dao.GetMessageList(roomId, start, stop)
}

// GetMessagePageList 取回消息分页
func (s *Business) GetMessagePageList(roomId, min, max string, page, limit int64) ([]string, int64, error) {
	return s.dao.GetMessagePageList(roomId, min, max, page, limit)
}

// AddMessageList 将消息添加到对应房间 roomId.
func (s *Business) AddMessageList(roomId string, id int64, msg string) error {
	return s.dao.AddMessageList(roomId, id, msg)
}

// MsgClear 清理数据
func (s *Business) MsgClear(ctx context.Context) error {
	s.dao.MsgClear()
	return nil
}
