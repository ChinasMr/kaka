package rtsp

import (
	"github.com/ChinasMr/kaka/pkg/transport/rtsp/method"
	"github.com/ChinasMr/kaka/pkg/transport/rtsp/transport"
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
