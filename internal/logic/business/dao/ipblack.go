package dao

import (
	"fmt"
	"time"

	"github.com/go-redis/redis"
)

const (
	// 存放黑名单数据
	_prefixListIpblack = "golang_im:shop_ip_black_list:%s"
)

func keyListIpblack(shopId string) string {
	return fmt.Sprintf(_prefixListIpblack, shopId)
}

// DelIpblack ip从黑名单删除
// zadd  shop_id  time() ip
func (d *Dao) DelIpblack(shopId string, ip string) error {
	err := d.RdsCli.ZRem(keyListIpblack(shopId), ip).Err()
	if err != nil {
		return err
	}

	return nil
}

// AddIpblack ip添加至黑名单
// zadd  shop_id  time() ip
func (d *Dao) AddIpblack(shopId string, ip string) error {
	score := time.Now().UnixNano()
	err := d.RdsCli.ZAdd(keyListIpblack(shopId), redis.Z{
		Score:  float64(score),
		Member: ip,
	}).Err()

	if err != nil {
		return err
	}

	return nil
}

// ListIpblack 查询某商户下的用户
// zrevrange  shop_id  0, 50
func (d *Dao) ListIpblack(shopId, min, max string, page, limit int64) ([]string, int64, error) {
	var total int64 // 条数
	var err error
	key := keyListIpblack(shopId)
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
