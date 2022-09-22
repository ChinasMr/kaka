package log

import "fmt"

var DefaultMessageKey = "msg"

type Option func(helper *Helper)

type Helper struct {
	logger Logger
	msgKey string
}

func NewHelper(logger Logger, opts ...Option) *Helper {
	options := &Helper{
		logger: logger,
		msgKey: DefaultMessageKey,
	}
	for _, o := range opts {
		o(options)
	}
	return options
}

func (h *Helper) Debug(a ...interface{}) {
	_ = h.logger.Log(LevelDebug, h.msgKey, fmt.Sprint(a...))
}

func (h *Helper) Debugf(format string, a ...interface{}) {
	_ = h.logger.Log(LevelDebug, h.msgKey, fmt.Sprintf(format, a...))
}

func (h *Helper) Debugw(a ...interface{}) {
	_ = h.logger.Log(LevelDebug, fmt.Sprint(a...))
}

func (h *Helper) Info(a ...interface{}) {
	_ = h.logger.Log(LevelInfo, h.msgKey, fmt.Sprint(a...))
}

func (h *Helper) Infof(format string, a ...interface{}) {
	_ = h.logger.Log(LevelInfo, h.msgKey, fmt.Sprintf(format, a...))
}

func (h *Helper) Infow(a ...interface{}) {
	_ = h.logger.Log(LevelInfo, fmt.Sprint(a...))
}
