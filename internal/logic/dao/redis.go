package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"goim-example/internal/logic/model"
	"strconv"

	log "github.com/golang/glog"
	"github.com/zhenjl/cityhash"
)

const (
	_prefixMidServer    = "golang_im:mid_%d" // mid -> key:server
	_prefixKeyServer    = "golang_im:key_%s" // key -> server
	_prefixServerOnline = "golang_im:ol_%s"  // server -> online
)

func keyMidServer(mid int64) string {
	return fmt.Sprintf(_prefixMidServer, mid)
}

func keyKeyServer(key string) string {
	return fmt.Sprintf(_prefixKeyServer, key)
}

func keyServerOnline(key string) string {
	return fmt.Sprintf(_prefixServerOnline, key)
}

// pingRedis check redis connection.
func (d *Dao) pingRedis(c context.Context) (err error) {
	return d.redis.Ping(c).Err()
}

// AddMapping add a mapping.
// Mapping:
//
//	HSET mid_123 2000aa78df60000 192.168.3.222
//	SET  key_2000aa78df60000  192.168.3.222
func (d *Dao) AddMapping(c context.Context, mid int64, key, server string) (err error) {
	d.redis.HSet(c, keyMidServer(mid), key, server).Result()
	d.redis.Expire(c, keyMidServer(mid), d.redisExpire)
	d.redis.Set(c, keyKeyServer(key), server, d.redisExpire).Result()
	return nil
}

// ExpireMapping expire a mapping.
// EXPIRE mid_123 2000aa78df60000 1000
// EXPIRE key_2000aa78df60000 1000
func (d *Dao) ExpireMapping(c context.Context, mid int64, key string) (has bool, err error) {
	d.redis.Expire(c, keyMidServer(mid), d.redisExpire)
	d.redis.Expire(c, keyKeyServer(key), d.redisExpire)
	return
}

// DelMapping del a mapping.
// HDEL mid_123 2000aa78df60000
// DEL  key_2000aa78df60000
func (d *Dao) DelMapping(c context.Context, mid int64, key, server string) (has bool, err error) {
	d.redis.HDel(c, keyMidServer(mid), key).Err()
	d.redis.Del(c, keyKeyServer(key)).Err()
	return
}

// ServersByKeys get a server by key.
// MGET key_2000aa78df60000  key_2000aa78df60000  key_2000aa78df60000
func (d *Dao) ServersByKeys(c context.Context, keys []string) (res []string, err error) {
	var args []string
	for _, key := range keys {
		args = append(args, keyKeyServer(key))
	}
	items := d.redis.MGet(c, args...).Val()
	for _, v := range items {
		res = append(res, v.(string))
	}
	return
}

// KeysByMids get a key server by mid.  HSET mid_123 2000aa78df60000 192.168.3.222
// HGETALL mid_123
func (d *Dao) KeysByMids(c context.Context, mids []int64) (map[string]string, []int64, error) {
	ress := make(map[string]string, len(mids))
	for _, mid := range mids {
		res := d.redis.HGetAll(c, keyMidServer(mid)).Val()
		for k, v := range res {
			ress[k] = v
		}
	}
	return ress, mids, nil
}

func (d *Dao) AddServerOnline(c context.Context, server string, online *model.Online) (err error) {
	var idx uint32
	roomsMap := map[uint32]map[string]int32{}
	for room, count := range online.RoomCount { // live://1000  1
		idx = cityhash.CityHash32([]byte(room), uint32(len(room))) % 64
		rMap := roomsMap[idx]
		if rMap == nil {
			rMap = make(map[string]int32)
			roomsMap[idx] = rMap
		}
		rMap[room] = count
	}
	key := keyServerOnline(server)
	for hashKey, value := range roomsMap {
		hashVal := strconv.FormatInt(int64(hashKey), 10)
		err = d.addServerOnline(c, key, hashVal, &model.Online{RoomCount: value, Server: online.Server, Updated: online.Updated})
		if err != nil {
			return
		}
	}
	return
}

// addServerOnline 将对应comet服务写入redis "HSET" "ol_192.168.3.222" "43" "{\"server\":\"192.168.3.222\",\"room_count\":{\"live://1000\":1},\"updated\":1577077540}"
func (d *Dao) addServerOnline(c context.Context, key string, hashKey string, online *model.Online) (err error) {
	b, _ := json.Marshal(online)
	err = d.redis.HSet(c, key, hashKey, b).Err()

	if err != nil {
		log.Errorf("Logic:addServerOnline  key:%s  d.redisExpire : %d  err :%s", key, d.redisExpire, err.Error())
	}
	d.redis.Expire(c, key, d.redisExpire).Err()
	return
}

// ServerOnline get a server online.  logic服务初始化时每隔10秒调了这里
func (d *Dao) GetServerOnline(c context.Context, server string) (online *model.Online, err error) {
	online = &model.Online{RoomCount: map[string]int32{}}
	key := keyServerOnline(server) // HGET "ol_192.168.3.100" (hash(live://1000) % 64)
	for i := 0; i < 64; i++ {
		ol, err := d.getServerOnline(c, key, strconv.FormatInt(int64(i), 10))
		if err == nil && ol != nil {
			online.Server = ol.Server
			if ol.Updated > online.Updated {
				online.Updated = ol.Updated
			}
			for room, count := range ol.RoomCount {
				online.RoomCount[room] = count
			}
		}
	}
	return
}

func (d *Dao) getServerOnline(c context.Context, key string, hashKey string) (online *model.Online, err error) {
	b, err := d.redis.HGet(c, key, hashKey).Bytes()
	if err != nil {
		return
	}

	if err = json.Unmarshal(b, online); err != nil {
		log.Errorf("serverOnline json.Unmarshal(%s) error(%v)", b, err)
		return
	}
	return
}

// DelServerOnline del a server online.
// DEL "ol_192.168.3.100"
func (d *Dao) DelServerOnline(c context.Context, server string) error {
	key := keyServerOnline(server)
	return d.redis.Del(c, key).Err()
}
