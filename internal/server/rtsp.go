package server

import (
	"github.com/ChinasMr/kaka/internal/conf"
	"github.com/ChinasMr/kaka/pkg/log"
	"github.com/ChinasMr/kaka/pkg/transport/rtsp"
)

func NewRTSPServer(c *conf.Server, logger log.Logger) *rtsp.Server {
	var opts = []rtsp.ServerOption{
		rtsp.Logger(logger),
	}
	if c.Grpc.Addr != "" {
		opts = append(opts, rtsp.Address(c.Rtsp.Addr))
	}
	srv := rtsp.NewServer(opts...)
	return srv
}
