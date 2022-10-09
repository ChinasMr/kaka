package rtsp

import (
	"github.com/ChinasMr/kaka/pkg/transport/rtsp/status"
	"github.com/google/uuid"
	"sync"
)

var _ Transaction = (*transaction)(nil)

type Transaction interface {
	ID() string
	Status() status.Status
	SetStatus(s status.Status)
	AddMedia(media string)
	Interleaved() bool
	SetInterleaved()
	Medias() int
	Transport() Transport
}

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

type transaction struct {
	id          string
	state       status.Status
	medias      map[string]bool
	interleaved bool
	trans       Transport
}

func (t *transaction) ID() string {
	return t.id
}

func (t *transaction) Transport() Transport {
	return t.trans
}

func (t *transaction) Interleaved() bool {
	return t.interleaved
}

func (t *transaction) SetInterleaved() {
	t.interleaved = true
}

func (t *transaction) SetStatus(s status.Status) {
	t.state = s
}

func (t *transaction) Medias() int {
	return len(t.medias)
}

func (t *transaction) AddMedia(media string) {
	t.medias[media] = true
}

func (t *transaction) Status() status.Status {
	return t.state
}

func newTransaction() *transaction {
	id, _ := uuid.NewUUID()
	return &transaction{
		id:          id.String(),
		state:       status.INIT,
		medias:      map[string]bool{},
		interleaved: false,
		trans:       nil,
	}
}
