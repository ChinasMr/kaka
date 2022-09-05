package log

import (
	"context"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var (
	DefaultTimestamp = TimeStamp(time.RFC3339)
	DefaultCaller    = Caller(4)
)

type Valuer func(ctx context.Context) interface{}

func Caller(depth int) Valuer {
	return func(ctx context.Context) interface{} {
		_, file, line, _ := runtime.Caller(depth)
		idx := strings.LastIndexByte(file, '/')
		return file[idx+1:] + ":" + strconv.Itoa(line)
	}
}

func TimeStamp(layout string) Valuer {
	return func(ctx context.Context) interface{} {
		return time.Now().Format(layout)
	}
}

func bindValues(ctx context.Context, keyValues []interface{}) {
	for i := 1; i < len(keyValues); i += 2 {
		if v, ok := keyValues[i].(Valuer); ok {
			keyValues[i] = v(ctx)
		}
	}
}

func containsValuer(kvs []interface{}) bool {
	for i := 1; i < len(kvs); i += 2 {
		if _, ok := kvs[i].(Valuer); ok {
			return true
		}
	}
	return false
}
