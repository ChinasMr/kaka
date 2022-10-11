package rtsp

import (
	"github.com/ChinasMr/kaka/pkg/transport/rtsp/status"
	"github.com/google/uuid"
	"sync"
	"time"
)

var _ TransactionController = (*transactionController)(nil)

type TransactionController interface {
	Create(trans *transport) *transaction
	Delete(id string)
}

type transactionController struct {
	chs map[string]map[string]*transaction
	txs map[string]*transaction
	rwm *sync.RWMutex
}

func newTransactionController(chs ...string) TransactionController {
	tc := &transactionController{
		rwm: &sync.RWMutex{},
		txs: map[string]*transaction{},
	}
	for _, ch := range chs {
		if len(ch) == 0 {
			continue
		}
		tc.chs[ch] = make(map[string]*transaction)
	}
	return tc
}

func (t *transactionController) Create(trans *transport) *transaction {
	tx := newTx(trans)
	t.rwm.Lock()
	t.txs[tx.id] = tx
	t.rwm.Unlock()
	return tx
}

func (t *transactionController) Delete(id string) {
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

var txPool = &sync.Pool{New: func() any { return &transaction{} }}

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
	tx.interleaved = false
	tx.timeout = 3 * time.Second
	tx.transport = nil
	txPool.Put(tx)
}
