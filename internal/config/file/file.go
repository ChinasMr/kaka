package file

import (
	"github.com/ChinasMr/kaka/internal/config"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var _ config.Source = (*file)(nil)

type file struct {
	path string
}

func (f *file) Load() ([]*config.KeyValue, error) {
	fi, err := os.Stat(f.path)
	if err != nil {
		return nil, err
	}
	if fi.IsDir() {
		return f.loadDir(f.path)
	}
	kv, err := f.loadFile(f.path)
	if err != nil {
		return nil, err
	}
	return []*config.KeyValue{kv}, nil
}

func (f *file) Watch() (config.Watcher, error) {
	return newWatcher(f)
}

func NewSource(path string) config.Source {
	return &file{
		path: path,
	}
}

func (f *file) loadDir(path string) ([]*config.KeyValue, error) {
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	rv := make([]*config.KeyValue, 0, len(files))
	for _, fi := range files {
		if fi.IsDir() || strings.HasPrefix(fi.Name(), ".") {
			continue
		}
		kv, err1 := f.loadFile(filepath.Join(path, fi.Name()))
		if err1 != nil {
			return nil, err1
		}
		rv = append(rv, kv)
	}
	return rv, nil
}

func (f *file) loadFile(path string) (*config.KeyValue, error) {
	fi, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = fi.Close()
	}()
	data, err := io.ReadAll(fi)
	if err != nil {
		return nil, err
	}
	info, err := fi.Stat()
	if err != nil {
		return nil, err
	}
	return &config.KeyValue{
		Key:    info.Name(),
		Value:  data,
		Format: format(info.Name()),
	}, nil
}

func format(name string) string {
	if p := strings.Split(name, "."); len(p) > 1 {
		return p[len(p)-1]
	}
	return ""
}
