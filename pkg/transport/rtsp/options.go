package rtsp

import "github.com/ChinasMr/kaka/pkg/log"

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
