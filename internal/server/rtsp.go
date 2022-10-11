package server

import (
	"github.com/ChinasMr/kaka/internal/conf"
	"github.com/ChinasMr/kaka/internal/service"
	"github.com/ChinasMr/kaka/pkg/log"
	"github.com/ChinasMr/kaka/pkg/transport/rtsp"
)

func NewRTSPServer(c *conf.Server, kaka *service.KakaService, logger log.Logger) *rtsp.Server {
	var opts = []rtsp.ServerOption{
		rtsp.Channel("live"),
		rtsp.Logger(logger),
	}
	if c.Rtsp.Addr != "" {
		opts = append(opts, rtsp.Address(c.Rtsp.Addr))
	}
	if c.Rtsp.Network != "" {
		opts = append(opts, rtsp.Network(c.Rtsp.Network))
	}
	if c.Rtsp.Rtp != "" {
		opts = append(opts, rtsp.RTP(c.Rtsp.Rtp))
	}
	if c.Rtsp.Rtcp != "" {
		opts = append(opts, rtsp.RTCP(c.Rtsp.Rtcp))
	}
	if c.Rtsp.Timeout != nil {
		opts = append(opts, rtsp.Timeout(c.Rtsp.Timeout.AsDuration()))
	}

	srv := rtsp.NewServer(opts...)
	//srv.RegisterHandler(methods.ANNOUNCE, kaka.ANNOUNCE)
	return srv
}
