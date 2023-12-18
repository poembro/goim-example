package dao

import (
	"fmt"
	"time"

	"github.com/go-redis/redis"
)

const (
	// 存后台用户登录数据
	_prefixShopList      = "golang_im:shop_list" //商户列表
	_prefixShopUsersList = "golang_im:shop_user_list:%s"
)

func keyShopList() string {
	return _prefixShopList
}

func keyShopUsersList(shopId string) string {
	return fmt.Sprintf(_prefixShopUsersList, shopId)
}

// ShopCreate 添加后台商户
func (d *Dao) ShopCreate(nickname, item string) error {
	d.RDSCli.HSet(keyShopList(), nickname, item).Result()
	return nil
}

// ShopFindOne 获取后台商户
func (d *Dao) ShopFindOne(nickname string) ([]byte, error) {
	return d.RDSCli.HGet(keyShopList(), nickname).Bytes()
}

// ShopByUsers 查询某商户下的用户
// zrevrange  shop_id  0, 50
func (d *Dao) ShopByUsers(shopId, min, max string, page, limit int64) ([]string, int64, error) {
	var total int64 // 条数
	var err error
	key := keyShopUsersList(shopId)
	total, err = d.RDSCli.ZCount(key, min, max).Result()

	ids, err := d.RDSCli.ZRevRangeByScore(key, redis.ZRangeBy{
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

// ShopAppendUser 将用户添加到商户列表
// zadd  shop_id  time() user_id
func (d *Dao) ShopAppendUserId(shopId string, userId string) error {
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
