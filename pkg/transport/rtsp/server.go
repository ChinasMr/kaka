package rtsp

import "context"

type Server struct {
	rtspPort string
	rtpPort  string
	rtcpPort string
}

func (s *Server) Start(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

func (s *Server) Stop(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

func NewServer(opts ...ServerOption) *Server {
	srv := &Server{
		rtspPort: ":8554",
		rtpPort:  ":8000",
		rtcpPort: ":8001",
	}
	for _, o := range opts {
		o(srv)
	}

	return srv
}
