package service

import (
	"context"
	"encoding/json"
	"goim-example/internal/logic/http/util"
	"strconv"

	log "github.com/golang/glog"
)

// MsgSync 消息同步 (comet服务通过grpc发来的body参数)
func (s *Service) MsgSync(ctx context.Context, mid int64, body []byte) (int32, []string, []byte, error) {
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

	items, err := s.dao.MsgListByTop(ctx, arg.RoomID, idx, idx+50) // 取回最近50条消息
	if err != nil || len(items) == 0 {
		return 0, nil, nil, err
	}
	max := len(items)
	jsonStr := "["
	for i := max - 1; i >= 0; i-- {
		jsonStr += items[i]
		if i == 0 {
			continue
		}
		jsonStr += ","
	}
	jsonStr = jsonStr + "]"
	return arg.Op, []string{arg.Key}, util.S2B(jsonStr), nil
}

// MsgList 取回消息
func (s *Service) MsgListByTop(ctx context.Context, roomId string, start, stop int64) ([]string, error) {
	return s.dao.MsgListByTop(ctx, roomId, start, stop)
}

// MessageACK 消息确认机制 (comet服务通过grpc发来的body参数)
func (s *Service) MessageACK(ctx context.Context, mid int64, body []byte) error {
	var arg struct {
		ID     string `json:"id"`
		Key    string `json:"key"`
		RoomID string `json:"room_id"`
	}
	if err := json.Unmarshal(body, &arg); err != nil {
		log.Errorf("json.Unmarshal(%s) error(%v)", body, err)
		return err
	}
	id, _ := strconv.ParseInt(arg.ID, 10, 64)
	s.dao.MsgACKMappingCreate(ctx, arg.Key, arg.RoomID, id)
	return nil
}

// MsgCount 统计未读
func (s *Service) MsgCount(ctx context.Context, roomId, start, stop string) (int64, error) {
	return s.dao.MsgCount(ctx, roomId, start, stop)
}

// MsgList 取回消息分页
func (s *Service) MsgList(ctx context.Context, roomId, min, max string, page, limit int64) ([]string, int64, error) {
	return s.dao.MsgList(ctx, roomId, min, max, page, limit)
}

// MsgPush 将消息添加到对应房间 roomId.
func (s *Service) MsgPush(ctx context.Context, roomId string, id int64, msg string) error {
	return s.dao.MsgPush(ctx, roomId, id, msg)
}

// MsgClear 清理数据
func (s *Service) MsgClear(ctx context.Context) error {
	s.dao.MsgClear(ctx)
	return nil
}

// MsgAckMapping 读取某个用户的已读偏移
func (s *Service) MsgAckMapping(ctx context.Context, deviceId, roomId string) (string, error) {
	return s.dao.MsgAckMapping(ctx, deviceId, roomId)
}
