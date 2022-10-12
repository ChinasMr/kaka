package rtsp

import (
	"encoding/binary"
	"fmt"
	"github.com/ChinasMr/kaka/pkg/transport/rtsp/status"
	"gortc.io/sdp"
	"io"
	"sync"
)

var _ Transaction = (*transaction)(nil)

type Media struct {
	control     string
	interleaved bool
	record      bool
	rtp         int64
	rtcp        int64
}

type Transaction interface {
	ID() string
	Status() status.Status
	Forward(p *Package, wg *sync.WaitGroup) error
	Response(res Response) error
	Request(req Request) error
	Medias() map[string]*Media
	AddMedia(media *Media)
	Ready(sdp *sdp.Message) bool
	Record(sdp *sdp.Message) bool
	RecordServe(ch Channel) error
	Close() error
}

type transaction struct {
	id          string
	state       status.Status
	transport   Transport
	medias      map[string]*Media
	rwm         sync.RWMutex
	interleaved bool
}

func (t *transaction) RecordServe(ch Channel) error {
	if t.interleaved {
		input := ch.Input()
		for {
			p := ch.Package()
			c, l, err := t.ReadInterleavedFrame(p.Data)
			if err != nil {
				return err
			}
			p.Len = l
			p.Ch = c
			p.Interleaved = true
			input <- p
		}
	} else {
		// todo register fast route for udp trans.

		return nil
	}
}

func (t *transaction) Record(sdp *sdp.Message) bool {
	interleaved := true
	for _, m := range t.medias {
		interleaved = m.interleaved
		break
	}
	for _, m := range sdp.Medias {
		s := m.Attribute("control")
		media, ok := t.medias[s]
		if !ok {
			return false
		}
		if media.record != true {
			return false
		}
		if media.interleaved != interleaved {
			return false
		}
	}
	t.interleaved = interleaved
	t.setStatus(status.RECORDING)
	return true
}

func (t *transaction) Ready(sdp *sdp.Message) bool {
	for _, m := range sdp.Medias {
		s := m.Attribute("control")
		_, ok := t.medias[s]
		if !ok {
			return false
		}
	}

	t.setStatus(status.READY)
	return true
}

func (t *transaction) Medias() map[string]*Media {
	t.rwm.RLock()
	defer t.rwm.RUnlock()
	return t.medias
}

func (t *transaction) AddMedia(media *Media) {
	t.rwm.Lock()
	t.rwm.Unlock()
	t.medias[media.control] = media
}

func (t *transaction) Request(req Request) error {
	return t.transport.Write(req.Encode())
}

func (t *transaction) Response(res Response) error {
	return t.transport.Write(res.Encoding())
}

func (t *transaction) Close() error {
	return t.transport.Close()
}

func (t *transaction) Forward(p *Package, wg *sync.WaitGroup) error {
	defer wg.Done()
	if p.Interleaved {
		return t.writeInterleavedFrame(p.Ch, p.Data[:p.Len])
	} else {
		return nil
	}
}

func (t *transaction) ID() string {
	return t.id
}

func (t *transaction) setStatus(s status.Status) {
	t.rwm.Lock()
	t.rwm.Unlock()
	t.state = s
}

func (t *transaction) Status() status.Status {
	t.rwm.RLock()
	t.rwm.RUnlock()
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
