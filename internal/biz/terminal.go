package biz

import (
	"github.com/ChinasMr/kaka/pkg/log"
	"github.com/ChinasMr/kaka/pkg/transport/rtsp"
	"sync"
)

type TerminalsOperator interface {
}

type terminalsOperator struct {
	input     chan []byte
	terminals []*rtsp.Transaction
	rwm       sync.RWMutex
}

func (t *terminalsOperator) serve() {
	for {
		data := <-t.input
		log.Debugf("live rome received rtp/rtcp data: %d bytes", len(data))
		// todo donation
	}
}

func NewTerminalsOperator(ch chan []byte) TerminalsOperator {
	nt := &terminalsOperator{
		input:     ch,
		terminals: nil,
		rwm:       sync.RWMutex{},
	}
	go nt.serve()
	return &nt
}
