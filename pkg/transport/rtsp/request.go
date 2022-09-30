package rtsp

import (
	"github.com/ChinasMr/kaka/pkg/transport/rtsp/method"
	"github.com/ChinasMr/kaka/pkg/transport/rtsp/transport"
	"strings"
)

var _ transport.Request = (*Request)(nil)

type Request struct {
	method  method.Method
	path    string
	headers map[string][]string
	content []byte
	cSeq    string
	proto   string
}

func (r Request) Transport() (map[string]struct{}, bool) {
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

func (r Request) Header(key string) ([]string, bool) {
	rv, ok := r.headers[key]
	return rv, ok
}

func (r Request) Content() []byte {
	return r.content
}

func (r Request) Path() string {
	return r.path
}

func (r Request) Headers() map[string][]string {
	return r.headers
}

func (r Request) Method() method.Method {
	return r.method
}
