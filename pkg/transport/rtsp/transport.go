package rtsp

import (
	"net"
)

const Version1 = "RTSP/1.0"

var _ Transport = (*transport)(nil)

type Transport interface {
	Addr() string
	IP() net.IP
	Parse() (*request, error)
	Write(data []byte) error
	Read(buf []byte) (int, error)
	Conn() net.Conn
	Close() error
}

type transport struct {
	conn net.Conn
	addr *net.TCPAddr
}

func (g *transport) IP() net.IP {
	return g.addr.IP
}

func (g *transport) Conn() net.Conn {
	return g.conn
}

func newTransport(conn net.Conn) (*transport, error) {
	conn.RemoteAddr()
	addr, err := net.ResolveTCPAddr(conn.RemoteAddr().Network(), conn.RemoteAddr().String())
	if err != nil {
		return nil, err
	}
	return &transport{
		conn: conn,
		addr: addr,
	}, nil
}

func (g *transport) Write(data []byte) error {
	_, err := g.conn.Write(data)
	return err
}

func (g *transport) Read(buf []byte) (int, error) {
	return g.conn.Read(buf)
}

func (g *transport) Parse() (*request, error) {
	return parse0(g.conn)
}

func (g *transport) Addr() string {
	return g.conn.RemoteAddr().String()
}

func (g *transport) Close() error {
	return g.conn.Close()
}
