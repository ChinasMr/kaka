package rtsp

import (
	"bufio"
	"bytes"
	"fmt"
	method2 "github.com/ChinasMr/kaka/pkg/transport/rtsp/method"
	"io"
	"net"
	"net/textproto"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type ServerTransport interface {
	Request() (*Request, error)
	Response(response *Response) error
	Addr() net.Addr
	Close() error
}

type grpcTransport struct {
	conn net.Conn
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

func (g grpcTransport) Request() (*Request, error) {
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

	method, url, proto, ok := parseRequestLine(s)
	if !ok {
		return nil, fmt.Errorf("malformed RTSP request: %s", s)
	}

	if proto != "RTSP/1.0" {
		return nil, fmt.Errorf("unsupported rtsp version: %s", method)
	}

	mimeHeader, err := tp.ReadMIMEHeader()
	if err != nil {
		return nil, err
	}

	cSeq, ok := mimeHeader["Cseq"]
	if !ok || len(cSeq) == 0 {
		return nil, fmt.Errorf("can parse cseq header, request: %s", s)
	}

	var content []byte
	cl, ok := mimeHeader["Content-Length"]
	if ok && len(cl) > 0 {
		ln, err1 := strconv.ParseUint(cl[0], 10, 64)
		if err1 != nil {
			return nil, err1
		}
		content = make([]byte, ln)
		n, err1 := io.ReadFull(g.conn, content)
		if err1 != nil {
			return nil, err1
		}

		if uint64(n) != ln {
			return nil, fmt.Errorf("err content lenth")
		}
	}

	req := &Request{
		method:  method2.Method(method),
		path:    url,
		headers: mimeHeader,
		content: content,
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

func (g grpcTransport) Response(res *Response) error {
	buf := bytes.Buffer{}
	buf.WriteString(fmt.Sprintf("%s %s %s\r\n",
		res.proto, strconv.FormatUint(res.statusCode, 10), res.status))
	cl := uint64(len(res.Content))
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
		buf.Write(res.Content)
	}

	_, err := g.conn.Write(buf.Bytes())
	return err
}

func (g grpcTransport) Addr() net.Addr {
	//TODO implement me
	panic("implement me")
}

func (g grpcTransport) Close() error {
	//TODO implement me
	panic("implement me")
}

func NewGrpcTransport(conn net.Conn) ServerTransport {
	return &grpcTransport{
		conn: conn,
	}
}
