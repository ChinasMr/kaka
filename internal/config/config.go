package config

import (
	"context"
	"errors"
	"github.com/ChinasMr/kaka/internal/log"
	"sync"
	"time"
)

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
		// load bytes data from dir or file.
		kvs, err := src.Load()
		if err != nil {
			return err
		}
		for _, v := range kvs {
			log.Debugf("config loaded: %s format: %s", v.Key, v.Format)
		}

		// merge the configs.
		err = c.reader.Merge(kvs...)
		if err != nil {
			log.Errorf("failed to merge config source: %v", err)
			return err
		}

		w, err := src.Watch()
		if err != nil {
			log.Errorf("failed to watch config source: %v", err)
			return err
		}
		// collect the watcher.
		c.watchers = append(c.watchers, w)
		go c.watch(w)
	}
	err := c.reader.Resolve()
	if err != nil {
		log.Errorf("failed to resolve config source: %v", err)
		return err
	}
	return nil
}

func (c *config) watch(w Watcher) {
	for {
		kvs, err := w.Next()
		if errors.Is(err, context.Canceled) {
			log.Infof("watcher's ctx cancel: %v", err)
			return
		}
		if err != nil {
			time.Sleep(time.Second)
			log.Errorf("failed to watch next config: %v", err)
			continue
		}
		// re merge.
		err = c.reader.Merge(kvs...)
		if err != nil {
			log.Errorf("failed to merge next config: %v", err)
			continue
		}
		err = c.reader.Resolve()
		if err != nil {
			log.Errorf("failed to resolve next config: %v", err)
			continue
		}
		c.cached.Range(func(key, value any) bool {
			//k := key.(string)
			//v := value.(Value)
			//if
			return true
		})
	}
}

func (c *config) Scan(v interface{}) error {
	data, err := c.reader.Source()
	if err != nil {
		return err
	}
	return unmarshalJSON(data, v)
}

func (c *config) Value(key string) Value {
	v, ok := c.cached.Load(key)
	if ok {
		return v.(Value)
	}
	v, ok := c.reader.Value(key)
	if ok {
		c.cached.Store(key, v)
		return v
	}
	return &er

}

func (c *config) Watch(key string, o Observer) error {
	//TODO implement me
	panic("implement me")
}

func (c *config) Close() error {
	for _, w := range c.watchers {
		err := w.Stop()
		if err != nil {
			return err
		}
	}
	return nil
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
