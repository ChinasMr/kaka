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
	p.Len = 0
	p.Ch = 0
	p.Interleaved = false
	packagePool.Put(p)
}

type Package struct {
	Ch          int
	Len         uint32
	Interleaved bool
	Data        []byte
}
