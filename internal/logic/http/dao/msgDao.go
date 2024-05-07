package dao

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	// 存放推送消息
	_prefixListMsg = "golang_im:messagelist:%s"
	// 存放 已读未读偏移
	_prefixMessageAck = "golang_im:deviceId_msg_ack_%s" // deviceId -> RoomID:ack
)

func KeyListMsg(roomId string) string {
	return fmt.Sprintf(_prefixListMsg, roomId)
}

func keyMessageAck(deviceId string) string {
	return fmt.Sprintf(_prefixMessageAck, deviceId)
}

// MsgACKMappingCreate add a msg ack mapping. 记录用户已读偏移
//
//	HSET userId_123 8000 100000000
func (d *Dao) MsgACKMappingCreate(ctx context.Context, deviceId, roomId string, deviceAck int64) error {
	// 一个用户有N个房间 每个房间都有个已读偏移位置
	_, err := d.RDSCli.HSet(ctx, keyMessageAck(deviceId), roomId, deviceAck).Result()
	if err != nil {
		return err
	}
	d.RDSCli.Expire(ctx, keyMessageAck(deviceId), d.expire).Err()

	return nil
}

// MsgAckMapping 读取某个用户的已读偏移
func (d *Dao) MsgAckMapping(ctx context.Context, deviceId, roomId string) (string, error) {
	// 一个用户有N个房间 每个房间都有个已读偏移位置
	dst, err := d.RDSCli.HGet(ctx, keyMessageAck(deviceId), roomId).Result()
	if err != nil {
		return dst, err
	}

	return dst, err
}

////////////////////////////////////////////////////////

// MsgPush 将消息添加到对应房间 roomId.
// zadd  roomId  time() msg
func (d *Dao) MsgPush(ctx context.Context, roomId string, id int64, msg string) error {
	// NX: 不更新存在的成员,只添加新成员
	// XX: 仅仅更新存在的成员，不添加新成员
	// CH: 更改的元素是新添加的成员，已经存在的成员更新分数
	// INCR: 成员的操作就等同 ZINCRBY 命令，对成员的分数进行递增操作
	err := d.RDSCli.ZAddNX(ctx, KeyListMsg(roomId), redis.Z{
		Score:  float64(id),
		Member: msg,
	}).Err()

	if err != nil {
		return err
	}
	d.RDSCli.Expire(ctx, KeyListMsg(roomId), d.expire).Err()

	return nil
}

// MsgCount 统计未读
func (d *Dao) MsgCount(ctx context.Context, roomId, start, stop string) (int64, error) {
	dst, err := d.RDSCli.ZCount(ctx, KeyListMsg(roomId), start, stop).Result()
	if err != nil {
		return dst, err
	}

	return dst, nil
}

// MsgListByTop 取回消息 返回切片
func (d *Dao) MsgListByTop(ctx context.Context, roomId string, start, stop int64) ([]string, error) {
	dst, err := d.RDSCli.ZRevRange(ctx, KeyListMsg(roomId), start, stop).Result()
	if err != nil {
		return dst, err
	}

	return dst, nil
}

// MsgList 取回消息分页  func("10010", "-", "+", 0, 3)
func (d *Dao) MsgList(ctx context.Context, roomId, min, max string, page, limit int64) ([]string, int64, error) {
	var total int64 // 条数
	var err error
	key := KeyListMsg(roomId)
	total, err = d.RDSCli.ZCount(ctx, key, min, max).Result()

	ids, err := d.RDSCli.ZRevRangeByScore(ctx, key, &redis.ZRangeBy{
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

/////////////////////清理///////////////////////

// MsgClear 前1个月的用户清理掉
func (d *Dao) MsgClear(ctx context.Context) error {
	// 获取所有商户
	shopUsers, err := d.RDSCli.HGetAll(ctx, keyShopList()).Result()
	if err != nil {
		return nil
	}
	for shopId, _ := range shopUsers {
		d.shopIdByUsers(ctx, shopId)
	}
	return nil
}

// 通过shop_id 拿到对应商户下的userId
func (d *Dao) shopIdByUsers(ctx context.Context, shopId string) error {
	t := time.Now().AddDate(0, -1, 0).UnixNano() // 前1个月的记录清理掉
	dst, err := d.RDSCli.ZRangeByScore(ctx, keyShopUsersList(shopId), &redis.ZRangeBy{
		Min: "-",
		Max: strconv.FormatInt(t, 10),
		//Offset:(page - 1) * limit,
		//Count: limit,
	}).Result()

	for _, userId := range dst {
		d.userIdByDeviceId(ctx, userId)
		d.RDSCli.ZRem(ctx, keyShopUsersList(shopId), userId) // 单个删除一个月前的用户元素
	}
	return err
}

// userIdByDeviceId 通过 userId 拿到对应设备编号 deviceId (删除用户信息)
func (d *Dao) userIdByDeviceId(ctx context.Context, userId string) error {
	key := KeyUserIdStrServer(userId)
	deviceIds, err := d.RDSCli.HGetAll(ctx, key).Result()
	if err != nil {
		return nil
	}
	for deviceId, _ := range deviceIds {
		d.deviceIdByRoomID(ctx, deviceId)
		d.RDSCli.HDel(ctx, key, deviceId).Result() // 单个删除用户信息
	}

	d.RDSCli.Del(ctx, key).Result() // 总的删除已读未读偏移
	return err
}

// 通过DeviceId 拿到对应房间 roomId   (删除消息,已读未读偏移)
func (d *Dao) deviceIdByRoomID(ctx context.Context, deviceId string) error {
	key := keyMessageAck(deviceId)
	roomIds, err := d.RDSCli.HGetAll(ctx, key).Result()
	if err != nil {
		return nil
	}

	// 部分删除 前1个月的记录清理掉
	//t := time.Now().AddDate(0, -1, 0).UnixNano()
	//dateline := fmt.Sprintf("%d", t)
	//skey := fmt.Sprintf("messagelist:%s", roomId)
	//d.RDSCli.ZRemRangeByScore(skey, "-", dateline)
	for roomId, _ := range roomIds {
		d.RDSCli.Del(ctx, KeyListMsg(roomId)).Result() // 总的删除消息
		d.RDSCli.HDel(ctx, key, roomId).Result()       // 单个删除已读未读偏移
	}

	d.RDSCli.Del(ctx, key).Result() // 总的删除已读未读偏移
	return nil
}
