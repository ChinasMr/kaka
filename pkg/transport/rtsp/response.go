package rtsp

type Response interface {
	Code(c uint64)
	SetHeader(key string, value ...string)
	SetBody(body []byte)
}

var _ Response = (*response)(nil)

type response struct {
	proto      string
	statusCode uint64
	status     string
	headers    map[string][]string
	body       []byte
}

func (r *response) SetBody(body []byte) {
	r.body = body
}

func NewResponse(proto string, cSeq string) *response {
	return &response{
		proto:      proto,
		statusCode: 200,
		status:     "OK",
		headers: map[string][]string{
			"CSeq": {cSeq},
		},
		body: nil,
	}
}

func (r *response) SetHeader(key string, value ...string) {
	r.headers[key] = value
}

func (r *response) Code(c uint64) {
	r.statusCode = c
}
