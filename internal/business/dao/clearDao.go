package dao

import (
	"strconv"
	"time"

	"github.com/go-redis/redis"
)

type clear struct{}

var Clear = new(clear)

// ClearData 前1个月的用户清理掉
func (d *Dao) ClearData() error {
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
		d.deviceIdByRoomId(deviceId)
		d.RdsCli.HDel(key, deviceId).Result() // 单个删除用户信息
	}

	d.RdsCli.Del(key).Result() // 总的删除已读未读偏移
	return err
}

// 通过DeviceId 拿到对应房间 roomId   (删除消息,已读未读偏移)
func (d *Dao) deviceIdByRoomId(deviceId string) error {
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
