package biz

import (
	"github.com/ChinasMr/kaka/pkg/log"
	"github.com/ChinasMr/kaka/pkg/transport/rtsp"
	"github.com/ChinasMr/kaka/pkg/transport/rtsp/status"
	"sync"
)

type TerminalsOperator interface {
	Add(tx rtsp.Transaction)
	Input() chan *rtsp.Package
	Num() int
	ListTx() []rtsp.Transaction
}

type terminalsOperator struct {
	input     chan *rtsp.Package
	terminals map[string]rtsp.Transaction
	rwm       sync.RWMutex
}

func (t *terminalsOperator) ListTx() []rtsp.Transaction {
	t.rwm.RLock()
	defer t.rwm.RUnlock()
	rv := make([]rtsp.Transaction, 0, len(t.terminals))
	for _, tx := range t.terminals {
		rv = append(rv, tx)
	}
	return rv
}

func (t *terminalsOperator) Num() int {
	return len(t.terminals)
}

func (t *terminalsOperator) Input() chan *rtsp.Package {
	return t.input
}

func (t *terminalsOperator) Add(tx rtsp.Transaction) {
	t.rwm.Lock()
	defer t.rwm.Unlock()
	t.terminals[tx.ID()] = tx
}

func (t *terminalsOperator) serve() {
	for {
		data := <-t.input
		t.rwm.RLock()
		for _, ter := range t.terminals {
			if ter.Status() == status.PLAYING {
				//go func() {
				err := ter.Forward(data)
				if err != nil {
					log.Errorf("can not forward: %v", err)
				}
				//}()
			}
		}
		t.rwm.RUnlock()
		// todo donation
	}
}

func NewTerminalsOperator(ch chan *rtsp.Package) TerminalsOperator {
	nt := &terminalsOperator{
		input:     ch,
		terminals: map[string]rtsp.Transaction{},
		rwm:       sync.RWMutex{},
	}
	go nt.serve()
	return nt
}
