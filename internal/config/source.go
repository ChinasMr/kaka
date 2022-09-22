package config

// KeyValue is config key value.
// if source is a file, key is file name,
// value is the file content,
// format is file format suffix.
type KeyValue struct {
	Key    string
	Value  []byte
	Format string
}

// Watcher is config watcher.
type Watcher interface {
	Next() ([]*KeyValue, error)
	Stop() error
}

// Source is config source.
type Source interface {
	Load() ([]*KeyValue, error)
	Watch() (Watcher, error)
}
