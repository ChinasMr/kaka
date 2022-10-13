package rtsp

import (
	"github.com/ChinasMr/kaka/pkg/transport/rtsp/status"
	"gortc.io/sdp"
	"io"
	"sync"
)

var _ Channel = (*channel)(nil)

type Channel interface {
	SetSDP(sdp *sdp.Message, raw []byte)
	SDP() *sdp.Message
	Raw() []byte
	Lock(id string) bool
	Play(tx Transaction) error
	Record(tx Transaction) error
}

func NewChannel(ch string) Channel {
	rv := &channel{
		name:   ch,
		txs:    map[string]Transaction{},
		rwm:    sync.RWMutex{},
		sdp:    nil,
		raw:    nil,
		source: "",
		pool: sync.Pool{
			New: func() any {
				return &Package{
					Data: make([]byte, 2048),
				}
			},
		},
		input: make(chan *Package, 2),
	}
	go rv.serve()
	return rv
}

type channel struct {
	name   string
	txs    map[string]Transaction
	rwm    sync.RWMutex
	sdp    *sdp.Message
	raw    []byte
	source string
	pool   sync.Pool
	input  chan *Package
}

func (c *channel) Record(tx Transaction) error {
	if tx.Interleaved() {
		for {
			p := c.newPackage()
			ch, l, err := tx.ReadInterleavedFrame(p.Data)
			if err != nil {
				return err
			}
			p.Len = l
			p.Ch = ch
			p.Interleaved = true
			c.input <- p
		}
	} else {
		return nil
	}
}

func (c *channel) Play(tx Transaction) error {
	c.rwm.Lock()
	c.txs[tx.ID()] = tx
	c.rwm.Unlock()

	if tx.Interleaved() {
		// blocking the connection.
		buf := make([]byte, 2048)
		for {
			_, err := tx.Read(buf)
			if err != nil {
				return io.EOF
			}
		}
	}
	return nil
}

func (c *channel) Raw() []byte {
	return c.raw
}

func (c *channel) serve() {
	for p := range c.input {
		pack := p
		wg := &sync.WaitGroup{}
		c.rwm.RLock()
		for _, tx := range c.txs {
			if tx.Status() == status.PLAYING {
				wg.Add(1)
				go tx.Forward(pack, wg)
			}
		}
		c.rwm.RUnlock()
		go func() {
			wg.Wait()
			c.putPackage(pack)
			// todo refresh the live keeper.
		}()
	}
}

func (c *channel) unlock() {
	c.source = ""
}

func (c *channel) newPackage() *Package {
	return c.pool.Get().(*Package)
}

func (c *channel) putPackage(p *Package) {
	p.Len = 0
	p.Ch = 0
	p.Interleaved = false
	c.pool.Put(p)
}

func (c *channel) Lock(id string) bool {
	if c.source == "" {
		c.source = id
		return true
	} else {
		if c.source == id {
			return true
		}
	}
	return false
}

func (c *channel) SDP() *sdp.Message {
	c.rwm.RLock()
	defer c.rwm.RUnlock()
	return c.sdp
}

func (c *channel) SetSDP(sdp *sdp.Message, raw []byte) {
	c.rwm.Lock()
	c.sdp = sdp
	c.raw = raw
	c.rwm.Unlock()
	// A server MAY refuse to change parameters of an existing stream.
	// todo they maybe a call back function.
}
