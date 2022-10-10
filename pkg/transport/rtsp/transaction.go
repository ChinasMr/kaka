package rtsp

import (
	"github.com/ChinasMr/kaka/pkg/transport/rtsp/status"
	"github.com/google/uuid"
	"sync"
	"time"
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
	Forward(data *Package, wg *sync.WaitGroup)
	Close()
}

type transaction struct {
	id          string
	state       status.Status
	medias      map[string]bool
	interleaved bool
	trans       Transport
	timeout     time.Duration
}

func newTransaction() *transaction {
	id, _ := uuid.NewUUID()
	tx := &transaction{
		id:          id.String(),
		state:       status.INIT,
		medias:      map[string]bool{},
		interleaved: false,
		trans:       nil,
		timeout:     3 * time.Second,
	}
	go tx.serve()
	return tx
}

func (t *transaction) serve() {
	for {

	}
}

func (t *transaction) Close() {

}

func (t *transaction) Forward(data *Package, wg *sync.WaitGroup) {
	go func() {
		if t.interleaved {
			_ = t.trans.SendData(data.Ch, data.Data[:data.Len])
		}
		wg.Done()
	}()
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
