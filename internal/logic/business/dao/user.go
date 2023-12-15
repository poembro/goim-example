package dao

import (
	"fmt"
	"time"

	"github.com/go-redis/redis"
)

const (
	// 存放用户详细
	_prefixMidServer    = "golang_im:userId_%d" // hset userId deviceId  json
	_prefixMidStrServer = "golang_im:userId_%s" // userId -> DeviceId:userinfo
)

func KeyUserIdServer(userId int64) string {
	return fmt.Sprintf(_prefixMidServer, userId)
}

// msg 消息删除时用到了
func KeyUserIdStrServer(userId string) string {
	return fmt.Sprintf(_prefixMidStrServer, userId)
}

// KeysByUserIds get a deviceId server by userId.
// HGETALL userId_123
func (d *Dao) KeysByUserIds(userIds []int64) (map[string]string, error) {
	dst := make(map[string]string)
	for _, userId := range userIds {
		data, err := d.RDSCli.HGetAll(KeyUserIdServer(userId)).Result()
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
	_, err := d.RDSCli.HSet(KeyUserIdServer(userId), deviceId, userinfo).Result()
	if err != nil {
		return err
	}
	return nil
}

// AddUserByShop 将用户添加到商户列表
// zadd  shop_id  time() user_id
func (d *Dao) AddUserByShop(shopId string, userId string) error {
	score := time.Now().UnixNano()
	err := d.RDSCli.ZAdd(keyShopUsersList(shopId), redis.Z{
		Score:  float64(score),
		Member: userId,
	}).Err()

	if err != nil {
		return err
	}

	return nil
}
