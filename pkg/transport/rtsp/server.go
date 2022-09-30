package rtsp

import (
	"bufio"
	"context"
	"encoding/binary"
	"github.com/ChinasMr/kaka/pkg/log"
	"github.com/ChinasMr/kaka/pkg/transport/rtsp/method"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	network string
	address string
	lis     net.Listener
	err     error
	timeout time.Duration
	baseCtx context.Context

	log     *log.Helper
	handler Handler
	conns   map[ServerTransport]bool
	serveWG sync.WaitGroup
	mutex   sync.Mutex
}

func NewServer(opts ...ServerOption) *Server {
	srv := &Server{
		network: "tcp",
		address: ":0",
		mutex:   sync.Mutex{},
		handler: &UnimplementedServerHandler{},
	}
	for _, o := range opts {
		o(srv)
	}

	return srv
}

func (s *Server) serveRTP(conn net.Conn) {
	log.Debugf("-------start recording ....")
	packet := make([]byte, 2048)
	bf := bufio.NewReader(conn)
	for {
		// sender report
		bs, err := bf.Peek(4)
		if err != nil {
			return
		}
		_, _ = bf.Discard(4)
		if bs[0] != 0x24 || bs[1] != 0x00 {
			return
		}

		pl := binary.BigEndian.Uint32(bs[2:])
		if pl > 2048 {
			return
		}
		_, err = io.ReadFull(bf, packet[:pl])
		if err != nil {
			return
		}
		// todo forward to other client.

	}
}

func (s *Server) serveStream(trans ServerTransport) {
	for {
		request, err := trans.Request()
		if err != nil {
			if err == io.EOF {
				continue
			}
			s.log.Errorf("can not parse rtsp request: %v", err)
			return
		}
		var (
			response = NewResponse(request.proto, request.cSeq)
			err1     error
		)

		switch request.Method() {
		case method.OPTIONS:
			err1 = s.handler.OPTIONS(request, response)
		case method.DESCRIBE:
			err1 = s.handler.DESCRIBE(request, response)
		case method.SETUP:
			err1 = s.handler.SETUP(request, response)
		case method.ANNOUNCE:
			err1 = s.handler.ANNOUNCE(request, response)
		case method.RECORD:
			err1 = s.handler.RECORD(request, response)
			err1 = trans.Response(response)
			if err1 != nil {
				s.log.Errorf("can not response to %s: %v", trans.Addr(), err1)
			}
			s.serveRTP(trans.RawConn())
			return
		default:
			s.log.Errorf("unknown method: %s", request.Method())
			continue
		}
		if err1 != nil {
			s.log.Errorf("can not serve: %v", err)
			continue
		}

		err1 = trans.Response(response)
		if err1 != nil {
			s.log.Errorf("can not response to %s: %v", trans.Addr(), err1)
		}
	}
}

func (s *Server) handleRawConn(conn net.Conn) {
	// build transport
	nt := s.newRTSPTransport(conn)
	// add
	go func() {
		s.serveStream(nt)
		// remove
	}()
}

func (s *Server) newRTSPTransport(c net.Conn) ServerTransport {
	return NewTransport(c)
}

func (s *Server) serve(lis net.Listener) error {
	for {
		rawConn, err := lis.Accept()
		if err != nil {
			return nil
		}
		s.log.Debugf("new connection created from: %v", rawConn.RemoteAddr().String())
		s.serveWG.Add(1)
		go func() {
			s.handleRawConn(rawConn)
			s.serveWG.Done()
		}()
	}

}

func (s *Server) Start(ctx context.Context) error {
	err := s.listen()
	if err != nil {
		return err
	}
	s.baseCtx = ctx
	log.Infof("[RTSP] server listening on: %s", s.lis.Addr().String())
	return s.serve(s.lis)
}

func (s *Server) Stop(_ context.Context) error {
	s.GracefulStop()
	log.Info("[gRPC] server stopping")
	return nil
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

func (s *Server) RegisterHandler(handler Handler) {
	if handler == nil {
		return
	}
	s.handler = handler
}

func (s *Server) GracefulStop() {
	_ = s.lis.Close()
}
