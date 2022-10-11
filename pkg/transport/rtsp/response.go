package rtsp

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type Response interface {
	SetHeader(key string, value ...string)
	SetBody(body []byte)
	SetStatus(status string)
	SetCode(code uint64)
	Encoding() []byte
}

var _ Response = (*response)(nil)

type response struct {
	proto   string
	code    uint64
	status  string
	headers map[string][]string
	body    []byte
}

func (r *response) Encoding() []byte {
	buf := bytes.Buffer{}
	buf.WriteString(fmt.Sprintf("%s %s %s\r\n",
		r.proto, strconv.FormatUint(r.code, 10), r.status))
	cl := uint64(len(r.body))
	if cl > 0 {
		r.SetHeader("Content-Length", strconv.FormatUint(cl, 10))
	}

	var keys []string
	for key := range r.headers {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		buf.WriteString(fmt.Sprintf("%s: %s\r\n", key,
			strings.Join(r.headers[key], " ")))
	}
	buf.WriteString("\r\n")

	if cl > 0 {
		buf.Write(r.body)
	}
	return buf.Bytes()
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
