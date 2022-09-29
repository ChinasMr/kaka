package rtsp

import "github.com/ChinasMr/kaka/pkg/transport/rtsp/transport"

var _ transport.Response = (*Response)(nil)

type Response struct {
	proto      string
	statusCode uint64
	status     string
	headers    map[string][]string
	Content    []byte
}

func (r *Response) SetContent(body []byte) {
	r.Content = body
}

func NewResponse(proto string, cSeq string) *Response {
	return &Response{
		proto:      proto,
		statusCode: 200,
		status:     "OK",
		headers: map[string][]string{
			"CSeq": {cSeq},
		},
		Content: nil,
	}
}

func (r *Response) SetHeader(key string, value ...string) {
	r.headers[key] = value
}

func (r *Response) Code(c uint64) {
	r.statusCode = c
}
