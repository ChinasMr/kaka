package rtsp

import (
	"context"
	"fmt"
	"github.com/ChinasMr/kaka/pkg/log"
	"github.com/ChinasMr/kaka/pkg/transport/rtsp/method"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	network  string
	address  string
	rtp      string
	rtcp     string
	lis      net.Listener
	conn     *net.UDPConn
	rtpConn  *net.UDPConn
	rtcpConn *net.UDPConn
	err      error
	timeout  time.Duration
	baseCtx  context.Context
	log      *log.Helper
	handler  Handler
	serveWG  sync.WaitGroup
	mutex    sync.Mutex
}

func NewServer(opts ...ServerOption) *Server {
	srv := &Server{
		network: "*",
		address: ":0",
		rtp:     "30000",
		rtcp:    ":30001",
		mutex:   sync.Mutex{},
		handler: &UnimplementedServerHandler{},
	}
	for _, o := range opts {
		o(srv)
	}

	return srv
}

func (s *Server) serveStream(trans *transport) {
	for {
		req, err := trans.parseRequest()
		if err != nil {
			if err == io.EOF {
				continue
			}
			s.log.Errorf("can not parse rtsp request: %v", err)
			return
		}
		var (
			res  = NewResponse(req.proto, req.cSeq)
			err1 error
		)

		switch req.Method() {
		case method.OPTIONS:
			err1 = s.handler.OPTIONS(req, res, trans)
		case method.DESCRIBE:
			err1 = s.handler.DESCRIBE(req, res, trans)
		case method.SETUP:
			err1 = s.handler.SETUP(req, res, trans)
		case method.ANNOUNCE:
			err1 = s.handler.ANNOUNCE(req, res, trans)
		case method.RECORD:
			err1 = s.handler.RECORD(req, res, trans)
			if err1 != nil {
				s.log.Errorf("can not record: %v", err1)
			}
			return
		case method.PLAY:
			err1 = s.handler.PLAY(req, res, trans)
			if err1 != nil {
				s.log.Errorf("can not play: %v", err1)
			}
			return
		case method.TEARDOWN:
			_ = s.handler.TEARDOWN(req, res, trans)
			return
		default:
			err1 = fmt.Errorf("unknown method: %s", req.Method())
		}
		if err1 != nil {
			s.log.Errorf("can not serve: %v", err1)
			errRes := NewResponse(req.proto, req.cSeq)
			errRes.statusCode = 400
			errRes.status = "Bad Request"
			_ = trans.sendResponse(errRes)
			return
		}

		err1 = trans.sendResponse(res)
		if err1 != nil {
			s.log.Errorf("can not response to %s: %v", trans.Addr(), err1)
			return
		}
	}
}

func (s *Server) handleRawConn(conn net.Conn) {
	// build transport
	nt := NewTransport(conn)
	s.serveStream(nt)
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
			_ = rawConn.Close()
			s.log.Debugf("connection closed to: %v", rawConn.RemoteAddr().String())
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
	log.Infof("[RTP ] server listening on: %s", s.rtpConn.LocalAddr())
	log.Infof("[RTCP] server listening on: %s", s.rtcpConn.LocalAddr())
	return s.serve(s.lis)
}

func (s *Server) Stop(_ context.Context) error {
	s.GracefulStop()
	log.Info("[gRPC] server stopping")
	return nil
}

func (s *Server) listen() error {
	if s.lis == nil && (s.network == "*" || s.network == "tcp") {
		lis, err := net.Listen("tcp", s.address)
		if err != nil {
			s.err = err
			return err
		}
		s.lis = lis
	}

	if s.conn == nil && (s.network == "*" || s.network == "udp") {
		addr, err := net.ResolveUDPAddr("udp", s.address)
		if err != nil {
			s.err = err
			return err
		}
		conn, err := net.ListenUDP("udp", addr)
		if err != nil {
			s.err = err
			return err
		}
		s.conn = conn
	}

	addr, err := net.ResolveUDPAddr("udp", s.rtp)
	if err != nil {
		s.err = err
		return err
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		s.err = err
		return err
	}
	s.rtpConn = conn
	addr, err = net.ResolveUDPAddr("udp", s.rtcp)
	if err != nil {
		s.err = err
		return err
	}
	conn, err = net.ListenUDP("udp", addr)
	if err != nil {
		s.err = err
		return err
	}
	s.rtcpConn = conn
	return s.err
}

func RegisterHandler(srv *Server, handler Handler) {
	if handler == nil {
		return
	}
	srv.handler = handler
}

func (s *Server) GracefulStop() {
	_ = s.lis.Close()
}
