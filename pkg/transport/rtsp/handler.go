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
	RECORD(req *Request, res *Response) error
}

var _ Handler = (*UnimplementedServerHandler)(nil)

type UnimplementedServerHandler struct {
	sdp []byte
}

func (u *UnimplementedServerHandler) RECORD(req *Request, res *Response) error {
	log.Debugf("-->record request input data: %+v", req)
	res.SetHeader("Session", "12345678")
	log.Debugf("<--record response output data: %+v", res)
	return nil
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
	log.Debugf("<--announce response output data: %+v", res)
	return nil
}

func (u *UnimplementedServerHandler) SETUP(req *Request, res *Response) error {
	log.Debugf("-->setup request input data: %+v", req)
	transports, ok := req.Transport()
	if !ok {
		return fmt.Errorf("err setup can not get transport")
	}
	_, ok = transports["unicast"]
	if !ok {
		return fmt.Errorf("err setup can not get unicast")
	}
	_, ok = transports["mode=record"]
	if !ok {
		return fmt.Errorf("err setup can not get mode=record")
	}

	_, isUDP := transports["RTP/AVP/UDP"]
	_, isTCP := transports["RTP/AVP/TCP"]
	if isUDP == false && isTCP == false {
		return fmt.Errorf("err setup can not get RTP/AVP/UDP or RTP/AVP/TCP")
	}
	if isTCP {
		res.SetHeader("Transport", strings.Join([]string{
			"RTP/AVP/TCP",
			"unicast",
			"destionation=127.0.0.1",
			"source=127.0.0.1",
		}, ";"))
		res.SetHeader("Session", "12345678")
	}

	log.Debugf("<--setup response output data: %+v", transports)
	return nil
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
