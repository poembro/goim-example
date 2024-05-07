package dao

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
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
func (d *Dao) ShopCreate(ctx context.Context, nickname, item string) error {
	d.RDSCli.HSet(ctx, keyShopList(), nickname, item).Result()

	d.RDSCli.Expire(ctx, keyShopList(), d.expire).Err()

	return nil
}

// ShopFindOne 获取后台商户
func (d *Dao) ShopFindOne(ctx context.Context, nickname string) ([]byte, error) {
	return d.RDSCli.HGet(ctx, keyShopList(), nickname).Bytes()
}

// ShopByUsers 查询某商户下的用户
// zrevrange  shop_id  0, 50
func (d *Dao) ShopByUsers(ctx context.Context, shopId, min, max string, page, limit int64) ([]string, int64, error) {
	var total int64 // 条数
	var err error
	key := keyShopUsersList(shopId)
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

// ShopAppendUser 将用户添加到商户列表
// zadd  shop_id  time() user_id
func (d *Dao) ShopAppendUserId(ctx context.Context, shopId string, userId string) error {
	score := time.Now().UnixNano()
	err := d.RDSCli.ZAdd(ctx, keyShopUsersList(shopId), redis.Z{
		Score:  float64(score),
		Member: userId,
	}).Err()

	if err != nil {
		return err
	}
	d.RDSCli.Expire(ctx, keyShopUsersList(shopId), d.expire).Err()

	return nil
}
