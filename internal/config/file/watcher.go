package file

import (
	"context"
	"github.com/ChinasMr/kaka/internal/config"
	"github.com/fsnotify/fsnotify"
	"os"
	"path/filepath"
)

type watcher struct {
	f  *file
	fw *fsnotify.Watcher

	ctx    context.Context
	cancel context.CancelFunc
}

func (w *watcher) Next() ([]*config.KeyValue, error) {
	select {
	case <-w.ctx.Done():
		return nil, w.ctx.Err()
	case event := <-w.fw.Events:
		// rename file.
		if event.Op == fsnotify.Rename {
			_, err := os.Stat(event.Name)
			if err == nil || os.IsExist(err) {
				err1 := w.fw.Add(event.Name)
				if err1 != nil {
					return nil, err1
				}
			}

		}

		fi, err := os.Stat(w.f.path)
		if err != nil {
			return nil, err
		}
		path := w.f.path
		if fi.IsDir() {
			path = filepath.Join(w.f.path, filepath.Base(event.Name))
		}

		kv, err := w.f.loadFile(path)
		if err != nil {
			return nil, err
		}
		return []*config.KeyValue{kv}, nil

	case err := <-w.fw.Errors:
		return nil, err
	}
}

func (w *watcher) Stop() error {
	w.cancel()
	return w.fw.Close()
}

var _ config.Watcher = (*watcher)(nil)

func newWatcher(f *file) (config.Watcher, error) {
	fw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	err = fw.Add(f.path)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &watcher{
		f:      f,
		fw:     fw,
		ctx:    ctx,
		cancel: cancel,
	}, nil
}
