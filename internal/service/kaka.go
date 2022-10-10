package service

import (
	"context"
	"fmt"
	pb "github.com/ChinasMr/kaka/api/kaka/v1"
	"github.com/ChinasMr/kaka/internal/biz"
	"github.com/ChinasMr/kaka/pkg/log"
	"github.com/ChinasMr/kaka/pkg/transport/rtsp"
	"github.com/ChinasMr/kaka/pkg/transport/rtsp/header"
	"github.com/ChinasMr/kaka/pkg/transport/rtsp/status"
	"gortc.io/sdp"
	"io"
	"strings"
)

type KakaService struct {
	pb.UnimplementedKakaServer
	rtsp.UnimplementedServerHandler
	uc  *biz.KakaUseCase
	log *log.Helper
}

func NewKakaService(logger log.Logger, useCase *biz.KakaUseCase) *KakaService {
	return &KakaService{
		log: log.NewHelper(logger),
		uc:  useCase,
	}
}

func (s *KakaService) Debug(_ context.Context, _ *pb.DebugRequest) (*pb.DebugReply, error) {
	s.log.Debugf("debug request incoming!")
	return &pb.DebugReply{
		Id:       "1",
		Name:     "2",
		Version:  "3",
		Metadata: map[string]string{},
	}, nil
}

func (s *KakaService) ANNOUNCE(req rtsp.Request, res rtsp.Response) {
	s.log.Debugf("announce request from %s", req.URL().String())
	if req.ContentType() != header.ContentTypeSDP {
		s.log.Errorf("unsupported presentation description format")
		rtsp.Err500(res)
		return
	}
	message, err := decodeSDP(req.Body())
	if err != nil {
		s.log.Errorf("can not decode sdp: %v", err)
		rtsp.Err500(res)
		return
	}
	id := parseChannelId(req.Path())

	err = s.uc.SetChannelPresentationDescription(context.Background(), id, message, req.Body())
	if err != nil {
		s.log.Errorf("can not set channel presentation description: %v", err)
		rtsp.Err500(res)
		return
	}
}

func (s *KakaService) SETUP(req rtsp.Request, res rtsp.Response, tx rtsp.Transaction) error {
	s.log.Debugf("setup request: %+v", req.URL().String())
	transports, _ := req.Transport()
	record := transports.Has("mode=record")
	channelId := parseChannelId(req.Path())
	channel, err := s.uc.GetChannel(context.Background(), channelId)
	if err != nil {
		s.log.Errorf("can not get channel: %v", err)
		rtsp.Err404(res)
		return tx.Transport().SendResponse(res)
	}
	// record
	if record {
		isUDP := transports.Has("RTP/AVP/UDP")
		isTCP := transports.Has("RTP/AVP/TCP")
		if isUDP == false && isTCP == false {
			s.log.Errorf("err setup can not get RTP/AVP/UDP or RTP/AVP/TCP")
			rtsp.ErrUnsupportedTransport(res)
			return nil
		}
		defer func() {
			if tx.Medias() >= len(channel.SDP.Medias) {
				tx.SetStatus(status.READY)
			}
		}()
		// tcp
		if isTCP {
			interleaved := transports.Value("interleaved")
			if interleaved == "" {
				s.log.Errorf("can not get interleaved")
				rtsp.ErrUnsupportedTransport(res)
				return nil
			}
			// todo check the stream channel
			res.SetHeader("Transport", strings.Join([]string{
				"RTP/AVP/TCP",
				"unicast",
				fmt.Sprintf("interleaved=%s", interleaved),
			}, ";"))
			res.SetHeader("Session", tx.ID())
			err1 := tx.Transport().SendResponse(res)
			if err1 != nil {
				return err1
			}
			// acknowledged response.
			tx.AddMedia(interleaved)
			tx.SetInterleaved()
			s.log.Debugf("%s set interleaved stream: %s", req.URL().String(), interleaved)
			return nil
		}
		if isUDP {
			return nil
		}

	}

	// play
	if !record {
		defer func() {
			if tx.Medias() >= len(channel.SDP.Medias) {
				tx.SetStatus(status.READY)
				// if ready, add this ter to clients set.
				channel.Terminals.Add(tx)
			}
		}()
		isTCP := transports.Has("RTP/AVP/TCP")
		if isTCP == false {
			return fmt.Errorf("can not get transport info")
		}
		if isTCP {
			interleaved := transports.Value("interleaved")
			if interleaved == "" {
				return fmt.Errorf("can not get interleaved")
			}
			// todo check the stream channel
			res.SetHeader("Transport", strings.Join([]string{
				"RTP/AVP/TCP",
				"unicast",
				fmt.Sprintf("interleaved=%s", interleaved),
			}, ";"))
			res.SetHeader("Session", tx.ID())
			err1 := tx.Transport().SendResponse(res)
			if err1 != nil {
				return err1
			}
			// acknowledged response.
			tx.AddMedia(interleaved)
			tx.SetInterleaved()
			s.log.Debugf("%s set interleaved stream: %s", req.URL().String(), interleaved)
			return nil
		}
	}

	return nil
}

