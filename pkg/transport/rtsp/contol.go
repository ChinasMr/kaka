package rtsp

import (
	"github.com/ChinasMr/kaka/pkg/transport/rtsp/status"
	"github.com/google/uuid"
	"gortc.io/sdp"
	"sync"
)

var _ TransactionController = (*transactionController)(nil)

type Channel interface {
	SetSDP(sdp *sdp.Message, raw []byte)
	SDP() *sdp.Message
}

// A server MAY refuse to change parameters of an existing stream.

type channel struct {
	name string
	txs  map[string]*transaction
	rwm  sync.RWMutex
	sdp  *sdp.Message
	raw  []byte
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
	// todo they maybe a call back function.
}

func NewChannel(ch string) Channel {
	return &channel{
		name: ch,
		txs:  map[string]*transaction{},
		rwm:  sync.RWMutex{},
		sdp:  nil,
		raw:  nil,
	}
}

type TransactionController interface {
	CreateTx(trans *transport) *transaction
	DeleteTx(id string)
	GetCh(ch string) (Channel, bool)
	GetOrCreateCh(ch string) Channel
}

type transactionController struct {
	chs map[string]Channel
	txs map[string]*transaction
	rwm sync.RWMutex
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
		rwm: sync.RWMutex{},
		txs: map[string]*transaction{},
		chs: map[string]Channel{},
	}
	for _, ch := range chs {
		if len(ch) == 0 {
			continue
		}
		tc.chs[ch] = NewChannel(ch)
	}
	return tc
}

func (t *transactionController) CreateTx(trans *transport) *transaction {
	tx := newTx(trans)
	t.rwm.Lock()
	t.txs[tx.id] = tx
	t.rwm.Unlock()
	return tx
}

func (t *transactionController) DeleteTx(id string) {
	t.rwm.Lock()
	tx, ok := t.txs[id]
	if !ok {
		return
	}
	delete(t.txs, tx.id)
	t.rwm.Unlock()
	_ = tx.Close()
	putTx(tx)
}

var txPool = &sync.Pool{
	New: func() any {
		return &transaction{
			id:        "",
			state:     status.INIT,
			transport: nil,
			medias:    map[string]*Media{},
		}
	},
}

func newTx(trans *transport) *transaction {
	id, _ := uuid.NewUUID()
	tx := txPool.Get().(*transaction)
	tx.id = id.String()
	tx.transport = trans
	return tx
}

func putTx(tx *transaction) {
	tx.id = ""
	tx.state = status.INIT
	tx.transport = nil
	tx.medias = map[string]*Media{}
	txPool.Put(tx)
}
