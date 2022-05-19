package business

import (
	"context"
)

// Sync 消息同步
func (s *Business) Sync(ctx context.Context, roomId string, seq int64) (string, error) {
	dst, err := s.dao.GetMessageList(roomId, 0, 50) // 取回最近50条消息
	if err != nil {
		return "", err
	}
	max := len(dst)
	//sort.Sort(sort.Reverse(sort.StringSlice(dst))) //倒序失败

	jsonStr := "["
	for i := max - 1; i >= 0; i-- {
		jsonStr += dst[i]
		if i == 0 {
			continue
		}
		jsonStr += ","
	}
	jsonStr = jsonStr + "]"

	return jsonStr, nil
}

// MessageACK 消息确认机制
func (s *Business) MessageACK(ctx context.Context, deviceId, roomId string, userId, deviceAck, receiveTime int64) error {
	s.dao.AddMessageACKMapping(deviceId, roomId, deviceAck)
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
