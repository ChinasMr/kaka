package rtsp

import (
	"github.com/ChinasMr/kaka/pkg/log"
	"github.com/ChinasMr/kaka/pkg/transport/rtsp/header"
	"io"
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
	ok = ch.SetSDP(tx, sdp, req.Body())
	if !ok {
		return tx.Response(ErrInternal(res))
	}
	return tx.Response(res)
}

func (u *UnimplementedServerHandler) DESCRIBE(req Request, res Response, tx Transaction) error {
	log.Debugf("describe request url: %s", req.URL().String())
	ch, ok := u.tc.GetCh(req.Channel())
	if !ok {
		return tx.Response(ErrInternal(res))
	}
	raw := ch.Raw()
	if len(raw) == 0 {
		return tx.Response(ErrInternal(res))
	}
	res.SetHeader(header.ContentType, header.ContentTypeSDP)
	res.SetBody(raw)
	return tx.Response(res)
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
						header.NewInterleavedParam(p1, p1),
					))
			}
			// add udp stream.
			if p1, p2, ok1 := tr.ClientPort(); ok1 && !tr.LowerTransportTCP() {
				log.Debugf("rtp udp port: %d, rtcp udp port: %d", p1, p2)
				tx.AddMedia(&Media{
					interleaved: false,
					rtp:         p1,
					rtcp:        p2,
					control:     stream,
					record:      tr.Record(),
				})
				res.SetHeader(header.Transport,
					header.NewTransportHeader(header.LoweTransUDP,
						header.ParamUnicast,
						header.NewClientPort(p1, p2),
						header.NewServerPort(tx.RTP(), tx.RTCP()),
					))
			}

			// send response.
			err := tx.Response(res)
			if err != nil {
				return err
			}
			// refresh the session status.
			ok = tx.PreReady(ch.SDP())
			if ok {
				log.Debugf("session setup complete")
			}
			return nil
		}
	}

	return tx.Response(ErrInternal(res))
}

func (u *UnimplementedServerHandler) PLAY(req Request, res Response, tx Transaction) error {
	log.Debugf("play request url: %s", req.URL().String())
	ch, ok := u.tc.GetCh(req.Channel())
	if !ok {
		return tx.Response(ErrInternal(res))
	}
	ok = tx.PrePlay(ch.SDP())
	if !ok {
		return tx.Response(ErrInternal(res))
	}
	err := tx.Response(res)
	if err != nil {
		return err
	}

	// just return io.EOF or nil.
	return ch.Play(tx)
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
	ok = tx.PreRecord(ch.SDP())
	if !ok {
		return tx.Response(ErrInternal(res))
	}
	err := tx.Response(res)
	if err != nil {
		return err
	}
	// if is interleaved, fun recordServe serve
	// and blocked until err.
	// if is udp, fun recordServe just return.
	err = ch.Record(tx)
	if err != nil {
		if err == io.EOF {
			return err
		}
		log.Errorf("can not record: %v", err)
	}
	return nil
}

func (u *UnimplementedServerHandler) TEARDOWN(req Request, res Response, tx Transaction) error {
	log.Debugf("teardown request url: %s", req.URL().String())
	ch, ok := u.tc.GetCh(req.Channel())
	if !ok {
		return tx.Response(ErrInternal(res))
	}
	ch.Teardown(tx)
	return tx.Response(res)
}
