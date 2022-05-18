package dao

import (
	"fmt"
	"time"

	"github.com/go-redis/redis"
)

const (
	_prefixMidServer    = "golang_im:userId_%d"
	_prefixMidStrServer = "golang_im:userId_%s"           // userId -> DeviceId:userinfo
	_prefixKeyServer    = "golang_im:deviceId_%s"         // deviceId -> server
	_prefixServerOnline = "golang_im:ol_%s"               // server -> online
	_prefixMessageAck   = "golang_im:deviceId_msg_ack_%s" // deviceId -> RoomID:ack
)

func KeyUserIdServer(userId int64) string {
	return fmt.Sprintf(_prefixMidServer, userId)
}

func KeyUserIdStrServer(userId string) string {
	return fmt.Sprintf(_prefixMidStrServer, userId)
}

func KeyDeviceIdServer(deviceId string) string {
	return fmt.Sprintf(_prefixKeyServer, deviceId)
}

func keyServerOnline(deviceId string) string {
	return fmt.Sprintf(_prefixServerOnline, deviceId)
}

func keyMessageAck(deviceId string) string {
	return fmt.Sprintf(_prefixMessageAck, deviceId)
}

// KeysByUserIds get a deviceId server by userId.
// HGETALL userId_123
func (d *Dao) KeysByUserIds(userIds []int64) (map[string]string, error) {
	dst := make(map[string]string)
	for _, userId := range userIds {
		data, err := d.RdsCli.HGetAll(KeyUserIdServer(userId)).Result()
		if err != nil {
			continue
		}

		for k, v := range data {
			if v != "" {
				dst[k] = v
			}
		}
	}
	return dst, nil
}

// AddMapping add a mapping.
//    HSET userId_123 2000aa78df60000 {id:1,nickname:张三,face:p.png,}
//    SET  deviceId_2000aa78df60000  192.168.3.222
func (d *Dao) AddMapping(userId int64, deviceId, server, userinfo string) error {
	// 一个用户有N个设备 全部在hset上面
	_, err := d.RdsCli.HSet(KeyUserIdServer(userId), deviceId, userinfo).Result()
	if err != nil {
		return err
	}

	_, err = d.RdsCli.Set(KeyDeviceIdServer(deviceId), server, d.expire).Result()
	if err != nil {
		return err
	}

	return nil
}

// ExpireMapping expire a mapping.
//EXPIRE userId_123 2000aa78df60000 1000
//EXPIRE deviceId_2000aa78df60000 1000
func (d *Dao) ExpireMapping(userId int64, deviceId string) error {
	_, err := d.RdsCli.Expire(KeyDeviceIdServer(deviceId), d.expire).Result()
	if err != nil {
		return err
	}
	return nil
}

func (d *Dao) IsOnline(deviceId string) bool {
	return d.RdsCli.TTL(KeyDeviceIdServer(deviceId)).Val().Nanoseconds() > 0
}

// DelMapping del a mapping.
// HDEL userId_123 2000aa78df60000
// DEL  deviceId_2000aa78df60000
func (d *Dao) DelMapping(userId int64, deviceId string) error {
	_, err := d.RdsCli.Del(KeyDeviceIdServer(deviceId)).Result()
	if err != nil {
		return err
	}

	return nil
}

// AddMessageACKMapping add a msg ack mapping. 记录用户已读偏移
//    HSET userId_123 8000 100000000
func (d *Dao) AddMessageACKMapping(deviceId, roomId string, deviceAck int64) error {
	// 一个用户有N个房间 每个房间都有个已读偏移位置
	_, err := d.RdsCli.HSet(keyMessageAck(deviceId), roomId, deviceAck).Result()
	if err != nil {
		return err
	}

	return nil
}

// GetMessageAckMapping 读取某个用户的已读偏移
func (d *Dao) GetMessageAckMapping(deviceId, roomId string) (string, error) {
	// 一个用户有N个房间 每个房间都有个已读偏移位置
	dst, err := d.RdsCli.HGet(keyMessageAck(deviceId), roomId).Result()
	if err != nil {
		return dst, err
	}

	return dst, err
}

// AddUserByShop 将用户添加到商户列表
// zadd  shop_id  time() user_id
func (d *Dao) AddUserByShop(shopId string, userId string) error {
	score := time.Now().UnixNano()
	err := d.RdsCli.ZAdd(keyShopUsersList(shopId), redis.Z{
		Score:  float64(score),
		Member: userId,
	}).Err()

	if err != nil {
		return err
	}

	return nil
}