func (s *KakaService) RECORD(req rtsp.Request, res rtsp.Response, tx rtsp.Transaction) error {
	s.log.Debugf("record request: %s", req.URL().String())
	channelId := parseChannelId(req.Path())
	channel, err := s.uc.GetChannel(context.Background(), channelId)
	if err != nil {
		rtsp.Err404(res)
		return tx.Transport().SendResponse(res)
	}
	res.SetHeader("Session", tx.ID())
	err = tx.Transport().SendResponse(res)
	if err != nil {
		return err
	}
	tx.SetStatus(status.RECORDING)
	// interleaved, the tpc connection for RTSP turns to RTP/RTCP.
	if tx.Interleaved() {
		tpcTrans, ok := tx.Transport().(*rtsp.TcpTransport)
		if !ok {
			return fmt.Errorf("interleaved tranports error")
		}
		s.log.Debugf("interleaved connect to %s", tpcTrans.Addr().String())

		input := channel.Terminals.Input()
		for {
			buf := make([]byte, 2048)
			ch, n, err1 := tpcTrans.ReadInterleavedFrame(buf)
			if err1 != nil {
				if err1 == io.EOF {
					s.log.Debugf("client close the interleaved connection")
					return nil
				}
				return err
			}
			input <- &rtsp.Package{
				Ch:   ch,
				Data: buf[:n],
			}
		}
	}
	return nil
}

func (s *KakaService) TEARDOWN(req rtsp.Request, res rtsp.Response, tx rtsp.Transaction) error {
	s.log.Debugf("teardown/down request: %s", req.URL().String())
	res.SetHeader("Session", tx.ID())
	_ = tx.Transport().SendResponse(res)
	// todo clear serve resources.
	return nil
}

func (s *KakaService) DESCRIBE(req rtsp.Request, res rtsp.Response) {
	log.Debugf("describe request from: %s", req.URL().String())
	id := parseChannelId(req.Path())
	channel, err := s.uc.GetChannel(context.Background(), id)
	if err != nil {
		s.log.Errorf("can not get channel: %v", err)
		return
	}
	res.SetHeader("Content-Base", req.URL().String())
	res.SetHeader("Content-Type", header.ContentTypeSDP)
	res.SetBody(channel.RawSDP)
}

func (s *KakaService) PLAY(req rtsp.Request, res rtsp.Response, tx rtsp.Transaction) error {
	log.Debugf("play request from: %s", req.URL().String())
	id := parseChannelId(req.Path())
	_, err := s.uc.GetChannel(context.Background(), id)
	if err != nil {
		rtsp.Err404(res)
		return tx.Transport().SendResponse(res)
	}
	res.SetHeader("Session", tx.ID())
	err = tx.Transport().SendResponse(res)
	if err != nil {
		return err
	}
	tx.SetStatus(status.PLAYING)

	if tx.Interleaved() {
		buf := make([]byte, 2048)
		for {
			_, err1 := tx.Transport().ReadData(buf)
			if err1 != nil {
				tx.SetStatus(status.READY)
				return io.EOF
			}
		}
	}

	return nil
}

func parseChannelId(p string) string {
	str := strings.TrimLeft(p, "/")
	index := strings.Index(str, "/")
	if index != -1 {
		return str[:index]
	}
	return str
}

func decodeSDP(content []byte) (*sdp.Message, error) {
	s, err := sdp.DecodeSession(content, nil)
	if err != nil {
		return nil, err
	}
	rv := &sdp.Message{}
	d := sdp.NewDecoder(s)
	err = d.Decode(rv)
	if err != nil {
		return nil, err
	}
	return rv, nil
}
