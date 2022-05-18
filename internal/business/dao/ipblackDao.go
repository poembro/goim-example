package dao

import (
	"fmt"
	"goim-demo/pkg/gerrors"
	"time"

	"github.com/go-redis/redis"
)

const (
	_prefixIpblackList = "golang_im:shop_ip_black_list:%s"
)

func keyIpblackList(shopId string) string {
	return fmt.Sprintf(_prefixIpblackList, shopId)
}

// DelIpblack ip从黑名单删除
// zadd  shop_id  time() ip
func (d *Dao) DelIpblack(shopId string, ip string) error {
	err := d.RdsCli.ZRem(keyIpblackList(shopId), ip).Err()
	if err != nil {
		return gerrors.WrapError(err)
	}

	return nil
}

// AddIpblack ip添加至黑名单
// zadd  shop_id  time() ip
func (d *Dao) AddIpblack(shopId string, ip string) error {
	score := time.Now().UnixNano()
	err := d.RdsCli.ZAdd(keyIpblackList(shopId), redis.Z{
		Score:  float64(score),
		Member: ip,
	}).Err()

	if err != nil {
		return gerrors.WrapError(err)
	}

	return nil
}

// ListIpblack 查询某商户下的用户
// zrevrange  shop_id  0, 50
func (d *Dao) ListIpblack(shopId, min, max string, page, limit int64) ([]string, int64, error) {
	var total int64 // 条数
	var err error
	key := keyIpblackList(shopId)
	total, err = d.RdsCli.ZCount(key, min, max).Result()

	ids, err := d.RdsCli.ZRevRangeByScore(key, redis.ZRangeBy{
		Min:    min, //"-inf"
		Max:    max, // "+inf"
		Offset: (page - 1) * limit,
		Count:  limit,
	}).Result()

	if err != nil {
		return ids, total, gerrors.WrapError(err)
	}

	return ids, total, nil
}
