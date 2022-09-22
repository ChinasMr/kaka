package server

import (
	v1 "github.com/ChinasMr/kaka/api/kaka/v1"
	"github.com/ChinasMr/kaka/internal/conf"
	"github.com/ChinasMr/kaka/internal/service"
	"github.com/ChinasMr/kaka/pkg/transport/grpc"
)

func NewGRPCServer(c *conf.Server, kaka *service.KakaService) *grpc.Server {
	var opts = []grpc.ServerOption{
		grpc.Middleware(),
	}
	if c.Grpc.Network != "" {
		opts = append(opts, grpc.NetWork(c.Grpc.Network))
	}
	if c.Grpc.Addr != "" {
		opts = append(opts, grpc.Address(c.Grpc.Addr))
	}
	if c.Grpc.Timeout != nil {
		opts = append(opts, grpc.Timeout(c.Grpc.Timeout.AsDuration()))
	}
	srv := grpc.NewServer(opts...)
	v1.RegisterKakaServer(srv, kaka)
	return srv
}
