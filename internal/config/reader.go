package config

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"github.com/ChinasMr/kaka/internal/log"
	"github.com/imdario/mergo"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"strings"
	"sync"
)

type Reader interface {
	Merge(...*KeyValue) error
	Value(string) (Value, bool)
	Source() ([]byte, error)
	Resolve() error
}

type reader struct {
	opts   options
	values map[string]interface{}
	lock   sync.Mutex
}

// NewReader return a reader.
func NewReader(opts options) Reader {
	return &reader{
		opts:   opts,
		values: make(map[string]interface{}),
		lock:   sync.Mutex{},
	}
}

func (r *reader) Merge(kvs ...*KeyValue) error {
	r.lock.Lock()
	merged, err := cloneMap(r.values)
	r.lock.Unlock()
	if err != nil {
		return err
	}
	for _, kv := range kvs {
		next := make(map[string]interface{})
		err1 := r.opts.decoder(kv, next)
		if err1 != nil {
			log.Errorf("Failed to config decode error: %v key: %s value: %s", err1, kv.Key, string(kv.Value))
			return err1
		}
		// todo there may be a bug.
		err1 = mergo.Map(&merged, next, mergo.WithOverride)
		if err1 != nil {
			log.Errorf("Failed to config merge error: %v key: %s value: %s", err1, kv.Key, string(kv.Value))
			return err1
		}
	}
	r.lock.Lock()
	r.values = merged
	r.lock.Unlock()
	return nil
}

func (r *reader) Value(path string) (Value, bool) {
	r.lock.Lock()
	defer r.lock.Unlock()
	return readValue(r.values, path)
}

func (r *reader) Source() ([]byte, error) {
	r.lock.Lock()
	defer r.lock.Unlock()
	return marshalJSON(r.values)
}

func (r *reader) Resolve() error {
	r.lock.Lock()
	defer r.lock.Unlock()
	return r.opts.resolver(r.values)
}

func cloneMap(src map[string]interface{}) (map[string]interface{}, error) {
	var buf bytes.Buffer
	gob.Register(map[string]interface{}{})
	gob.Register([]interface{}{})
	enc := gob.NewEncoder(&buf)
	dec := gob.NewDecoder(&buf)
	err := enc.Encode(src)
	if err != nil {
		return nil, err
	}
	var instance map[string]interface{}
	err = dec.Decode(&instance)
	if err != nil {
		return nil, err
	}
	return instance, nil
}

func readValue(values map[string]interface{}, path string) (Value, bool) {
	var (
		next = values
		keys = strings.Split(path, ".")
		last = len(keys) - 1
	)
	for idx, key := range keys {
		value, ok := next[key]
		if !ok {
			return nil, false
		}
		if idx == last {
			av := &atomicValue{}
			av.Store(value)
			return av, true
		}
		switch vm := value.(type) {
		case map[string]interface{}:
			next = vm
		default:
			return nil, false
		}
	}
	return nil, false
}

func marshalJSON(v interface{}) ([]byte, error) {
	if m, ok := v.(proto.Message); ok {
		return protojson.MarshalOptions{
			EmitUnpopulated: true,
		}.Marshal(m)
	}
	return json.Marshal(v)
}

func unmarshalJSON(data []byte, v interface{}) error {
	m, ok := v.(proto.Message)
	if ok {
		return protojson.UnmarshalOptions{DiscardUnknown: true}.Unmarshal(data, m)
	}
	return json.Unmarshal(data, v)
}
