package rtsp

import (
	"bufio"
	"fmt"
	"github.com/ChinasMr/kaka/pkg/transport/rtsp/methods"
	"io"
	"net/textproto"
	"net/url"
	"strconv"
	"strings"
	"sync"
)

// text proto reader pool.
var readerPool sync.Pool

// parse a request from reader.
func parse0(rd io.Reader) (*request, error) {
	br := bufio.NewReader(rd)
	tp := newTextProtoReader(br)
	defer func() {
		putTextProtoReader(tp)
	}()
	var s string
	var err error
	if s, err = tp.ReadLine(); err != nil {
		return nil, err
	}
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
		n, err1 := io.ReadFull(br, body)
		if err1 != nil {
			return nil, err1
		}
		if uint64(n) != ln {
			return nil, fmt.Errorf("err content lenth")
		}
	}
	return &request{
		method:  methods.Method(method),
		url:     urlParsed,
		headers: mimeHeader,
		body:    body,
		cSeq:    cSeq[0],
		proto:   proto,
	}, nil
}

// parse the request line.
// OPTIONS rtsp://192.168.0.1:554/live RTSP/1.0
func parseRequestLine(line string) (string, string, string, bool) {
	method, rest, ok1 := strings.Cut(line, " ")
	requestURI, proto, ok2 := strings.Cut(rest, " ")
	if !ok1 || !ok2 {
		return "", "", "", false
	}
	return method, requestURI, proto, true
}

// get new proto reader from the pool.
func newTextProtoReader(br *bufio.Reader) *textproto.Reader {
	if v := readerPool.Get(); v != nil {
		tr := v.(*textproto.Reader)
		tr.R = br
		return tr
	}
	return textproto.NewReader(br)
}

// put the proto reader back to the pool.
func putTextProtoReader(r *textproto.Reader) {
	r.R = nil
	readerPool.Put(r)
}
