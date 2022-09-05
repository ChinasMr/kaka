package log

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"sync"
)

var _ Logger = (*stdLogger)(nil)

type stdLogger struct {
	log  *log.Logger
	pool *sync.Pool
}

func (s *stdLogger) Log(level Level, keyValues ...interface{}) error {
	if len(keyValues) == 0 {
		return nil
	}
	if (len(keyValues) & 1) == 1 {
		keyValues = append(keyValues, "KEYVALS UNPAIRED")
	}
	buf := s.pool.Get().(*bytes.Buffer)
	buf.WriteString(level.String())
	for i := 0; i < len(keyValues); i += 2 {
		_, _ = fmt.Fprintf(buf, " %s=%v", keyValues[i], keyValues[i+1])
	}
	_ = s.log.Output(4, buf.String())
	buf.Reset()
	s.pool.Put(buf)
	return nil
}

func NewStdLogger(w io.Writer) Logger {
	return &stdLogger{
		log: log.New(w, "", 0),
		pool: &sync.Pool{
			New: func() interface{} {
				return new(bytes.Buffer)
			},
		},
	}
}
