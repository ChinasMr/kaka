package config

type Decoder func(*KeyValue, map[string]interface{}) error

type Resolver func(map[string]interface{}) error

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
	//if src.Format == "" {
	//	keys := strings.Split(src.Key, ".")
	//	for i, k := range keys {
	//		if i == len(keys) - 1 {
	//			target[k] = src.Value
	//		} else {
	//			sub := make(map[string]interface{})
	//			target[k] = sub
	//			target = sub
	//		}
	//	}
	//	return nil
	//}
	//if codec := enco
	return nil
}

func DefaultResolver(input map[string]interface{}) error {
	return nil
}
