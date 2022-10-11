package http

import (
	"context"
	"errors"
	"github.com/ChinasMr/kaka/pkg/log"
	"github.com/gorilla/mux"
	"net"
	"net/http"
	"time"
)

type Server struct {
	*http.Server
	lis         net.Listener
	err         error
	network     string
	address     string
	timeout     time.Duration
	router      *mux.Router
	strictSlash bool
}

func NewServer(opts ...ServerOption) *Server {
	srv := &Server{
		network:     "tcp",
		address:     ":0",
		timeout:     1 * time.Second,
		strictSlash: true,
	}
	for _, opt := range opts {
		opt(srv)
	}
	srv.router = mux.NewRouter().StrictSlash(srv.strictSlash)
	srv.router.NotFoundHandler = http.DefaultServeMux
	srv.router.MethodNotAllowedHandler = http.DefaultServeMux
	srv.Server = &http.Server{
		Handler: srv.router,
	}
	return srv
}

func (s *Server) Start(ctx context.Context) error {
	if err := s.listen(); err != nil {
		return err
	}
	s.BaseContext = func(listener net.Listener) context.Context {
		return ctx
	}
	log.Infof("[HTTP] server listening on: %s", s.lis.Addr().String())
	err := s.Serve(s.lis)
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	log.Info("[HTTP] server stopping")
	return s.Shutdown(ctx)
}

func (s *Server) listen() error {
	if s.lis == nil {
		lis, err := net.Listen(s.network, s.address)
		if err != nil {
			s.err = err
			return err
		}
		s.lis = lis
	}
	return s.err
}
