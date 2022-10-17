package rtsp

import (
	"github.com/ChinasMr/kaka/pkg/transport/rtsp/status"
	"github.com/google/uuid"
	"net"
	"sync"
)

var _ TransactionController = (*transactionController)(nil)

type TransactionController interface {
	CreateTx(trans *transport, rf *rtcpFamily) *transaction
	DeleteTx(id *transaction)
	GetCh(ch string) (Channel, bool)
	GetOrCreateCh(ch string) Channel
	Forward(p *Package, addr *net.UDPAddr)
}

type forwarder struct {
	addr   *net.UDPAddr
	order  int
	input  chan *Package
	output chan *Package
}

func (t *transactionController) serve(f *forwarder) {
	for p := range f.input {
		if f.output != nil {
			p.Order = f.order
			f.output <- p
			continue
		}
		// try to find the corrected channel.
		ok := func() bool {
			for _, ch := range t.chs {
				if !f.addr.IP.Equal(ch.Source().IP()) {
					continue
				}
				for _, m := range ch.Source().Medias() {
					if m.interleaved || !m.record {
						continue
					}
					if f.addr.Port == m.rtp || f.addr.Port == m.rtcp {
						f.output = ch.Input()
						f.order = m.order
						return true
					}
				}
			}
			return false
		}()
		if ok {
			f.output <- p
		} else {
			// can not find relative ch, return.
			// todo need clear relative resource.
		}
	}
}

type transactionController struct {
	chs        map[string]Channel
	rwm        sync.RWMutex
	forwarders map[string]*forwarder
}

func (t *transactionController) Forward(p *Package, addr *net.UDPAddr) {
	f, ok := t.forwarders[addr.String()]
	if ok {
		f.input <- p
		return
	}
	f = &forwarder{addr: addr, input: make(chan *Package, 5), output: nil}
	t.forwarders[addr.String()] = f
	go t.serve(f)
	f.input <- p
}

func (t *transactionController) GetCh(ch string) (Channel, bool) {
	t.rwm.RLock()
	defer t.rwm.RUnlock()
	rv, ok := t.chs[ch]
	return rv, ok
}

func (t *transactionController) GetOrCreateCh(ch string) Channel {
	t.rwm.Lock()
	defer t.rwm.Unlock()
	rv, ok := t.chs[ch]
	if ok {
		return rv
	}
	nc := NewChannel(ch)
	t.chs[ch] = nc
	return nc
}

func newTransactionController(chs ...string) TransactionController {
	tc := &transactionController{
		rwm:        sync.RWMutex{},
		chs:        map[string]Channel{},
		forwarders: map[string]*forwarder{},
	}
	for _, ch := range chs {
		if len(ch) == 0 {
			continue
		}
		tc.chs[ch] = NewChannel(ch)
	}
	return tc
}

func (t *transactionController) CreateTx(trans *transport, rf *rtcpFamily) *transaction {
	id, _ := uuid.NewUUID()
	tx := txPool.Get().(*transaction)
	tx.id = id.String()
	tx.transport = trans
	tx.rf = rf
	return tx
}

func (t *transactionController) DeleteTx(tx *transaction) {
	for _, ch := range t.chs {
		_ = ch.Teardown(tx)
	}
	_ = tx.Close()

	tx.id = ""
	tx.state = status.INIT
	tx.transport = nil
	tx.rf = nil
	tx.medias = map[string]*Media{}
	tx.rwm = sync.RWMutex{}
	tx.interleaved = false
	txPool.Put(tx)
}

var txPool = &sync.Pool{
	New: func() any {
		return &transaction{
			id:          "",
			state:       status.INIT,
			transport:   nil,
			rf:          nil,
			medias:      map[string]*Media{},
			interleaved: false,
			rwm:         sync.RWMutex{},
		}
	},
}
