package encoding

import "strings"

type Codec interface {
	Marshal(v interface{}) ([]byte, error)
	Unmarshal(data []byte, v interface{}) error
	Name() string
}

var registeredCodecs = make(map[string]Codec)

func RegisterCodec(codec Codec) {
	if codec == nil {
		panic("can not register a nil Codec")
	}
	if codec.Name() == "" {
		panic("can not register Codec with empty string result for Name()")
	}
	contentSubtype := strings.ToLower(codec.Name())
	registeredCodecs[contentSubtype] = codec
}

// GetCodec gets a registered Codec by content subtype.
func GetCodec(contentSubtype string) Codec {
	return registeredCodecs[strings.ToLower(contentSubtype)]
}
