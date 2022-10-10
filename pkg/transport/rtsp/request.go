package rtsp

import (
	"github.com/ChinasMr/kaka/pkg/transport/rtsp/header"
	"github.com/ChinasMr/kaka/pkg/transport/rtsp/method"
	"net/url"
	"strings"
)

var _ Request = (*request)(nil)

type Request interface {
	Method() method.Method
	URL() *url.URL
	Path() string
	Headers() map[string][]string
	Header(key string) ([]string, bool)
	Transport() (header.TransportHeader, bool)
	CSeq() string
	Proto() string
	SessionID() string
	ContentType() string
	Body() []byte
}

type request struct {
	method  method.Method
	url     *url.URL
	headers map[string][]string
	body    []byte
	cSeq    string
	proto   string
}

func (r request) ContentType() string {
	ct, ok := r.headers[header.ContentType]
	if !ok || len(ct) == 0 {
		return ""
	}
	return ct[0]
}

func (r request) SessionID() string {
	session, ok := r.headers["Session"]
	if !ok || len(session) == 0 {
		return ""
	}
	return session[0]
}

func (r request) URL() *url.URL {
	return r.url
}

func (r request) Body() []byte {
	return r.body
}

func (r request) Proto() string {
	return r.proto
}

func (r request) CSeq() string {
	return r.cSeq
}

func (r request) Transport() (header.TransportHeader, bool) {
	trans, ok := r.headers["Transport"]
	if ok == false || len(trans) == 0 {
		return nil, ok
	}
	transports := make(map[string]struct{})
	for _, part := range strings.Split(trans[0], ";") {
		transports[part] = struct{}{}
	}
	return transports, true
}

func (r request) Header(key string) ([]string, bool) {
	rv, ok := r.headers[key]
	return rv, ok
}

func (r request) Path() string {
	return r.url.Path
}

func (r request) Headers() map[string][]string {
	return r.headers
}

func (r request) Method() method.Method {
	return r.method
}
