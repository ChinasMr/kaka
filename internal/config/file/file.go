package file

import (
	"github.com/ChinasMr/kaka/internal/config"
	"os"
)

type file struct {
	path string
}

func (f *file) Load() ([]*config.KeyValue, error) {
	_, err := os.Stat(f.path)
	return nil, err
}

func (f *file) Watch() (config.Watcher, error) {
	//TODO implement me
	panic("implement me")
}

func NewSource(path string) config.Source {
	return &file{path: path}
}
