package app

import (
	"context"
	"github.com/ChinasMr/kaka/pkg/log"
	"github.com/google/uuid"
	"os"
	"sync"
	"syscall"
	"time"
)

type Info interface {
	ID() string
	Name() string
	Version() string
	Metadata() map[string]string
}

type App struct {
	opts   options
	ctx    context.Context
	cancel func()
	mu     sync.Mutex
}

func (a *App) Run() error {
	for true {
	}
	return nil
}

func (a *App) ID() string {
	return a.opts.id
}

func (a *App) Name() string {
	return a.opts.name
}

func (a *App) Version() string {
	return a.opts.version
}

func (a *App) Metadata() map[string]string {
	return a.opts.metadata
}

func New(opts ...Option) *App {
	o := options{
		ctx:         context.Background(),
		signals:     []os.Signal{syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT},
		stopTimeout: 10 * time.Second,
	}
	id, err := uuid.NewUUID()
	if err == nil {
		o.id = id.String()
	}
	for _, opt := range opts {
		opt(&o)
	}
	if o.logger != nil {
		log.SetLogger(o.logger)
	}
	ctx, cancel := context.WithCancel(o.ctx)
	return &App{
		opts:   o,
		ctx:    ctx,
		cancel: cancel,
		mu:     sync.Mutex{},
	}
}
