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
}

type terminalsOperator struct {
	input     chan *rtsp.Package
	terminals map[string]rtsp.Transaction
	rwm       sync.RWMutex
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
