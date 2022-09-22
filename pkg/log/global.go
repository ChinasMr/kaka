package log

import (
	"fmt"
	"sync"
)

var global = &loggerAppliance{}

func init() {
	global.SetLogger(DefaultLogger)
}

type loggerAppliance struct {
	lock sync.Mutex
	Logger
}

func (a *loggerAppliance) SetLogger(in Logger) {
	a.lock.Lock()
	defer a.lock.Unlock()
	a.Logger = in
}

func SetLogger(logger Logger) {
	global.SetLogger(logger)
}

func Errorf(format string, a ...interface{}) {
	_ = global.Log(LevelError, DefaultMessageKey, fmt.Sprintf(format, a...))
}

func Debugf(format string, a ...interface{}) {
	_ = global.Log(LevelDebug, DefaultMessageKey, fmt.Sprintf(format, a...))
}

func Infof(format string, a ...interface{}) {
	_ = global.Log(LevelInfo, DefaultMessageKey, fmt.Sprintf(format, a...))
}
