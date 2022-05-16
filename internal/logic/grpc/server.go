package grpc

import (
	"context"
	"net"
	"time"

	pb "goim-demo/api/logic/grpc"
	"goim-demo/internal/logic"
	"goim-demo/internal/logic/conf"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	// use gzip decoder
	log "github.com/golang/glog"
	_ "google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/metadata" // grpc metadata包
)

// interceptor 拦截器
func interceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Infoln("---->", "无Token认证信息")
	}

	log.Infoln("--logic gateway ",
		"    method:", info.FullMethod,
		"    md:", md,
		"    req:", req)
	// 继续处理请求
	return handler(ctx, req)
}

// New logic grpc server
func New(c *conf.RPCServer, l *logic.Logic) *grpc.Server {
	keepParams := grpc.KeepaliveParams(keepalive.ServerParameters{
		MaxConnectionIdle:     time.Duration(c.IdleTimeout),
		MaxConnectionAgeGrace: time.Duration(c.ForceCloseWait),
		Time:                  time.Duration(c.KeepAliveInterval),
		Timeout:               time.Duration(c.KeepAliveTimeout),
		MaxConnectionAge:      time.Duration(c.MaxLifeTime),
	})
	srv := grpc.NewServer(grpc.UnaryInterceptor(interceptor), keepParams)
	pb.RegisterLogicServer(srv, &server{l})
	lis, err := net.Listen(c.Network, c.Addr)
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
	srv *logic.Logic
}

var _ pb.LogicServer = &server{}

// Ping Service
func (s *server) Ping(ctx context.Context, req *pb.PingReq) (*pb.PingReply, error) {
	return &pb.PingReply{}, nil
}

// Close Service
func (s *server) Close(ctx context.Context, req *pb.CloseReq) (*pb.CloseReply, error) {
	return &pb.CloseReply{}, nil
}

// Connect connect a conn.
func (s *server) Connect(ctx context.Context, req *pb.ConnectReq) (*pb.ConnectReply, error) {
	mid, key, room, accepts, hb, err := s.srv.Connect(ctx, req.Server, req.Cookie, req.Token)
	if err != nil {
		return &pb.ConnectReply{}, err
	}
	return &pb.ConnectReply{Mid: mid, Key: key, RoomID: room, Accepts: accepts, Heartbeat: hb}, nil
}

// Disconnect disconnect a conn.
func (s *server) Disconnect(ctx context.Context, req *pb.DisconnectReq) (*pb.DisconnectReply, error) {
	has, err := s.srv.Disconnect(ctx, req.Mid, req.Key, req.Server)
	if err != nil {
		return &pb.DisconnectReply{}, err
	}
	return &pb.DisconnectReply{Has: has}, nil
}

// Heartbeat beartbeat a conn.
func (s *server) Heartbeat(ctx context.Context, req *pb.HeartbeatReq) (*pb.HeartbeatReply, error) {
	if err := s.srv.Heartbeat(ctx, req.Mid, req.Key, req.Server); err != nil {
		return &pb.HeartbeatReply{}, err
	}
	return &pb.HeartbeatReply{}, nil
}

// RenewOnline 将对应comet服务写入redis "HSET" "ol_192.168.3.222" "43" "{\"server\":\"192.168.3.222\",\"room_count\":{\"live://1000\":1},\"updated\":1577077540}"
func (s *server) RenewOnline(ctx context.Context, req *pb.OnlineReq) (*pb.OnlineReply, error) {
	allRoomCount, err := s.srv.RenewOnline(ctx, req.Server, req.RoomCount)
	if err != nil {
		return &pb.OnlineReply{}, err
	}

	return &pb.OnlineReply{AllRoomCount: allRoomCount}, nil
}

// Receive receive a message.
func (s *server) Receive(ctx context.Context, req *pb.ReceiveReq) (*pb.ReceiveReply, error) {
	if err := s.srv.Receive(ctx, req.Mid, req.Proto); err != nil {
		return &pb.ReceiveReply{}, err
	}
	return &pb.ReceiveReply{}, nil
}

// nodes return nodes.  节点的权重列表
func (s *server) Nodes(ctx context.Context, req *pb.NodesReq) (*pb.NodesReply, error) {
	return &pb.NodesReply{}, nil
}
