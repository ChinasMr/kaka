package app

import (
	"context"
	"github.com/ChinasMr/kaka/pkg/log"
	"github.com/ChinasMr/kaka/pkg/transport"
	"os"
	"time"
)

type Option func(o *options)

type options struct {
	id       string
	name     string
	version  string
	metadata map[string]string

	ctx     context.Context
	signals []os.Signal

	logger      log.Logger
	stopTimeout time.Duration
	servers     []transport.Server
}

func ID(id string) Option {
	return func(o *options) {
		o.id = id
	}
}

func Name(name string) Option {
	return func(o *options) {
		o.name = name
	}
}

func Version(version string) Option {
	return func(o *options) {
		o.version = version
	}
}

func Metadata(md map[string]string) Option {
	return func(o *options) {
		o.metadata = md
	}
}

func Logger(logger log.Logger) Option {
	return func(o *options) {
		o.logger = logger
	}
}

func Server(srv ...transport.Server) Option {
	return func(o *options) {
		o.servers = srv
	}
}
