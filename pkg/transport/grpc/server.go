package grpc

import (
	"context"
	"crypto/tls"
	"github.com/ChinasMr/kaka/pkg/log"
	"github.com/ChinasMr/kaka/pkg/matcher"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"net"
	"time"
)

type Server struct {
	*grpc.Server
	baseCtx    context.Context
	tlsConf    *tls.Config
	lis        net.Listener
	err        error
	network    string
	address    string
	timeout    time.Duration
	middleware matcher.Matcher

	unaryInterceptors  []grpc.UnaryServerInterceptor
	streamInterceptors []grpc.StreamServerInterceptor
	grpcOpts           []grpc.ServerOption
	health             *health.Server
}

func (s *Server) Start(ctx context.Context) error {
	err := s.listen()
	if err != nil {
		return s.err
	}
	s.baseCtx = ctx
	log.Infof("[gRPC] server listening on: %s", s.lis.Addr().String())
	s.health.Resume()
	return s.Serve(s.lis)
}

func (s *Server) Stop(ctx context.Context) error {
	s.health.Shutdown()
	s.GracefulStop()
	log.Info("[gRPC] server stopping")
	return nil
}

func NewServer(opts ...ServerOption) *Server {
	srv := &Server{
		baseCtx:    context.Background(),
		network:    "tcp",
		address:    ":0",
		timeout:    1 * time.Second,
		health:     health.NewServer(),
		middleware: matcher.New(),
	}
	for _, o := range opts {
		o(srv)
	}
	// todo there losing something.
	grpcOpts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(srv.unaryInterceptors...),
		grpc.ChainStreamInterceptor(srv.streamInterceptors...),
	}
	if srv.tlsConf != nil {
		grpcOpts = append(grpcOpts, grpc.Creds(credentials.NewTLS(srv.tlsConf)))
	}
	if len(srv.grpcOpts) > 0 {
		grpcOpts = append(grpcOpts, srv.grpcOpts...)
	}

	srv.Server = grpc.NewServer(grpcOpts...)

	grpc_health_v1.RegisterHealthServer(srv.Server, srv.health)
	reflection.Register(srv.Server)

	return srv
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
