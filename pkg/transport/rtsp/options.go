package rtsp

type ServerOption func(s *Server)

func RTSP(addr string) ServerOption {
	return func(s *Server) {
		s.rtspPort = addr
	}
}

func RTP(addr string) ServerOption {
	return func(s *Server) {
		s.rtpPort = addr
	}
}

func RTCP(addr string) ServerOption {
	return func(s *Server) {
		s.rtcpPort = addr
	}
}
