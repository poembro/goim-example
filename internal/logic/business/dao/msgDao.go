package dao

import (
	"fmt"

	"github.com/go-redis/redis"
)

const (
	_prefixMsgList = "golang_im:messagelist:%s"
)

func KeyMsgList(roomId string) string {
	return fmt.Sprintf(_prefixMsgList, roomId)
}

// AddMessageList 将消息添加到对应房间 roomId.
// zadd  roomId  time() msg
func (d *Dao) AddMessageList(roomId string, id int64, msg string) error {
	// NX: 不更新存在的成员,只添加新成员
	// XX: 仅仅更新存在的成员，不添加新成员
	// CH: 更改的元素是新添加的成员，已经存在的成员更新分数
	// INCR: 成员的操作就等同 ZINCRBY 命令，对成员的分数进行递增操作
	err := d.RdsCli.ZAddNX(KeyMsgList(roomId), redis.Z{
		Score:  float64(id),
		Member: msg,
	}).Err()

	if err != nil {
		return err
	}

	return nil
}

// GetMessageCount 统计未读
func (d *Dao) GetMessageCount(roomId, start, stop string) (int64, error) {
	dst, err := d.RdsCli.ZCount(KeyMsgList(roomId), start, stop).Result()
	if err != nil {
		return dst, err
	}

	return dst, nil
}

// GetMessageList 取回消息 返回切片
func (d *Dao) GetMessageList(roomId string, start, stop int64) ([]string, error) {
	dst, err := d.RdsCli.ZRevRange(KeyMsgList(roomId), start, stop).Result()
	if err != nil {
		return dst, err
	}

	return dst, nil
}

// GetMessagePageList 取回消息分页  func("10010", "-", "+", 0, 3)
func (d *Dao) GetMessagePageList(roomId, min, max string, page, limit int64) ([]string, int64, error) {
	var total int64 // 条数
	var err error
	key := KeyMsgList(roomId)
	total, err = d.RdsCli.ZCount(key, min, max).Result()

	ids, err := d.RdsCli.ZRevRangeByScore(key, redis.ZRangeBy{
		Min:    min, //"-inf"
		Max:    max, // "+inf"
		Offset: (page - 1) * limit,
		Count:  limit,
	}).Result()

	if err != nil {
		return ids, total, err
	}

	return ids, total, nil
}
