package rtsp

import (
	"encoding/binary"
	"fmt"
	"github.com/ChinasMr/kaka/pkg/transport/rtsp/status"
	"io"
	"sync"
	"time"
)

var _ Transaction = (*transaction)(nil)

type Transaction interface {
	ID() string
	Status() status.Status
	SetStatus(s status.Status)
	Interleaved() bool
	SetInterleaved()
	Forward(data *Package, wg *sync.WaitGroup)
	Response(res Response) error
	Close() error
}

type transaction struct {
	id          string
	state       status.Status
	interleaved bool
	timeout     time.Duration
	transport   Transport
}

func (t *transaction) Response(res Response) error {
	return t.transport.Write(res.Encoding())
}

func (t *transaction) Close() error {
	return t.transport.Close()
}

func (t *transaction) Forward(data *Package, wg *sync.WaitGroup) {

}

func (t *transaction) ID() string {
	return t.id
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

func (t *transaction) Status() status.Status {
	return t.state
}

func (t *transaction) writeInterleavedFrame(channel int, frame []byte) error {
	buf := make([]byte, 2048)
	buf[0] = 0x24
	buf[1] = byte(channel)
	binary.BigEndian.PutUint16(buf[2:], uint16(len(frame)))
	n := copy(buf[4:], frame)
	return t.transport.Write(buf[:4+n])
}

func (t *transaction) ReadInterleavedFrame(frame []byte) (int, uint32, error) {
	// interleavedHeader example
	// Magic:0x24   bytes 1
	// Channel:0x01 bytes 2
	// Length:84    bytes 3-4
	interleavedHeader := make([]byte, 4)
	conn := t.transport.Conn()
	_, err := io.ReadFull(conn, interleavedHeader)
	if err != nil {
		return -1, 0, err
	}

	if interleavedHeader[0] == 0x54 {
		return -1, 0, io.EOF
	}

	if interleavedHeader[1] == 0x24 {
		return -1, 0, fmt.Errorf("magic byte error")
	}

	frameLen := binary.BigEndian.Uint16(interleavedHeader[2:])
	if frameLen > 2048 {
		return -1, 0, fmt.Errorf("freame len greater than 2048")
	}
	_, err = io.ReadFull(conn, frame[:frameLen])
	if err != nil {
		return -1, 0, err
	}
	return int(interleavedHeader[1]), uint32(frameLen), nil
}
