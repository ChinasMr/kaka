package config

import "sync"

var (
	_ Config = (*config)(nil)
)

// Observer is config observer.
type Observer func(string, Value)

// Config is a config interface.
type Config interface {
	Load() error
	Scan(v interface{}) error
	Value(key string) Value
	Watch(key string, o Observer) error
	Close() error
}

type config struct {
	opts      options
	reader    Reader
	cached    sync.Map
	Observers sync.Map
	watchers  []Watcher
}

func (c *config) Load() error {
	for _, src := range c.opts.sources {
		kvs, err := src.Load()
		if err != nil {
			return err
		}
	}
}

func (c *config) Scan(v interface{}) error {
	//TODO implement me
	panic("implement me")
}

func (c *config) Value(key string) Value {
	//TODO implement me
	panic("implement me")
}

func (c *config) Watch(key string, o Observer) error {
	//TODO implement me
	panic("implement me")
}

func (c *config) Close() error {
	//TODO implement me
	panic("implement me")
}

// New a config with options.
func New(opts ...Option) Config {
	o := options{
		sources:  nil,
		decoder:  DefaultDecoder,
		resolver: DefaultResolver,
	}
	for _, opt := range opts {
		opt(&o)
	}
	return &config{
		opts:      o,
		reader:    NewReader(o),
		cached:    sync.Map{},
		Observers: sync.Map{},
		watchers:  nil,
	}
}
