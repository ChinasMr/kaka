package rtsp

import "sync"

var packagePool = sync.Pool{
	New: func() any {
		return &Package{Data: make([]byte, 2048)}
	},
}

func newPackage() *Package {
	return packagePool.Get().(*Package)
}

func putPackage(p *Package) {
	p.Ch = 0
	p.Len = 0
	p.Order = 0
	p.Interleaved = false
	packagePool.Put(p)
}

type Package struct {
	Ch          int
	Len         uint32
	Order       int
	Interleaved bool
	Data        []byte
}
