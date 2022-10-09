package rtsp

type Response interface {
	SetHeader(key string, value ...string)
	SetBody(body []byte)
	SetStatus(status string)
	SetCode(code uint64)
}

var _ Response = (*response)(nil)

func Err404(res Response) {
	res.SetStatus("Not Found")
	res.SetCode(404)
}

func Err500(res Response) {
	res.SetStatus("Internal Server Error")
	res.SetCode(500)
}

func ErrUnsupportedTransport(res Response) {
	res.SetStatus("Unsupported Transport")
	res.SetCode(461)
}

func ErrMethodNotAllowed(res Response) {
	res.SetStatus("Method Not Allowed")
	res.SetCode(405)
}

type response struct {
	proto   string
	code    uint64
	status  string
	headers map[string][]string
	body    []byte
}

func (r *response) SetStatus(status string) {
	r.status = status
}

func (r *response) SetCode(code uint64) {
	r.code = code
}

func (r *response) SetBody(body []byte) {
	r.body = body
}

func NewResponse(proto string, cSeq string) *response {
	return &response{
		proto:  proto,
		code:   200,
		status: "OK",
		headers: map[string][]string{
			"CSeq": {cSeq},
		},
		body: nil,
	}
}

func (r *response) SetHeader(key string, value ...string) {
	r.headers[key] = value
}
