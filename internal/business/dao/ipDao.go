package dao

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	// 存放黑名单数据
	_prefixListIp = "golang_im:shop_ip_black_list:%s"
)

func keyListIp(shopId string) string {
	return fmt.Sprintf(_prefixListIp, shopId)
}

// IpRemove ip从黑名单删除
// zadd  shop_id  time() ip
func (d *Dao) IpRemove(ctx context.Context, shopId string, ip string) error {
	err := d.RDSCli.ZRem(ctx, keyListIp(shopId), ip).Err()
	if err != nil {
		return err
	}

	return nil
}

// IpCreate ip添加至黑名单
// zadd  shop_id  time() ip
func (d *Dao) IpCreate(ctx context.Context, shopId string, ip string) error {
	score := time.Now().UnixNano()
	err := d.RDSCli.ZAdd(ctx, keyListIp(shopId), redis.Z{
		Score:  float64(score),
		Member: ip,
	}).Err()

	if err != nil {
		return err
	}

	return nil
}

// ListIp 查询某商户下的用户
// zrevrange  shop_id  0, 50
func (d *Dao) IpList(ctx context.Context, shopId, min, max string, page, limit int64) ([]string, int64, error) {
	var total int64 // 条数
	var err error
	key := keyListIp(shopId)
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
