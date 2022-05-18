package service

import (
	"context"
	"goim-demo/pkg/grpclib"
	"goim-demo/pkg/rpc"
	"goim-demo/pkg/util"

	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc/metadata"

	"goim-demo/pkg/logger"
	"goim-demo/pkg/pb"

	"goim-demo/pkg/gerrors"
)

// SendOne 一对一消息发送
func (s *Service) SendOne(ctx context.Context, msg *pb.PushMsgReq) error {
	requestId := grpclib.GetCtxRequestIdStr(ctx)
	userId, deviceId, err := grpclib.GetCtxDataStr(ctx)
	if err != nil {
		logger.Sugar.Infow("logic 服务 SendMessage 头信息error")
		return err
	}
	// 加上grpc头防止api授权拦截
	metaCtx := metadata.NewOutgoingContext(context.TODO(), metadata.Pairs(
		"user_id", userId,
		"device_id", deviceId,
		"token", "TODO token verify",
		"request_id", requestId))

	rpc.ConnectInt(msg.Message.Server).DeliverMessage(metaCtx, msg)

	return nil
}

// SendRoom 群组消息发送
func (s *Service) SendRoom(ctx context.Context, msg *pb.PushMsg) error {
	bytes, err := proto.Marshal(msg)
	if err != nil {
		return gerrors.WrapError(err)
	}
	err = s.dao.Publish(s.c.Global.PushAllTopic, bytes)
	if err != nil {
		return err
	}

	return nil
}

// Sync 消息同步
func (s *Service) Sync(ctx context.Context, roomId string, seq int64) (*pb.SyncResp, error) {
	dst, err := s.dao.GetMessageList(roomId, 0, 50) // 取回最近50条消息
	if err != nil {
		return nil, gerrors.WrapError(err)
	}
	max := len(dst)
	//sort.Sort(sort.Reverse(sort.StringSlice(dst))) //倒序失败

	jsonStr := "["
	for i := max - 1; i >= 0; i-- {
		jsonStr += dst[i]
		if i == 0 {
			continue
		}
		jsonStr += ","
	}
	jsonStr = jsonStr + "]"

	resp := &pb.SyncResp{Messages: util.S2B(jsonStr), HasMore: false}
	return resp, nil
}

// MessageACK 消息确认机制
func (s *Service) MessageACK(ctx context.Context, deviceId, roomId string, userId, deviceAck, receiveTime int64) error {
	s.dao.AddMessageACKMapping(deviceId, roomId, deviceAck)
	return nil
}

// GetMessageCount 统计未读
func (s *Service) GetMessageCount(roomId, start, stop string) (int64, error) {
	return s.dao.GetMessageCount(roomId, start, stop)
}

// GetMessageList 取回消息
func (s *Service) GetMessageList(roomId string, start, stop int64) ([]string, error) {
	return s.dao.GetMessageList(roomId, start, stop)
}

// GetMessagePageList 取回消息分页
func (s *Service) GetMessagePageList(roomId, min, max string, page, limit int64) ([]string, int64, error) {
	return s.dao.GetMessagePageList(roomId, min, max, page, limit)
}

// AddMessageList 将消息添加到对应房间 roomId.
func (s *Service) AddMessageList(roomId string, id int64, msg string) error {
	return s.dao.AddMessageList(roomId, id, msg)
}
