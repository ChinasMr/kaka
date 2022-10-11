package rtsp

import (
	"fmt"
	"github.com/ChinasMr/kaka/pkg/transport/rtsp/header"
	"github.com/ChinasMr/kaka/pkg/transport/rtsp/methods"
	"gortc.io/sdp"
	"net/url"
	"strings"
)

var _ Request = (*request)(nil)

type Request interface {
	Method() methods.Method
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
	Encode() []byte
	ParseSDP() (*sdp.Message, error)
}

type request struct {
	method  methods.Method
	url     *url.URL
	headers map[string][]string
	body    []byte
	cSeq    string
	proto   string
}

func (r request) ParseSDP() (*sdp.Message, error) {
	if r.ContentType() != header.ContentTypeSDP {
		return nil, fmt.Errorf("unsupported presentation description format: %s", r.ContentType())
	}
	if len(r.body) == 0 {
		return nil, fmt.Errorf("paerse sdp error empty request body")
	}
	s, err := sdp.DecodeSession(r.body, nil)
	if err != nil {
		return nil, err
	}
	rv := &sdp.Message{}
	d := sdp.NewDecoder(s)
	err = d.Decode(rv)
	if err != nil {
		return nil, err
	}
	return rv, nil
}

func (r request) Encode() []byte {
	//TODO implement me
	panic("implement me")
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

func (r request) Method() methods.Method {
	return r.method
}

func (r request) Channel() string {
	str := strings.TrimLeft(r.Path(), "/")
	index := strings.Index(str, "/")
	if index != -1 {
		return str[:index]
	}
	return str
}
