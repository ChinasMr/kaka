package rtsp

import (
	"github.com/ChinasMr/kaka/pkg/log"
	"github.com/ChinasMr/kaka/pkg/transport/rtsp/header"
	"github.com/ChinasMr/kaka/pkg/transport/rtsp/methods"
	"strings"
)

// Many methods in RTSP do not contribute to state.
// However, the following play a central role in defining
// the allocation and usage of stream resources on the server:
// SETUP, PLAY, RECORD, PAUSE, and TEARDOWN.

type HandlerFunc func(req Request, res Response, tx Transaction)

// A minimal server implementation MUST be able to do the following:
// Implement the following methods: SETUP, TEARDOWN, OPTIONS and
// either PLAY (for a minimal playback server) or RECORD (for a
// minimal recording server). If RECORD is implemented, ANNOUNCE
// should be implemented as well.

type miniHandler interface {
	OPTIONS(req Request, res Response, tx Transaction)
	DESCRIBE(req Request, res Response, tx Transaction)
	ANNOUNCE(req Request, res Response, tx Transaction)
	SETUP(req Request, res Response, tx Transaction)
	RECORD(req Request, res Response, tx Transaction)
	PLAY(req Request, res Response, tx Transaction)
	TEARDOWN(req Request, res Response, tx Transaction)
}

var unimplementedServerHandler miniHandler = &UnimplementedServerHandler{}

type UnimplementedServerHandler struct {
}

func (u *UnimplementedServerHandler) OPTIONS(req Request, res Response, tx Transaction) {
	log.Debugf("options request from: %s", req.URL().String())
	method := strings.Join([]string{
		methods.DESCRIBE.String(),
		methods.ANNOUNCE.String(),
		methods.SETUP.String(),
		methods.PLAY.String(),
		methods.PAUSE.String(),
		methods.RECORD.String(),
		methods.TEARDOWN.String(),
	}, ", ")
	res.SetHeader(header.Public, method)
	return
}

func (u *UnimplementedServerHandler) ANNOUNCE(req Request, res Response, tx Transaction) {
	panic("implement me")
}

func (u *UnimplementedServerHandler) DESCRIBE(req Request, res Response, tx Transaction) {
	panic("implement me")
}

func (u *UnimplementedServerHandler) SETUP(req Request, res Response, tx Transaction) {
	panic("implement me")
}

func (u *UnimplementedServerHandler) PLAY(req Request, res Response, tx Transaction) {
	panic("implement me")
}

func (u *UnimplementedServerHandler) RECORD(req Request, res Response, tx Transaction) {
	panic("implement me")
}

func (u *UnimplementedServerHandler) TEARDOWN(req Request, res Response, tx Transaction) {
	panic("implement me")
}
