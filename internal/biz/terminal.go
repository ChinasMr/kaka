package biz

import (
	"github.com/ChinasMr/kaka/pkg/transport/rtsp"
	"github.com/ChinasMr/kaka/pkg/transport/rtsp/status"
	"sync"
)

type TerminalsOperator interface {
	Add(tx rtsp.Transaction)
	Input() chan *rtsp.Package
	Num() int
	ListTx() []rtsp.Transaction
	Malloc() *rtsp.Package
}

type terminalsOperator struct {
	input     chan *rtsp.Package
	terminals map[string]rtsp.Transaction
	rwm       sync.RWMutex
	pool      *sync.Pool
}

func (t *terminalsOperator) Malloc() *rtsp.Package {
	return t.pool.Get().(*rtsp.Package)
}

func NewTerminalsOperator() TerminalsOperator {
	operator := &terminalsOperator{
		input:     make(chan *rtsp.Package, 2),
		terminals: map[string]rtsp.Transaction{},
		rwm:       sync.RWMutex{},
		pool:      &sync.Pool{New: func() any { return &rtsp.Package{Data: make([]byte, 2048)} }},
	}
	go operator.serve()
	return operator
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
	for pack := range t.input {
		p := pack
		wg := &sync.WaitGroup{}
		t.rwm.RLock()
		for _, terminal := range t.terminals {
			if terminal.Status() == status.PLAYING {
				wg.Add(1)
				terminal.Forward(p, wg)
			}
		}
		go func() {
			wg.Wait()
			p.Len = 0
			p.Ch = 0
			t.pool.Put(p)
		}()
		t.rwm.RUnlock()
	}
}
