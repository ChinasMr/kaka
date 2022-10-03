package rtsp

type Handler interface {
	OPTIONS(req Request, res Response, tx Transport) error
	DESCRIBE(req Request, res Response, tx Transport) error
	SETUP(req Request, res Response, tx Transport) error
	ANNOUNCE(req Request, res Response, tx Transport) error
	RECORD(req Request, res Response, tx Transport) error
	PLAY(req Request, res Response, tx Transport) error
	TEARDOWN(req Request, res Response, tx Transport) error
}

var _ Handler = (*UnimplementedServerHandler)(nil)

type UnimplementedServerHandler struct {
	sdp []byte
}

func (u *UnimplementedServerHandler) PLAY(req Request, res Response, tx Transport) error {
	panic("implement me")
}

func (u *UnimplementedServerHandler) TEARDOWN(req Request, res Response, tx Transport) error {
	panic("implement me")
}

func (u *UnimplementedServerHandler) RECORD(req Request, res Response, tx Transport) error {
	panic("implement me")
}

func (u *UnimplementedServerHandler) ANNOUNCE(req Request, res Response, tx Transport) error {
	panic("implement me")
}

func (u *UnimplementedServerHandler) SETUP(req Request, res Response, tx Transport) error {
	panic("implement me")
}

func (u *UnimplementedServerHandler) DESCRIBE(req Request, res Response, tx Transport) error {
	panic("implement me")
}

func (u *UnimplementedServerHandler) OPTIONS(_ Request, _ Response, _ Transport) error {
	panic("implement me")
}
