package dao

import (
	"fmt"
	"strconv"
	"time"

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

// ClearData 前1个月的用户清理掉
func (d *Dao) MsgClear() error {
	// 获取所有商户
	shopUsers, err := d.RdsCli.HGetAll(keyShopList()).Result()
	if err != nil {
		return nil
	}
	for shopId, _ := range shopUsers {
		d.shopIdByUsers(shopId)
	}
	return nil
}

// 通过shop_id 拿到对应商户下的userId
func (d *Dao) shopIdByUsers(shopId string) error {
	t := time.Now().AddDate(0, -1, 0).UnixNano() // 前1个月的记录清理掉
	dst, err := d.RdsCli.ZRangeByScore(keyShopUsersList(shopId), redis.ZRangeBy{
		Min: "-",
		Max: strconv.FormatInt(t, 10),
		//Offset:(page - 1) * limit,
		//Count: limit,
	}).Result()

	for _, userId := range dst {
		d.userIdByDeviceId(userId)
		d.RdsCli.ZRem(keyShopUsersList(shopId), userId) // 单个删除一个月前的用户元素
	}
	return err
}

// userIdByDeviceId 通过 userId 拿到对应设备编号 deviceId (删除用户信息)
func (d *Dao) userIdByDeviceId(userId string) error {
	key := KeyUserIdStrServer(userId)
	deviceIds, err := d.RdsCli.HGetAll(key).Result()
	if err != nil {
		return nil
	}
	for deviceId, _ := range deviceIds {
		d.deviceIdByRoomID(deviceId)
		d.RdsCli.HDel(key, deviceId).Result() // 单个删除用户信息
	}

	d.RdsCli.Del(key).Result() // 总的删除已读未读偏移
	return err
}

// 通过DeviceId 拿到对应房间 roomId   (删除消息,已读未读偏移)
func (d *Dao) deviceIdByRoomID(deviceId string) error {
	key := keyMessageAck(deviceId)
	roomIds, err := d.RdsCli.HGetAll(key).Result()
	if err != nil {
		return nil
	}

	// 部分删除 前1个月的记录清理掉
	//t := time.Now().AddDate(0, -1, 0).UnixNano()
	//dateline := fmt.Sprintf("%d", t)
	//skey := fmt.Sprintf("messagelist:%s", roomId)
	//d.RdsCli.ZRemRangeByScore(skey, "-", dateline)
	for roomId, _ := range roomIds {
		d.RdsCli.Del(KeyMsgList(roomId)).Result() // 总的删除消息
		d.RdsCli.HDel(key, roomId).Result()       // 单个删除已读未读偏移
	}

	d.RdsCli.Del(key).Result() // 总的删除已读未读偏移
	return nil
}
