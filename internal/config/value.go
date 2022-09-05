package config

import (
	"fmt"
	"reflect"
	"strconv"
	"sync/atomic"
	"time"
)

type Value interface {
	Bool() (bool, error)
	Int() (int64, error)
	Float() (float64, error)
	String() (string, error)
	Duration() (time.Duration, error)
	Slice() ([]Value, error)
	Map() (map[string]Value, error)
	Scan(interface{}) error
	Load() interface{}
	Store(interface{})
}

type atomicValue struct {
	atomic.Value
}

func (v *atomicValue) Bool() (bool, error) {
	switch val := v.Load().(type) {
	case bool:
		return val, nil
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, string:
		return strconv.ParseBool(fmt.Sprint(val))
	}
	return false, fmt.Errorf("type assert to %v failed", reflect.TypeOf(v.Load()))
}

func (v *atomicValue) Int() (int64, error) {
	switch val := v.Load().(type) {
	case int:
		return int64(val), nil
	case int8:
		return int64(val), nil
	case int16:
		return int64(val), nil
	case int32:
		return int64(val), nil
	case int64:
		return val, nil
	case uint:
		return int64(val), nil
	case uint8:
		return int64(val), nil
	case uint16:
		return int64(val), nil
	case uint32:
		return int64(val), nil
	case uint64:
		return int64(val), nil
	case float32:
		return int64(val), nil
	case float64:
		return int64(val), nil
	case string:
		return strconv.ParseInt(val, 10, 64)
	}
	return 0, fmt.Errorf("type assert to %v failed", reflect.TypeOf(v.Load()))
}

func (v *atomicValue) Float() (float64, error) {
	//TODO implement me
	panic("implement me")
}

func (v *atomicValue) String() (string, error) {
	switch val := v.Load().(type) {
	case string:
		return val, nil
	case bool, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return fmt.Sprint(val), nil
	case []byte:
		return string(val), nil
	default:
		if s, ok := val.(fmt.Stringer); ok {
			return s.String(), nil
		}
	}
	return "", fmt.Errorf("type assert to %v failed", reflect.TypeOf(v.Load()))
}

func (v *atomicValue) Duration() (time.Duration, error) {
	//TODO implement me
	panic("implement me")
}

func (v *atomicValue) Slice() ([]Value, error) {
	//TODO implement me
	panic("implement me")
}

func (v *atomicValue) Map() (map[string]Value, error) {
	//TODO implement me
	panic("implement me")
}

func (v *atomicValue) Scan(i interface{}) error {
	//TODO implement me
	panic("implement me")
}
