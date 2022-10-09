package rtsp

import (
	"github.com/ChinasMr/kaka/pkg/transport/rtsp/status"
	"github.com/google/uuid"
	"sync"
)

type TransactionOperator interface {
	GetTx(id string) *Transaction
	DeleteTx(id string)
}

type transactionOperator struct {
	txs map[string]*Transaction
	rwm sync.RWMutex
}

func NewTxOperator() TransactionOperator {
	return &transactionOperator{
		txs: map[string]*Transaction{},
		rwm: sync.RWMutex{},
	}
}

func (t *transactionOperator) GetTx(id string) *Transaction {
	t.rwm.RLock()
	tx, ok := t.txs[id]
	t.rwm.RUnlock()
	if ok {
		return tx
	}
	tx = newTransaction()
	t.rwm.Lock()
	t.txs[tx.Id] = tx
	t.rwm.Unlock()
	return tx
}

func (t *transactionOperator) DeleteTx(id string) {
	t.rwm.Lock()
	defer t.rwm.Unlock()
	delete(t.txs, id)
}

type Transaction struct {
	Id    string
	State status.Status
	Mu    sync.Mutex
}

func newTransaction() *Transaction {
	id, _ := uuid.NewUUID()
	return &Transaction{
		Id:    id.String(),
		State: status.INIT,
		Mu:    sync.Mutex{},
	}
}
