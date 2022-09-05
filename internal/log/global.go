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

func Errorf(format string, a ...interface{}) {
	_ = global.Log(LevelError, DefaultMessageKey, fmt.Sprintf(format, a...))
}
