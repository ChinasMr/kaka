package rtsp

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	method2 "github.com/ChinasMr/kaka/pkg/transport/rtsp/method"
	"io"
	"net"
	"net/textproto"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
)

const Version1 = "RTSP/1.0"

var _ Transport = (*TcpTransport)(nil)

type Transport interface {
	SendResponse(res Response) error
	SendData(ch int, data []byte) error
	ReadData(buf []byte) (int, error)
	Addr() net.Addr
	parseRequest() (*request, error)
	Close() error
}

type TcpTransport struct {
	conn net.Conn
}

func (g *TcpTransport) ReadData(buf []byte) (int, error) {
	return g.conn.Read(buf)
}

func (g *TcpTransport) SendData(ch int, data []byte) error {
	return g.writeInterleavedFrame(ch, data)
}

func NewTcpTransport(conn net.Conn) *TcpTransport {
	return &TcpTransport{
		conn: conn}
}

func (g *TcpTransport) writeInterleavedFrame(channel int, frame []byte) error {
	buf := make([]byte, 2048)
	buf[0] = 0x24
	buf[1] = byte(channel)
	binary.BigEndian.PutUint16(buf[2:], uint16(len(frame)))
	n := copy(buf[4:], frame)
	_, err := g.conn.Write(buf[:4+n])
	return err
}

func (g *TcpTransport) ReadInterleavedFrame(frame []byte) (int, uint32, error) {

	// interleavedHeader example
	// Magic:0x24   bytes 1
	// Channel:0x01 bytes 2
	// Length:84    bytes 3-4
	interleavedHeader := make([]byte, 4)
	_, err := io.ReadFull(g.conn, interleavedHeader)
	if err != nil {
		return -1, 0, err
	}

	if interleavedHeader[0] == 0x54 {
		return -1, 0, io.EOF
	}

	if interleavedHeader[1] == 0x24 {
		return -1, 0, fmt.Errorf("magic byte error")
	}

	frameLen := binary.BigEndian.Uint16(interleavedHeader[2:])
	if frameLen > 2048 {
		return -1, 0, fmt.Errorf("freame len greater than 2048")
	}
	_, err = io.ReadFull(g.conn, frame[:frameLen])
	if err != nil {
		return -1, 0, err
	}
	return int(interleavedHeader[1]), uint32(frameLen), nil
}

func (g *TcpTransport) SendResponse(res Response) error {
	r, ok := res.(*response)
	if !ok {
		return fmt.Errorf("res is not a response")
	}
	return g.sendResponse(r)
}

func (g *TcpTransport) RawConn() net.Conn {
	return g.conn
}

var readerPool sync.Pool

func newTextProtoReader(br *bufio.Reader) *textproto.Reader {
	if v := readerPool.Get(); v != nil {
		tr := v.(*textproto.Reader)
		tr.R = br
		return tr
	}
	return textproto.NewReader(br)
}
func putTextProtoReader(r *textproto.Reader) {
	r.R = nil
	readerPool.Put(r)
}

func (g *TcpTransport) parseRequest() (*request, error) {
	br := bufio.NewReader(g.conn)
	tp := newTextProtoReader(br)
	var s string
	var err error
	if s, err = tp.ReadLine(); err != nil {
		return nil, err
	}

	defer func() {
		putTextProtoReader(tp)
	}()

	method, urlRaw, proto, ok := parseRequestLine(s)
	if !ok {
		return nil, fmt.Errorf("malformed RTSP request: %s", s)
	}

	if proto != Version1 {
		return nil, fmt.Errorf("unsupported rtsp version: %s", method)
	}

	urlParsed, err := url.Parse(urlRaw)
	if err != nil {
		return nil, err
	}

	mimeHeader, err := tp.ReadMIMEHeader()
	if err != nil {
		return nil, err
	}

	cSeq, ok := mimeHeader["Cseq"]
	if !ok || len(cSeq) == 0 {
		return nil, fmt.Errorf("can parse cseq header, request: %s", s)
	}

	var body []byte
	cl, ok := mimeHeader["Content-Length"]
	if ok && len(cl) > 0 {
		ln, err1 := strconv.ParseUint(cl[0], 10, 64)
		if err1 != nil {
			return nil, err1
		}
		body = make([]byte, ln)
		n, err1 := io.ReadFull(g.conn, body)
		if err1 != nil {
			return nil, err1
		}

		if uint64(n) != ln {
			return nil, fmt.Errorf("err content lenth")
		}
	}

	req := &request{
		method:  method2.Method(method),
		url:     urlParsed,
		headers: mimeHeader,
		body:    body,
		cSeq:    cSeq[0],
		proto:   proto,
	}

	return req, nil
}

func parseRequestLine(line string) (string, string, string, bool) {
	method, rest, ok1 := strings.Cut(line, " ")
	requestURI, proto, ok2 := strings.Cut(rest, " ")
	if !ok1 || !ok2 {
		return "", "", "", false
	}
	return method, requestURI, proto, true
}

func (g *TcpTransport) sendResponse(res *response) error {
	buf := bytes.Buffer{}
	buf.WriteString(fmt.Sprintf("%s %s %s\r\n",
		res.proto, strconv.FormatUint(res.code, 10), res.status))
	cl := uint64(len(res.body))
	if cl > 0 {
		res.SetHeader("Content-Length", strconv.FormatUint(cl, 10))
	}

	var keys []string
	for key := range res.headers {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		buf.WriteString(fmt.Sprintf("%s: %s\r\n", key, strings.Join(res.headers[key], " ")))
	}
	buf.WriteString("\r\n")

	if cl > 0 {
		buf.Write(res.body)
	}

	_, err := g.conn.Write(buf.Bytes())
	return err
}

func (g *TcpTransport) Addr() net.Addr {
	return g.conn.RemoteAddr()
}

func (g *TcpTransport) Close() error {
	return g.conn.Close()
}
