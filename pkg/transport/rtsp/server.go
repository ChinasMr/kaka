package rtsp

import (
	"context"
	"github.com/ChinasMr/kaka/pkg/log"
	"github.com/ChinasMr/kaka/pkg/transport/rtsp/header"
	"github.com/ChinasMr/kaka/pkg/transport/rtsp/methods"
	"github.com/ChinasMr/kaka/pkg/transport/rtsp/status"
	"io"
	"net"
	"net/netip"
	"sync"
	"time"
)

type Server struct {
	network          string
	address          string
	rtp              string
	rtcp             string
	lis              net.Listener
	rtpConn          *net.UDPConn
	rtcpConn         *net.UDPConn
	err              error
	timeout          time.Duration
	baseCtx          context.Context
	log              *log.Helper
	handlers         map[methods.Method]HandlerFunc
	handlerFunctions []string
	mutex            sync.Mutex
	chs              []string
	tc               TransactionController
}

func NewServer(opts ...ServerOption) *Server {
	srv := &Server{
		network:  "tcp",
		address:  ":0",
		rtp:      "30000",
		rtcp:     ":30001",
		mutex:    sync.Mutex{},
		handlers: map[methods.Method]HandlerFunc{},
		log:      log.NewHelper(log.DefaultLogger),
	}
	for _, o := range opts {
		o(srv)
	}
	srv.tc = newTransactionController(srv.chs...)
	srv.RegisterHandler(&UnimplementedServerHandler{
		tc: srv.tc,
		hs: srv.handlerFunctions,
	})
	return srv
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
	return s.serve()
}
func (s *Server) Stop(_ context.Context) error {
	s.GracefulStop()
	log.Info("[RTSP] server stopping")
	return nil
}
func (s *Server) listen() error {
	// listen rtsp tcp
	lis, err := net.Listen(s.network, s.address)
	if err != nil {
		s.err = err
		return err
	}
	s.lis = lis

	// listen rtp udp.
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

	// listen rtcp udp.
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
func (s *Server) serve() error {
	go func() {
		for true {
			buf := make([]byte, 2048)
			n, addr, err := s.rtpConn.ReadFromUDPAddrPort(buf)
			if err != nil {
				return
			}
			go s.serveRawRTP(addr, buf[:n])
		}
	}()

	go func() {
		for true {
			buf := make([]byte, 2048)
			n, addr, err := s.rtcpConn.ReadFromUDPAddrPort(buf)
			if err != nil {
				return
			}
			go s.serveRawRTCP(addr, buf[:n])
		}
	}()

	for {
		rawConn, err := s.lis.Accept()
		if err != nil {
			s.log.Errorf("can not accept new connection: %v", err)
			return err
		}

		go func() {
			s.log.Debugf("new tcp connection created from: %v", rawConn.RemoteAddr().String())
			s.handleRawConn(rawConn)
			s.log.Debugf("tcp connection closed to: %v", rawConn.RemoteAddr().String())
		}()
	}

}

func (s *Server) serveRawRTP(addr netip.AddrPort, bytes []byte) {

}
func (s *Server) serveRawRTCP(addr netip.AddrPort, bytes []byte) {

}
func (s *Server) handleRawConn(conn net.Conn) {
	trans := newTransport(conn)
	tx := s.tc.CreateTx(trans)
	s.log.Errorf("create new session for %s: %s", trans.Addr(), tx.id)
	defer func() {
		if tx != nil {
			s.tc.DeleteTx(tx.id)
			s.log.Errorf("destroy session for %s: %s", trans.Addr(), tx.id)
		}
	}()
	for {
		req, err := trans.Parse()
		if err != nil {
			if err == io.EOF {
				return
			}
			s.log.Errorf("can not parse rtsp request: %v", err)
			continue
		}
		s.log.Debugf("%s request from %s", req.method, trans.Addr())
		// create a corresponding response.
		res := NewResponse(req.proto, req.cSeq)
		// check presentation description or media path.
		if len(req.Path()) <= 1 {
			return
		}
		// handle the request.
		err = s.handleRequest(req, res, tx)
		if err != nil {
			if err == io.EOF {
				return
			}
			s.log.Errorf("can not handle request: %v", err)
			continue
		}
	}
}

func (s *Server) handleRequest(req *request, res *response, tx *transaction) error {
	// try to get the handle function.
	handlerFunc, ok := s.handlers[req.method]
	if !ok {
		return tx.Response(ErrMethodNotAllowed(res))
	}
	if req.method == methods.SETUP {
		// check state, buf every state can call the setup.
		// get and check transports header.
		if transports, has := req.Transport(); !has || !transports.Has("unicast") {
			return tx.Response(ErrUnsupportedTransport(res))
		}
		// check and set session id.
		sid := req.SessionID()
		if len(sid) != 0 && sid != sid {
			return tx.Response(ErrInternal(res))
		}
		res.SetHeader(header.Session, tx.id)
		// call the handle.
		return handlerFunc(req, res, tx)
	}

	if req.method == methods.RECORD {
		// state check
		if tx.state != status.READY && tx.state != status.RECORDING {
			return tx.Response(ErrMethodNotValidINThisState(res))
		}
		return handlerFunc(req, res, tx)
	}

	// state check
	if req.method == methods.PLAY {
		if tx.state != status.READY && tx.state != status.PLAYING {
			return tx.Response(ErrMethodNotValidINThisState(res))
		}
		return handlerFunc(req, res, tx)
	}

	if req.method == methods.TEARDOWN || req.Method() == methods.DOWN {
		_ = handlerFunc(req, res, tx)
		return io.EOF
	}

	return handlerFunc(req, res, tx)
}

func (s *Server) RegisterHandleFunc(method methods.Method, fn HandlerFunc) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.handlers[method] = fn
	s.handlerFunctions = append(s.handlerFunctions, method.String())
}
func (s *Server) RegisterHandler(handler minimalHandler) {
	s.RegisterHandleFunc(methods.OPTIONS, handler.OPTIONS)
	s.RegisterHandleFunc(methods.DESCRIBE, handler.DESCRIBE)
	s.RegisterHandleFunc(methods.ANNOUNCE, handler.ANNOUNCE)
	s.RegisterHandleFunc(methods.RECORD, handler.RECORD)
	s.RegisterHandleFunc(methods.PLAY, handler.PLAY)
	s.RegisterHandleFunc(methods.SETUP, handler.SETUP)
	s.RegisterHandleFunc(methods.TEARDOWN, handler.TEARDOWN)
	s.RegisterHandleFunc(methods.DOWN, handler.TEARDOWN)
}
func (s *Server) GracefulStop() {
	if s.lis != nil {
		_ = s.lis.Close()
	}
}
