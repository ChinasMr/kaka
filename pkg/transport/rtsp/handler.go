package rtsp

import (
	"github.com/ChinasMr/kaka/pkg/log"
	"github.com/ChinasMr/kaka/pkg/transport/rtsp/header"
	"github.com/ChinasMr/kaka/pkg/transport/rtsp/method"
	"strings"
)

// Many methods in RTSP do not contribute to state.
// However, the following play a central role in defining
// the allocation and usage of stream resources on the server:
// SETUP, PLAY, RECORD, PAUSE, and TEARDOWN.

type Handler interface {
	OPTIONS(req Request, res Response)
	DESCRIBE(req Request, res Response)
	ANNOUNCE(req Request, res Response)

	SETUP(req Request, res Response, tx Transaction) error
	RECORD(req Request, res Response, tx Transaction) error
	PLAY(req Request, res Response, tx Transaction) error
	TEARDOWN(req Request, res Response, tx Transaction) error
}

var _ Handler = (*UnimplementedServerHandler)(nil)

type UnimplementedServerHandler struct {
}

func (u *UnimplementedServerHandler) OPTIONS(req Request, res Response) {
	log.Debugf("options request from: %s", req.URL().String())
	methods := strings.Join([]string{
		method.DESCRIBE.String(),
		method.ANNOUNCE.String(),
		method.SETUP.String(),
		method.PLAY.String(),
		method.PAUSE.String(),
		method.RECORD.String(),
		method.TEARDOWN.String(),
	}, ", ")
	res.SetHeader(header.Public, methods)
	return
}

func (u *UnimplementedServerHandler) ANNOUNCE(req Request, res Response) {
	panic("implement me")
}

func (u *UnimplementedServerHandler) DESCRIBE(req Request, res Response) {
	panic("implement me")
}

func (u *UnimplementedServerHandler) SETUP(req Request, res Response, tx Transaction) error {
	panic("implement me")
}

func (u *UnimplementedServerHandler) PLAY(req Request, res Response, tx Transaction) error {
	panic("implement me")
}

func (u *UnimplementedServerHandler) RECORD(req Request, res Response, tx Transaction) error {
	panic("implement me")
}

func (u *UnimplementedServerHandler) TEARDOWN(req Request, res Response, tx Transaction) error {
	panic("implement me")
}
