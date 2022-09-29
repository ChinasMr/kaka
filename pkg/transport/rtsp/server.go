package rtsp

import (
	"context"
	"sync"
)

type Server struct {
	rtspPort string
	rtpPort  string
	rtcpPort string

	mutex sync.Mutex
}

func (s *Server) Start(ctx context.Context) error {
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	return nil
}

func NewServer(opts ...ServerOption) *Server {
	srv := &Server{
		rtspPort: ":8554",
		rtpPort:  ":8000",
		rtcpPort: ":8001",
		mutex:    sync.Mutex{},
	}
	for _, o := range opts {
		o(srv)
	}

	return srv
}
