package rtsp

import (
	"context"
	"github.com/ChinasMr/kaka/pkg/log"
	"github.com/ChinasMr/kaka/pkg/transport/rtsp/methods"
	"io"
	"net"
	"net/netip"
	"sync"
	"time"
)

type Server struct {
	network  string
	address  string
	rtp      string
	rtcp     string
	lis      net.Listener
	rtpConn  *net.UDPConn
	rtcpConn *net.UDPConn
	err      error
	timeout  time.Duration
	baseCtx  context.Context
	log      *log.Helper
	handlers map[methods.Method]HandlerFunc
	mutex    sync.Mutex
	txs      TransactionOperator
}

func NewServer(opts ...ServerOption) *Server {
	srv := &Server{
		network:  "tcp",
		address:  ":0",
		rtp:      "30000",
		rtcp:     ":30001",
		mutex:    sync.Mutex{},
		handlers: map[methods.Method]HandlerFunc{},
		txs:      NewTxOperator(),
	}
	for _, o := range opts {
		o(srv)
	}
	srv.RegisterHandler(methods.OPTIONS, unimplementedServerHandler.OPTIONS)
	srv.RegisterHandler(methods.DESCRIBE, unimplementedServerHandler.DESCRIBE)
	srv.RegisterHandler(methods.ANNOUNCE, unimplementedServerHandler.ANNOUNCE)
	srv.RegisterHandler(methods.RECORD, unimplementedServerHandler.RECORD)
	srv.RegisterHandler(methods.PLAY, unimplementedServerHandler.PLAY)
	srv.RegisterHandler(methods.SETUP, unimplementedServerHandler.SETUP)
	srv.RegisterHandler(methods.TEARDOWN, unimplementedServerHandler.TEARDOWN)
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

	// serve tcp
	for {
		rawConn, err := s.lis.Accept()
		if err != nil {
			s.log.Errorf("can not accept new connection: %v", err)
			return err
		}
		s.log.Debugf("new tcp connection created from: %v", rawConn.RemoteAddr().String())
		go func() {
			defer func() {
				_ = rawConn.Close()
			}()
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
	tc := NewTcpTransport(conn)
	for {
		req, err := tc.parseRequest()
		if err != nil {
			if err == io.EOF {
				continue
			}
			s.log.Errorf("can not parse rtsp request: %v", err)
			return
		}
		// todo remove the log print.
		s.log.Debugf("%s request from %s", req.Method(), tc.Addr().String())
		err = s.handleRequest(req, tc)
		if err != nil {
			if err == io.EOF {
				return
			}
			s.log.Errorf("can not handle request: %v", err)
		}
	}
}

func (s *Server) handleRequest(req *request, trans Transport) error {
	var (
		res = NewResponse(req.proto, req.cSeq)
	)
	// check presentation description or media path.
	if len(req.Path()) <= 0 {
		Err500(res)
		return trans.SendResponse(res)
	}

	// stateless functions.
	if req.method == methods.OPTIONS ||
		req.method == methods.DESCRIBE ||
		req.method == methods.ANNOUNCE {
		handlerFunc, ok := s.handlers[req.method]
		if !ok {
			ErrMethodNotAllowed(res)
			return trans.SendResponse(res)
		}
		// auto send response.
		handlerFunc(req, res, nil)
		return nil
	} // else if req.method == methods.SETUP {
	//	// check state, buf every state can call the setup.
	//	// get transports header.
	//	transports, ok := req.Transport()
	//	if !ok {
	//		ErrUnsupportedTransport(res)
	//		return trans.SendResponse(res)
	//	}
	//
	//	// disabled multicast.
	//	ok = transports.Has("unicast")
	//	if !ok {
	//		ErrUnsupportedTransport(res)
	//		return trans.SendResponse(res)
	//	}
	//
	//	// get tx
	//	tx := s.txs.GetTx(req.SessionID())
	//	// refresh the transport
	//	tx.trans = trans
	//	// handle request
	//	return s.handler.SETUP(req, res, tx)
	//} else if req.method == methods.RECORD {
	//	tx := s.txs.GetTx(req.SessionID())
	//	// state check
	//	if tx.state != status.READY && tx.state != status.RECORDING {
	//		ErrMethodNotValidINThisState(res)
	//		return trans.SendResponse(res)
	//	}
	//	// refresh the transport
	//	tx.trans = trans
	//	// handle request
	//	return s.handler.RECORD(req, res, tx)
	//} else if req.method == methods.PLAY {
	//	tx := s.txs.GetTx(req.SessionID())
	//	// state check
	//	if tx.state != status.READY && tx.state != status.PLAYING {
	//		ErrMethodNotValidINThisState(res)
	//		return trans.SendResponse(res)
	//	}
	//	// refresh the transport
	//	tx.trans = trans
	//	// handle request
	//	return s.handler.PLAY(req, res, tx)
	//} else if req.method == methods.TEARDOWN || req.method == methods.DOWN {
	//	tx := s.txs.GetTx(req.SessionID())
	//	tx.trans = trans
	//	// state check, you can call teardown anytime.
	//	return s.handler.TEARDOWN(req, res, tx)
	//} else {
	//	ErrMethodNotAllowed(res)
	//	return trans.SendResponse(res)
	//}
	return nil
}

func (s *Server) RegisterHandler(method methods.Method, fn HandlerFunc) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.handlers[method] = fn
}
func (s *Server) GracefulStop() {
	if s.lis != nil {
		_ = s.lis.Close()
	}
}
