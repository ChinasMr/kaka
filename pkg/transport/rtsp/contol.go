package rtsp

import (
	"github.com/ChinasMr/kaka/pkg/transport/rtsp/status"
	"github.com/google/uuid"
	"sync"
	"time"
)

var _ TransactionController = (*transactionController)(nil)

type TransactionController interface {
	Create(ch string, trans *transport) *transaction
	Delete(ch string, id string)
}

type transactionController struct {
	chs map[string]map[string]*transaction
	rwm *sync.RWMutex
}

func newTransactionController(chs ...string) TransactionController {
	tc := &transactionController{
		chs: map[string]map[string]*transaction{},
		rwm: &sync.RWMutex{},
	}
	for _, ch := range chs {
		if len(ch) == 0 {
			continue
		}
		tc.chs[ch] = make(map[string]*transaction)
	}
	return tc
}

func (t *transactionController) Create(ch string, trans *transport) *transaction {
	tx := newTx(trans)
	t.rwm.Lock()
	defer t.rwm.Unlock()
	txg, ok := t.chs[ch]
	if ok {
		txg[tx.id] = tx
	} else {
		t.chs[ch] = map[string]*transaction{tx.id: tx}
	}
	return tx
}

func (t *transactionController) Delete(ch string, id string) {
	// remove from collection.
	t.rwm.Lock()
	txg, ok := t.chs[ch]
	if !ok {
		return
	}
	c, ok := txg[id]
	if !ok {
		return
	}
	delete(txg, id)
	t.rwm.Unlock()

	// clear and reclaim.
	_ = c.Close()
	putTx(c)
}

var txPool = &sync.Pool{}

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
