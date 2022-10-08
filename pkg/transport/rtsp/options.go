package rtsp

import (
	"github.com/ChinasMr/kaka/pkg/log"
	"time"
)

type ServerOption func(s *Server)

func Address(addr string) ServerOption {
	return func(s *Server) {
		s.address = addr
	}
}

func Logger(logger log.Logger) ServerOption {
	return func(s *Server) {
		s.log = log.NewHelper(logger)
	}
}

func Network(network string) ServerOption {
	return func(s *Server) {
		s.network = network
	}
}

func Timeout(duration time.Duration) ServerOption {
	return func(s *Server) {
		s.timeout = duration
	}
}

func RTP(rtp string) ServerOption {
	return func(s *Server) {
		s.rtp = rtp
	}
}

func RTCP(rtcp string) ServerOption {
	return func(s *Server) {
		s.rtcp = rtcp
	}
}
