package rtsp

import (
	"fmt"
	"github.com/ChinasMr/kaka/pkg/log"
	"net/url"
)

type Handler interface {
	OPTIONS(req Request, res Response, tx Transport) error
	DESCRIBE(req Request, res Response, tx Transport) error
	SETUP(req Request, res Response, tx Transport) error
	ANNOUNCE(req Request, res Response, tx Transport) error
	RECORD(req Request, res Response, tx Transport) error
	TEARDOWN(req Request, res Response, tx Transport) error
}

var _ Handler = (*UnimplementedServerHandler)(nil)

type UnimplementedServerHandler struct {
	sdp []byte
}

func (u *UnimplementedServerHandler) TEARDOWN(req Request, res Response, tx Transport) error {
	panic("implement me")
}

func (u *UnimplementedServerHandler) RECORD(req Request, res Response, tx Transport) error {
	log.Debugf("-->record request input data: %+v", req)
	res.SetHeader("Session", "12345678")
	log.Debugf("<--record response output data: %+v", res)
	return nil
}

func (u *UnimplementedServerHandler) ANNOUNCE(req Request, res Response, tx Transport) error {
	panic("implement me")
}

func (u *UnimplementedServerHandler) SETUP(req Request, res Response, tx Transport) error {
	panic("implement me")
}

func (u *UnimplementedServerHandler) DESCRIBE(req Request, res Response, tx Transport) error {
	log.Debugf("-->options request input data: %+v", req)
	if u.sdp == nil || len(u.sdp) == 0 {
		return fmt.Errorf("this is no sdo info")
	}
	ur, err := url.Parse(req.Path())
	if err != nil {
		return nil
	}
	res.SetHeader("Content-Base", ur.String())
	res.SetHeader("Content-Type", "application/sdp")
	res.SetBody(u.sdp)
	log.Debugf("<--options response output data: %+v", res)
	return nil
}

func (u *UnimplementedServerHandler) OPTIONS(_ Request, _ Response, _ Transport) error {
	panic("implement me")
}
