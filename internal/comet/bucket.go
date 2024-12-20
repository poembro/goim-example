package comet

import (
	"sync"
	"sync/atomic"

	pb "goim-example/api/comet"
	"goim-example/api/protocol"
	"goim-example/internal/comet/conf"
)

// Bucket is a channel holder.
type Bucket struct {
	c     *conf.Bucket
	cLock sync.RWMutex        // protect the channels for chs
	chs   map[string]*Channel // map sub key to a channel

	// room
	rooms       map[string]*Room // bucket room channels
	routines    []chan *pb.BroadcastRoomReq
	routinesNum uint64

	ipCnts map[string]int32
}

// NewBucket 新建一个桶结构
func NewBucket(c *conf.Bucket) (b *Bucket) {
	b = new(Bucket)
	b.chs = make(map[string]*Channel, c.Channel) //c.Channel 通道1024个
	b.ipCnts = make(map[string]int32)
	b.c = c
	b.rooms = make(map[string]*Room, c.Room)                        //c.Room 房间数1024
	b.routines = make([]chan *pb.BroadcastRoomReq, c.RoutineAmount) //创建一个长度为32数组 每个元素都是用来推送到指定房间的结构
	for i := uint64(0); i < c.RoutineAmount; i++ {
		ch := make(chan *pb.BroadcastRoomReq, c.RoutineSize) //创建一个通道 通道长度为1024
		b.routines[i] = ch
		go b.roomproc(ch) //32个通道,每个通道用个协程监听
	}
	return
}

// ChannelCount 获取这个桶中有多少个 Channel 默认1024长度
func (b *Bucket) ChannelCount() int {
	return len(b.chs)
}

// RoomCount 获取这个桶中有多少个 Room 默认1024
func (b *Bucket) RoomCount() int {
	return len(b.rooms)
}

// RoomsCount 获取房间多少在线
func (b *Bucket) RoomsCount() (res map[string]int32) {
	var (
		roomID string
		room   *Room
	)
	b.cLock.RLock()
	res = make(map[string]int32)
	for roomID, room = range b.rooms {
		if room.Online > 0 {
			res[roomID] = room.Online
		}
	}
	b.cLock.RUnlock()
	return
}

// ChangeRoom change ro room
func (b *Bucket) ChangeRoom(nrid string, ch *Channel) (err error) {
	var (
		nroom *Room
		ok    bool
		oroom = ch.Room
	)
	// change to no room
	if nrid == "" {
		if oroom != nil && oroom.Del(ch) {
			b.DelRoom(oroom)
		}
		ch.Room = nil
		return
	}
	b.cLock.Lock()
	if nroom, ok = b.rooms[nrid]; !ok {
		nroom = NewRoom(nrid)
		b.rooms[nrid] = nroom
	}
	b.cLock.Unlock()
	if oroom != nil && oroom.Del(ch) {
		b.DelRoom(oroom)
	}

	if err = nroom.Put(ch); err != nil {
		return
	}
	ch.Room = nroom
	return
}

// Put put a channel according with sub key.
func (b *Bucket) Put(rid string, ch *Channel) (err error) {
	var (
		room *Room
		ok   bool
	)
	b.cLock.Lock()
	// close old channel
	if dch := b.chs[ch.Key]; dch != nil {
		//ch.Key 是登录接口返回的key值  先hash%32得到idx 再拿到对应*Bucket指针 s.buckets[idx]
		//*Bucket指针 b.chs map默认长度1024
		// 如果在map chs能找到对应的 这个设备id则调用value指针.Close()方法 即: c.signal <- &Proto{Op: OpProtoFinish}
		dch.Close()
	}
	//将Bucket.chs 加入uuid为key Channel结构为val 的数据进入map
	b.chs[ch.Key] = ch
	if rid != "" {
		//roomid不再Bucket.rooms这个map里 则新建个Room结构写入map
		if room, ok = b.rooms[rid]; !ok {
			room = NewRoom(rid)
			b.rooms[rid] = room
		}
		ch.Room = room
	}
	b.ipCnts[ch.IP]++
	b.cLock.Unlock()
	if room != nil {
		err = room.Put(ch) //room 结构里也加入channel
	}
	return
}

// Del delete the channel by sub key.
func (b *Bucket) Del(dch *Channel) {
	room := dch.Room
	b.cLock.Lock()
	if ch, ok := b.chs[dch.Key]; ok {
		if ch == dch {
			delete(b.chs, ch.Key)
		}
		// ip counter
		if b.ipCnts[ch.IP] > 1 {
			b.ipCnts[ch.IP]--
		} else {
			delete(b.ipCnts, ch.IP)
		}
	}
	b.cLock.Unlock()
	if room != nil && room.Del(dch) {
		// if empty room, must delete from bucket
		b.DelRoom(room)
	}
}

// Channel get a channel by sub key.
func (b *Bucket) Channel(key string) (ch *Channel) {
	b.cLock.RLock()
	ch = b.chs[key]
	b.cLock.RUnlock()
	return
}

// Broadcast
func (b *Bucket) Broadcast(p *protocol.Proto, op int32) {
	var ch *Channel
	b.cLock.RLock()
	for _, ch = range b.chs {
		// 推送至该用户 登录接口返回的 accept 房间号数组
		if !ch.NeedPush(op) {
			continue
		}
		_ = ch.Push(p)
	}
	b.cLock.RUnlock()
}

// Room 用roomid获取对应Room结构
func (b *Bucket) Room(rid string) (room *Room) {
	b.cLock.RLock()
	room = b.rooms[rid]
	b.cLock.RUnlock()
	return
}

// DelRoom 用roomid删除对应Room结构
func (b *Bucket) DelRoom(room *Room) {
	b.cLock.Lock()
	delete(b.rooms, room.ID)
	b.cLock.Unlock()
	room.Close()
}

// BroadcastRoom 广播一条消息到指定房间
func (b *Bucket) BroadcastRoom(arg *pb.BroadcastRoomReq) {
	num := atomic.AddUint64(&b.routinesNum, 1) % b.c.RoutineAmount
	b.routines[num] <- arg
}

// Rooms 获取所有房间的 roomID.     RoomsCount 获取每个房间在线多少人
func (b *Bucket) Rooms() (res map[string]struct{}) {
	var (
		roomID string
		room   *Room
	)
	res = make(map[string]struct{})
	b.cLock.RLock()
	for roomID, room = range b.rooms {
		if room.Online > 0 {
			res[roomID] = struct{}{}
		}
	}
	b.cLock.RUnlock()
	return
}

// IPCount get ip count.
func (b *Bucket) IPCount() (res map[string]struct{}) {
	var (
		ip string
	)
	b.cLock.RLock()
	res = make(map[string]struct{}, len(b.ipCnts))
	for ip = range b.ipCnts {
		res[ip] = struct{}{}
	}
	b.cLock.RUnlock()
	return
}

// UpRoomsCount update all room count
func (b *Bucket) UpRoomsCount(roomCountMap map[string]int32) {
	var (
		roomID string
		room   *Room
	)
	b.cLock.RLock()
	for roomID, room = range b.rooms {
		room.AllOnline = roomCountMap[roomID]
	}
	b.cLock.RUnlock()
}

// roomproc
func (b *Bucket) roomproc(c chan *pb.BroadcastRoomReq) {
	for {
		arg := <-c //等待c chan *grpc.BroadcastRoomReq这个通道是否有数据过来
		//一旦将拿到通过arg.RoomID 去 b.rooms下对应的*Room结构  并执行*Room结构的Push方法
		if room := b.Room(arg.RoomID); room != nil {
			room.Push(arg.Proto)
		}
	}
}
