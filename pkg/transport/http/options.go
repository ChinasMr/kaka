package http

import (
	"time"
)

type ServerOption func(s *Server)

func Network(network string) ServerOption {
	return func(s *Server) {
		s.network = network
	}
}

func Address(address string) ServerOption {
	return func(s *Server) {
		s.address = address
	}
}

func Timeout(duration time.Duration) ServerOption {
	return func(s *Server) {
		s.timeout = duration
	}
}
