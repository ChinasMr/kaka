package rtsp

import "sync"

var _ TransactionOperator = (*transactionOperator)(nil)

type TransactionOperator interface {
	GetTx(id string) *transaction
	DeleteTx(id string)
}

type transactionOperator struct {
	txs map[string]*transaction
	rwm sync.RWMutex
}

func NewTxOperator() TransactionOperator {
	return &transactionOperator{
		txs: map[string]*transaction{},
		rwm: sync.RWMutex{},
	}
}

func (t *transactionOperator) GetTx(id string) *transaction {
	t.rwm.RLock()
	tx, ok := t.txs[id]
	t.rwm.RUnlock()
	if ok {
		return tx
	}
	tx = newTransaction()
	t.rwm.Lock()
	t.txs[tx.id] = tx
	t.rwm.Unlock()
	return tx
}

func (t *transactionOperator) DeleteTx(id string) {
	t.rwm.Lock()
	defer t.rwm.Unlock()
	delete(t.txs, id)
}
