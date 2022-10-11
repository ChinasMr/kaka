package rtsp

import (
	"net"
)

const Version1 = "RTSP/1.0"

var _ Transport = (*transport)(nil)

type Transport interface {
	Addr() string
	Parse() (*request, error)
	Write(data []byte) error
	Read(buf []byte) (int, error)
	Conn() net.Conn
	Close() error
}

type transport struct {
	conn net.Conn
}

func (g *transport) Conn() net.Conn {
	return g.conn
}

func newTransport(conn net.Conn) *transport {
	return &transport{
		conn: conn,
	}
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
