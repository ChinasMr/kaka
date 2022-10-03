package rtsp

import (
	"bufio"
	"bytes"
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
	Addr() net.Addr
	RawConn() net.Conn
	Status() status.Status
	SetStatus(s status.Status)
	Close() error
}

type transport struct {
	conn   net.Conn
	status status.Status
	rwm    sync.RWMutex
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
