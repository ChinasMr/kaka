package config

import (
	"fmt"
	"github.com/ChinasMr/kaka/internal/encoding"
	"regexp"
	"strings"
)

// Decoder is config decoder.
type Decoder func(*KeyValue, map[string]interface{}) error

// Resolver resolve placeholder in config.
type Resolver func(map[string]interface{}) error

// Option is config option.
type Option func(*options)

type options struct {
	sources  []Source
	decoder  Decoder
	resolver Resolver
}

func WithSource(s ...Source) Option {
	return func(o *options) {
		o.sources = s
	}
}

func DefaultDecoder(src *KeyValue, target map[string]interface{}) error {
	if src.Format == "" {
		keys := strings.Split(src.Key, ".")
		for i, k := range keys {
			if i == len(keys)-1 {
				target[k] = src.Value
			} else {
				sub := make(map[string]interface{})
				target[k] = sub
				target = sub
			}
		}
		return nil
	}
	// get a decoder
	codec := encoding.GetCodec(src.Format)
	if codec != nil {
		return codec.Unmarshal(src.Value, &target)
	}
	return fmt.Errorf("unsupported key: %s format: %s", src.Key, src.Format)
}

func DefaultResolver(input map[string]interface{}) error {
	mapper := func(name string) string {
		args := strings.SplitN(strings.TrimSpace(name), ":", 2)
		if v, has := readValue(input, args[0]); has {
			s, _ := v.String()
			return s
		} else if len(args) > 1 {
			return args[1]
		}
		return ""
	}

	var resolve func(map[string]interface{}) error
	resolve = func(sub map[string]interface{}) error {
		for k, v := range sub {
			switch vt := v.(type) {
			case string:
				sub[k] = expand(vt, mapper)
			case map[string]interface{}:
				if err := resolve(vt); err != nil {
					return err
				}
			case []interface{}:
				for i, iface := range vt {
					switch it := iface.(type) {
					case string:
						vt[i] = expand(it, mapper)
					case map[string]interface{}:
						if err := resolve(it); err != nil {
							return err
						}
					}
				}
				sub[k] = vt
			}
		}
		return nil
	}
	return resolve(input)
}

func expand(s string, mapping func(string) string) string {
	r := regexp.MustCompile(`\${(.*?)}`)
	re := r.FindAllStringSubmatch(s, -1)
	for _, i := range re {
		if len(i) == 2 { //nolint:gomnd
			s = strings.ReplaceAll(s, i[0], mapping(i[1]))
		}
	}
	return s
}
