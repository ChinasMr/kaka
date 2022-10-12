package header

import (
	"fmt"
	"strings"
)

var name = ""

type TransportHeader map[string]struct{}

func (t TransportHeader) Has(keys ...string) bool {
	for _, key := range keys {
		_, ok := t[key]
		if !ok {
			return false
		}
	}
	return true
}

func (t TransportHeader) Value(k string) string {
	prefix := fmt.Sprintf("%s=", k)
	for key := range t {
		if strings.HasPrefix(key, prefix) {
			return key[len(prefix):]
		}
	}
	return ""
}

func (t TransportHeader) Type() {

}
