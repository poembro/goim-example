package grpc

import (
	"context"
	"net"
	"time"

	pb "goim-demo/api/comet/grpc"
	"goim-demo/internal/comet"
	"goim-demo/internal/comet/conf"
	"goim-demo/internal/comet/errors"

	log "github.com/golang/glog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	"google.golang.org/grpc/metadata" // grpc metadata包
)

// interceptor 拦截器
func interceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Infoln("---->", "无Token认证信息")
	}

	log.Infoln("--comet gateway ",
		"    method:", info.FullMethod,
		"    md:", md,
		"    req:", req)

	// 继续处理请求
	return handler(ctx, req)
}

// New comet grpc server.
func New(c *conf.RPCServer, s *comet.Server) *grpc.Server {
	keepParams := grpc.KeepaliveParams(keepalive.ServerParameters{
		MaxConnectionIdle:     time.Duration(c.IdleTimeout),       //60s 连接最大闲置时间
		MaxConnectionAgeGrace: time.Duration(c.ForceCloseWait),    //20s 连接最大闲置时间
		Time:                  time.Duration(c.KeepAliveInterval), //60s
		Timeout:               time.Duration(c.KeepAliveTimeout),  //20s
		MaxConnectionAge:      time.Duration(c.MaxLifeTime),       //2h  小时
	})
	srv := grpc.NewServer(grpc.UnaryInterceptor(interceptor), keepParams)
	pb.RegisterCometServer(srv, &server{s})
	lis, err := net.Listen(c.Network, c.Addr) // "tcp", ":3109"
	if err != nil {
		panic(err)
	}
	go func() {
		if err := srv.Serve(lis); err != nil {
			panic(err)
		}
	}()
	return srv
}

type server struct {
	srv *comet.Server
}

var _ pb.CometServer = &server{}

// Ping Service
func (s *server) Ping(ctx context.Context, req *pb.Empty) (*pb.Empty, error) {
	log.Infoln("Ping:")
	return &pb.Empty{}, nil
}

// Close Service
func (s *server) Close(ctx context.Context, req *pb.Empty) (*pb.Empty, error) {
	log.Infoln("Close:")
	// TODO: some graceful close
	return &pb.Empty{}, nil
}

// PushMsg push a message to specified sub keys.
func (s *server) PushMsg(ctx context.Context, req *pb.PushMsgReq) (reply *pb.PushMsgReply, err error) {
	log.Infoln("PushMsg:", req)
	if len(req.Keys) == 0 || req.Proto == nil {
		return nil, errors.ErrPushMsgArg
	}
	for _, key := range req.Keys {
		if channel := s.srv.Bucket(key).Channel(key); channel != nil {
			if !channel.NeedPush(req.ProtoOp) {
				continue
			}
			if err = channel.Push(req.Proto); err != nil {
				return
			}
		}
	}
	return &pb.PushMsgReply{}, nil
}

// Broadcast broadcast msg to all user. /* 推送至 登录接口返回的 accept 房间号 */
func (s *server) Broadcast(ctx context.Context, req *pb.BroadcastReq) (*pb.BroadcastReply, error) {
	log.Infoln("Broadcast:", req)
	if req.Proto == nil {
		return nil, errors.ErrBroadCastArg
	}
	// TODO use broadcast queue
	go func() {
		for _, bucket := range s.srv.Buckets() {
			bucket.Broadcast(req.GetProto(), req.ProtoOp)
			if req.Speed > 0 {
				t := bucket.ChannelCount() / int(req.Speed)
				time.Sleep(time.Duration(t) * time.Second)
			}
		}
	}()
	return &pb.BroadcastReply{}, nil
}

// BroadcastRoom broadcast msg to specified room.  /* 推送至 登录接口返回的 room 房间号 即登录接口必须返回room_id */
func (s *server) BroadcastRoom(ctx context.Context, req *pb.BroadcastRoomReq) (*pb.BroadcastRoomReply, error) {
	log.Infoln("BroadcastRoom:", req)
	if req.Proto == nil || req.RoomID == "" {
		return nil, errors.ErrBroadCastRoomArg
	}
	for _, bucket := range s.srv.Buckets() {
		bucket.BroadcastRoom(req)
	}
	return &pb.BroadcastRoomReply{}, nil
}

// Rooms gets all the room ids for the server. 获取服务器的所有房间id。
func (s *server) Rooms(ctx context.Context, req *pb.RoomsReq) (*pb.RoomsReply, error) {
	log.Infoln("Rooms:", req)
	var (
		roomIds = make(map[string]bool)
	)
	for _, bucket := range s.srv.Buckets() {
		for roomID := range bucket.Rooms() {
			roomIds[roomID] = true
		}
	}
	return &pb.RoomsReply{Rooms: roomIds}, nil
}
