package rtsp

import (
	"github.com/ChinasMr/kaka/pkg/log"
	"github.com/ChinasMr/kaka/pkg/transport/rtsp/header"
	"strings"
)

// Many methods in RTSP do not contribute to state.
// However, the following play a central role in defining
// the allocation and usage of stream resources on the server:
// SETUP, PLAY, RECORD, PAUSE, and TEARDOWN.

type HandlerFunc func(req Request, res Response, tx Transaction) error

// A minimal server implementation MUST be able to do the following:
// Implement the following methods: SETUP, TEARDOWN, OPTIONS and
// either PLAY (for a minimal playback server) or RECORD (for a
// minimal recording server). If RECORD is implemented, ANNOUNCE
// should be implemented as well.

type minimalHandler interface {
	OPTIONS(req Request, res Response, tx Transaction) error
	DESCRIBE(req Request, res Response, tx Transaction) error
	ANNOUNCE(req Request, res Response, tx Transaction) error
	SETUP(req Request, res Response, tx Transaction) error
	RECORD(req Request, res Response, tx Transaction) error
	PLAY(req Request, res Response, tx Transaction) error
	TEARDOWN(req Request, res Response, tx Transaction) error
}

type UnimplementedServerHandler struct {
	tc TransactionController
	hs []string
}

func (u *UnimplementedServerHandler) OPTIONS(req Request, res Response, tx Transaction) error {
	log.Debugf("options request url: %s", req.URL().String())
	res.SetHeader(header.Public, strings.Join(u.hs, ", "))
	return tx.Response(res)
}

func (u *UnimplementedServerHandler) ANNOUNCE(req Request, res Response, tx Transaction) error {
	log.Debugf("announce request url: %s", req.URL().String())
	sdp, err := req.ParseSDP()
	if err != nil {
		log.Errorf("can not pares sdp message: %v", err)
		return tx.Response(ErrInternal(res))
	}
	log.Debugf("source %s has media: %d", req.URL().String(), len(sdp.Medias))
	ch, ok := u.tc.GetCh(req.Channel())
	if !ok {
		return tx.Response(ErrInternal(res))
	}
	// occupy the channel with transactions id.
	ok = ch.Lock(tx.ID())
	if !ok {
		return tx.Response(ErrInternal(res))
	}
	ch.SetSDP(sdp, req.Body())
	return tx.Response(res)
}

func (u *UnimplementedServerHandler) DESCRIBE(req Request, res Response, tx Transaction) error {
	panic("implement me")
}

func (u *UnimplementedServerHandler) SETUP(req Request, res Response, tx Transaction) error {
	log.Debugf("setup request url: %s", req.URL().String())
	tr, _ := req.Transport()
	if tr.Multicast() {
		return tx.Response(ErrUnsupportedTransport(res))
	}

	ch, ok := u.tc.GetCh(req.Channel())
	if !ok {
		return tx.Response(ErrInternal(res))
	}
	for _, m := range ch.SDP().Medias {
		stream := m.Attribute("control")
		if stream == req.Stream() {
			// add tcp stream.
			if p1, p2, ok1 := tr.Interleaved(); ok1 && tr.LowerTransportTCP() {
				log.Debugf("rtp tpc channel: %d, rtcp tpc channel: %d", p1, p2)
				tx.AddMedia(&Media{
					interleaved: true,
					rtp:         p1,
					rtcp:        p2,
					control:     stream,
					record:      tr.Record(),
				})
				res.SetHeader(header.Transport,
					header.NewTransportHeader(header.LowerTransTCP,
						header.ParamUnicast,
						header.NewInterleavedParam(p1, p2)))
			}
			// add udp stream.

			// send response.
			err := tx.Response(res)
			if err != nil {
				return err
			}
			// refresh the session status.
			tx.Ready(ch.SDP())
			return nil
		}
	}

	return tx.Response(ErrInternal(res))
}

func (u *UnimplementedServerHandler) PLAY(req Request, res Response, tx Transaction) error {
	panic("implement me")
}

func (u *UnimplementedServerHandler) RECORD(req Request, res Response, tx Transaction) error {
	log.Debugf("record request url: %s", req.URL().String())
	ch, ok := u.tc.GetCh(req.Channel())
	if !ok {
		return tx.Response(ErrInternal(res))
	}
	ok = ch.Lock(tx.ID())
	if !ok {
		return tx.Response(ErrInternal(res))
	}
	ok = tx.Record(ch.SDP())
	if !ok {
		return tx.Response(ErrInternal(res))
	}
	err := tx.Response(res)
	if err != nil {
		return err
	}
	return tx.Serve(ch)
}

func (u *UnimplementedServerHandler) TEARDOWN(req Request, res Response, tx Transaction) error {
	panic("implement me")
}
