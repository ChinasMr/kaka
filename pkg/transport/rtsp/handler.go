package rtsp

import (
	"github.com/ChinasMr/kaka/pkg/log"
	"strings"
)

type Handler interface {
	OPTIONS(req *Request, res *Response) error
	DESCRIBE(req *Request, res *Response) error
}

var _ Handler = (*UnimplementedServerHandler)(nil)

type UnimplementedServerHandler struct {
}

func (u *UnimplementedServerHandler) DESCRIBE(req *Request, res *Response) error {
	log.Debugf("describe request input data: %+v", req)
	res.SetContent(nil)
	return nil
}

func (u *UnimplementedServerHandler) OPTIONS(req *Request, res *Response) error {
	log.Debugf("options request input data: %+v", req)
	res.SetHeader("Public",
		strings.Join([]string{
			"OPTIONS",
			"DESCRIBE",
			"ANNOUNCE",
			"SETUP",
			"PLAY",
			"PAUSE",
			"RECORD",
			"TEARDOWN",
		}, ", "),
	)
	return nil
}
