package rtsp

import (
	"fmt"
	"github.com/ChinasMr/kaka/pkg/log"
	"strings"
)

type Handler interface {
	OPTIONS(req *Request, res *Response) error
	DESCRIBE(req *Request, res *Response) error
	SETUP(req *Request, res *Response) error
	ANNOUNCE(req *Request, res *Response) error
}

var _ Handler = (*UnimplementedServerHandler)(nil)

type UnimplementedServerHandler struct {
	sdp []byte
}

func (u *UnimplementedServerHandler) ANNOUNCE(req *Request, res *Response) error {
	log.Debugf("-->announce request input data: %+v", req)
	ct, ok := req.Header("Content-Type")
	if !ok || len(ct) == 0 {
		return fmt.Errorf("can not get content type")
	}
	if ct[0] != "application/sdp" {
		return fmt.Errorf("err content type")
	}
	body := req.content
	u.sdp = body
	log.Debugf("content body: %s", string(body))
	return nil
}

func (u *UnimplementedServerHandler) SETUP(req *Request, res *Response) error {
	//TODO implement me
	panic("implement me")
}

func (u *UnimplementedServerHandler) DESCRIBE(req *Request, res *Response) error {
	//TODO implement me
	panic("implement me")
}

func (u *UnimplementedServerHandler) OPTIONS(req *Request, res *Response) error {
	log.Debugf("-->options request input data: %+v", req)
	res.SetHeader("Public",
		strings.Join([]string{
			"DESCRIBE",
			"ANNOUNCE",
			"SETUP",
			"PLAY",
			"PAUSE",
			"RECORD",
			"TEARDOWN",
		}, ", "),
	)
	log.Debugf("<--options response output data: %+v", res)
	return nil
}
