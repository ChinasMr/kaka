package rtsp

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	method2 "github.com/ChinasMr/kaka/pkg/transport/rtsp/method"
	"github.com/ChinasMr/kaka/pkg/transport/rtsp/status"
	"io"
	"net"
	"net/textproto"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
)

var _ Transport = (*transport)(nil)

type Transport interface {
	SendResponse(res Response) error
	Addr() net.Addr
	RawConn() net.Conn
	Status() status.Status
	SetStatus(s status.Status)
	ReadInterleavedFrame(frame []byte) (int, uint32, error)
	WriteInterleavedFrame(channel int, frame []byte) error
	Close() error
}

type transport struct {
	conn   net.Conn
	status status.Status
	rwm    sync.RWMutex
}

func (g *transport) WriteInterleavedFrame(channel int, frame []byte) error {
	buf := make([]byte, 2048)
	buf[0] = 0x24
	buf[1] = byte(channel)
	binary.BigEndian.PutUint16(buf[2:], uint16(len(frame)))
	n := copy(buf[4:], frame)
	_, err := g.conn.Write(buf[:4+n])
	return err
}

func (g *transport) ReadInterleavedFrame(frame []byte) (int, uint32, error) {
	if g.Status() != status.RECORD {
		return -1, 0, fmt.Errorf("status error")
	}

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

	if interleavedHeader[0] == 0x24 {
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

func (g *transport) SendResponse(res Response) error {
	r, ok := res.(*response)
	if !ok {
		return fmt.Errorf("res is not a response")
	}
	return g.sendResponse(r)
}

func (g *transport) SetStatus(s status.Status) {
	g.rwm.Lock()
	defer g.rwm.Unlock()
	g.status = s
}

func NewTransport(conn net.Conn) *transport {
	return &transport{
		conn:   conn,
		status: status.STARING,
	}
}

func (g *transport) Status() status.Status {
	g.rwm.RLock()
	g.rwm.RUnlock()
	return g.status
}

func (g *transport) RawConn() net.Conn {
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

func (g *transport) parseRequest() (*request, error) {
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

	if proto != "RTSP/1.0" {
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
		path:    urlParsed.Path,
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

func (g *transport) sendResponse(res *response) error {
	buf := bytes.Buffer{}
	buf.WriteString(fmt.Sprintf("%s %s %s\r\n",
		res.proto, strconv.FormatUint(res.statusCode, 10), res.status))
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

func (g *transport) Addr() net.Addr {
	return g.conn.RemoteAddr()
}

func (g *transport) Close() error {
	return g.conn.Close()
}
