package rtsp

import (
	"github.com/ChinasMr/kaka/pkg/transport/rtsp/status"
	"gortc.io/sdp"
	"io"
	"sync"
)

var _ Channel = (*channel)(nil)

type Channel interface {
	SetSDP(tx Transaction, sdp *sdp.Message, raw []byte) bool
	SDP() *sdp.Message
	Raw() []byte
	Lock(id string) bool
	Play(tx Transaction) error
	Record(tx Transaction) error
	Teardown(tx Transaction) error
}

func NewChannel(ch string) Channel {
	rv := &channel{
		name:   ch,
		txs:    map[string]Transaction{},
		rwm:    sync.RWMutex{},
		sdp:    &sdp.Message{},
		raw:    nil,
		source: "",
		input:  make(chan *Package, 2),
		pool: sync.Pool{
			New: func() any {
				return &Package{
					Data: make([]byte, 2048),
				}
			},
		},
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

func (c *channel) Teardown(tx Transaction) error {
	if tx.Status() == status.PLAYING {
		delete(c.txs, tx.ID())
		return nil
	}
	if tx.Status() == status.RECORDING {
		ok := c.Lock(tx.ID())
		if ok {
			// the register wanna to deregister the info.
			c.reset()
		}
	}
	// clear the status. avoid clear twice.
	tx.PreInit()
	return nil
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
	// add the tx to the channel.
	c.txs[tx.ID()] = tx

	if tx.Interleaved() {
		// blocking the connection.
		buf := make([]byte, 2048)
		for {
			_, err := tx.Read(buf)
			if err != nil {
				// delete the tx to the channel.
				delete(c.txs, tx.ID())
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
	for {
		select {
		// receive data packet.
		case p := <-c.input:
			pack := p
			wg := &sync.WaitGroup{}
			for _, tx := range c.txs {
				if tx.Status() == status.PLAYING {
					wg.Add(1)
					go tx.Forward(pack, wg)
				}
			}
			go func() {
				wg.Wait()
				c.putPackage(pack)
				// todo refresh the live keeper.
			}()
		}
	}
}

func (c *channel) reset() {
	c.rwm.Lock()
	c.sdp = &sdp.Message{}
	c.raw = nil
	c.source = ""
	c.rwm.Unlock()
}

func (c *channel) Lock(id string) bool {
	c.rwm.Lock()
	defer c.rwm.Unlock()
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

func (c *channel) SetSDP(tx Transaction, sdp *sdp.Message, raw []byte) bool {
	ok := c.Lock(tx.ID())
	if !ok {
		return false
	}
	c.rwm.Lock()
	c.sdp = sdp
	c.raw = raw
	c.rwm.Unlock()
	return true
	// A server MAY refuse to change parameters of an existing stream.
	// todo they maybe a call back function.
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
