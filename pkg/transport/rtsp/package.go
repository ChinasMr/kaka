package rtsp

type Package struct {
	Ch          int
	Len         uint32
	Interleaved bool
	Data        []byte
}
