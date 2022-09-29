package transport

import "github.com/ChinasMr/kaka/pkg/transport/rtsp/method"

type Request interface {
	Method() method.Method
	Path() string
	Headers() map[string][]string
	Content() []byte
}
