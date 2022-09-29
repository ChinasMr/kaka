package server

import (
	"github.com/ChinasMr/kaka/internal/conf"
	"github.com/ChinasMr/kaka/pkg/transport/rtsp"
)

func NewRTSPServer(c *conf.Server) *rtsp.Server {
	var opts []rtsp.ServerOption
	if c.Stream.Rtsp != "" {
		opts = append(opts, rtsp.RTSP(c.Stream.Rtsp))
	}
	if c.Stream.Rtp != "" {
		opts = append(opts, rtsp.RTP(c.Stream.Rtp))
	}
	if c.Stream.Rtcp != "" {
		opts = append(opts, rtsp.RTCP(c.Stream.Rtcp))
	}
	srv := rtsp.NewServer(opts...)
	return srv
}
