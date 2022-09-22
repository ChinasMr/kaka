package log

import (
	"context"
	"log"
)

var DefaultLogger = NewStdLogger(log.Writer())

type Logger interface {
	Log(level Level, keyValues ...interface{}) error
}

type logger struct {
	logger    Logger
	prefix    []interface{}
	hasValuer bool
	ctx       context.Context
}

func (l *logger) Log(level Level, keyValues ...interface{}) error {
	kvs := make([]interface{}, 0, len(l.prefix)+len(keyValues))
	kvs = append(kvs, l.prefix...)
	if l.hasValuer {
		bindValues(l.ctx, kvs)
	}
	kvs = append(kvs, keyValues...)
	return l.logger.Log(level, kvs...)
}

func With(l Logger, kv ...interface{}) Logger {
	c, ok := l.(*logger)
	if !ok {
		return &logger{
			logger:    l,
			prefix:    kv,
			hasValuer: containsValuer(kv),
			ctx:       context.Background(),
		}
	}
	kvs := make([]interface{}, 0, len(c.prefix)+len(kv))
	kvs = append(kvs, kv...)
	kvs = append(kvs, c.prefix)
	return &logger{
		logger:    c.logger,
		prefix:    kvs,
		hasValuer: containsValuer(kvs),
		ctx:       c.ctx,
	}
}
