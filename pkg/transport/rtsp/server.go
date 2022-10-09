package rtsp

import (
	"context"
	"github.com/ChinasMr/kaka/pkg/log"
	"github.com/ChinasMr/kaka/pkg/transport/rtsp/method"
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
	txs      TransactionOperator
}

func NewServer(opts ...ServerOption) *Server {
	srv := &Server{
		network: "*",
		address: ":0",
		rtp:     "30000",
		rtcp:    ":30001",
		mutex:   sync.Mutex{},
		handler: &UnimplementedServerHandler{},
		txs:     NewTxOperator(),
	}
	for _, o := range opts {
		o(srv)
	}

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

	// serve udp
	go func() {
		for true {
			buf := make([]byte, 2048)
			n, addr, err := s.conn.ReadFromUDPAddrPort(buf)
			if err != nil {
				return
			}
			go s.serverRawPackage(addr, buf[:n])
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
		s.serveWG.Add(1)
		go func() {
			s.handleRawConn(rawConn)
			_ = rawConn.Close()
			s.log.Debugf("tcp connection closed to: %v", rawConn.RemoteAddr().String())
			s.serveWG.Done()
		}()
	}

}

func (s *Server) serverRawPackage(addr netip.AddrPort, buf []byte) {

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
			s.log.Errorf("can not handle request: %v", err)
		}
	}
}

func (s *Server) handleRequest(req *request, trans Transport) error {
	var (
		res = NewResponse(req.proto, req.cSeq)
	)
	// check presentation description or media path.
	if len(req.Path()) <= 1 {
		Err500(res)
		return trans.SendResponse(res)
	}

	// stateless functions.
	if req.method == method.OPTIONS ||
		req.method == method.DESCRIBE ||
		req.method == method.ANNOUNCE {
		switch req.Method() {
		case method.OPTIONS:
			s.handler.OPTIONS(req, res)
		case method.DESCRIBE:
			s.handler.DESCRIBE(req, res)
		case method.ANNOUNCE:
			s.handler.ANNOUNCE(req, res)
		}
		return trans.SendResponse(res)
	} else if req.method == method.SETUP {
		// state functions.
		// check state, buf every state can call the setup.
		transports, ok := req.Transport()
		if !ok {
			ErrUnsupportedTransport(res)
			return trans.SendResponse(res)
		}
		ok = transports.Has("unicast")
		if !ok {
			ErrUnsupportedTransport(res)
			return trans.SendResponse(res)
		}

		sid := req.SessionID()
		tx := s.txs.GetTx(sid)
		err := s.handler.SETUP(req, res, tx)
		if err != nil {
			Err500(res)
			return trans.SendResponse(res)
		}
		res.SetHeader("Session", tx.id)
		return trans.SendResponse(res)
	} else {
		ErrMethodNotAllowed(res)
		return trans.SendResponse(res)
	}
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
