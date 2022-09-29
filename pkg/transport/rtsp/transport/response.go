package transport

type Response interface {
	Code(c uint64)
	SetHeader(key string, value ...string)
	SetContent(body []byte)
}
