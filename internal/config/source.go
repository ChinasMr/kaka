package config

type KeyValue struct {
	Key    string
	Value  []byte
	Format string
}

type Watcher interface {
	Next() ([]*KeyValue, error)
	Stop() error
}

type Source interface {
	Load() ([]*KeyValue, error)
	Watch() (Watcher, error)
}
