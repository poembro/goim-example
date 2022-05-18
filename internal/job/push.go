package job

import (
	"context"
	"fmt"

	"goim-demo/api/comet"
	pb "goim-demo/api/logic"
	"goim-demo/api/protocol"
	"goim-demo/pkg/bytes"

	log "github.com/golang/glog"
)

//Subscribe message: operation:1001 server:"127.0.0.1" keys:"123456123" msg:"11111111111111111111111"
func (j *Job) push(ctx context.Context, pushMsg *pb.PushMsg) (err error) {
	switch pushMsg.Type {
	case pb.PushMsg_PUSH:
		err = j.pushKeys(pushMsg.Operation, pushMsg.Server, pushMsg.Keys, pushMsg.Msg)
	case pb.PushMsg_ROOM:
		err = j.getRoom(pushMsg.Room).Push(pushMsg.Operation, pushMsg.Msg)
	case pb.PushMsg_BROADCAST:
		err = j.broadcast(pushMsg.Operation, pushMsg.Msg, pushMsg.Speed)
	default:
		err = fmt.Errorf("no match push type: %s", pushMsg.Type)
	}
	return
}

// pushKeys push a message to a batch of subkeys.
func (j *Job) pushKeys(operation int32, serverID string, subKeys []string, body []byte) (err error) {
	buf := bytes.NewWriterSize(len(body) + 64)
	p := &protocol.Proto{
		Ver:  1,
		Op:   operation,
		Body: body,
	}
	p.WriteTo(buf)
	p.Body = buf.Buffer()
	p.Op = protocol.OpRaw
	var args = comet.PushMsgReq{
		Keys:    subKeys,
		ProtoOp: operation,
		Proto:   p,
	}

	comets := j.cometServers
	for grpcAddr, c := range comets {
		//host, _, _ := net.SplitHostPort(grpcAddr)
		if grpcAddr != serverID {
			continue // 用户再某节点,某端口 则请求某节点,某端口
		}

		if err := c.Push(&args); err != nil {
			log.Errorf("c.Push(%v) serverID:%s error(%v)", args, serverID, err)
		}
		log.Infof("pushKey:%s comets:%d", serverID, len(j.cometServers))
	}

	return
}

// broadcast broadcast a message to all.
func (j *Job) broadcast(operation int32, body []byte, speed int32) (err error) {
	buf := bytes.NewWriterSize(len(body) + 64)
	p := &protocol.Proto{
		Ver:  1,
		Op:   operation,
		Body: body,
	}
	p.WriteTo(buf)
	p.Body = buf.Buffer()
	p.Op = protocol.OpRaw // 9
	comets := j.cometServers
	if len(comets) <= 0 {
		log.Infof("暂无comet节点在线:%d", len(comets))
		return
	}
	speed /= int32(len(comets))
	var args = comet.BroadcastReq{
		ProtoOp: operation,
		Proto:   p,
		Speed:   speed,
	}
	for serverID, c := range comets {
		if err = c.Broadcast(&args); err != nil {
			log.Errorf("c.Broadcast(%v) serverID:%s error(%v)", args, serverID, err)
		}
	}
	log.Infof("broadcast comets:%d", len(comets))
	return
}

// broadcastRoomRawBytes broadcast aggregation messages to room.
func (j *Job) broadcastRoomRawBytes(roomID string, body []byte) (err error) {
	args := comet.BroadcastRoomReq{
		RoomID: roomID,
		Proto: &protocol.Proto{
			Ver:  1,
			Op:   protocol.OpRaw,
			Body: body,
		},
	}
	comets := j.cometServers
	for serverID, c := range comets {
		if err = c.BroadcastRoom(&args); err != nil {
			log.Errorf("c.BroadcastRoom(%v) roomID:%s serverID:%s error(%v)", args, roomID, serverID, err)
		}
	}
	log.Infof("broadcastRoom comets:%d", len(comets))
	return
}
