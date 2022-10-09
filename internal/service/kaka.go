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

func (s *KakaService) Debug(ctx context.Context, _ *pb.DebugRequest) (*pb.DebugReply, error) {
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
		return err
	}
	s.log.Debugf("channel sdp: %+v", channel)
	s.log.Debugf("path: %v, channel: %v", req.Path(), channelId)
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
			tx.AddMedia(interleaved)
			return nil
		}
		if isUDP {
			return nil
		}

	}

	//// play
	//if !record {
	//	isTCP := transports.Has("RTP/AVP/TCP")
	//	if isTCP == false {
	//		return fmt.Errorf("can not get transport info")
	//	}
	//	if isTCP {
	//		interleaved := transports.Value("interleaved")
	//		if interleaved == "" {
	//			return fmt.Errorf("can not get interleaved")
	//		}
	//		// todo check the stream channel
	//		res.SetHeader("Transport", strings.Join([]string{
	//			"RTP/AVP/TCP",
	//			"unicast",
	//			fmt.Sprintf("interleaved=%s", interleaved),
	//		}, ";"))
	//		res.SetHeader("Session", "12345678")
	//		return nil
	//	}
	//}

	return nil
}

//func (s *KakaService) RECORD(req rtsp.Request, res rtsp.Response, tx rtsp.Transport) error {
//	s.log.Debugf("record request from: %s", tx.Addr().String())
//	id := parseRoomId(req.Path())
//	room, err := s.uc.GetRoom(context.Background(), id)
//	if err != nil {
//		return err
//	}
//
//	res.SetHeader("Session", "12345678")
//	err = tx.SendResponse(res)
//	if err != nil {
//		return err
//	}
//	s.log.Debugf("------- Room %s start recording from %s -------", id, tx.Addr().String())
//	defer func() {
//		room.Source = nil
//		s.log.Debugf("------- Room %s ended recording from %s -------", id, tx.Addr().String())
//	}()
//	buf := make([]byte, 2048)
//
//	for {
//		channel, frameLen, err1 := tx.ReadInterleavedFrame(buf)
//		if err1 != nil {
//
//			if err1 == io.EOF {
//				return nil
//			}
//			return err1
//		}
//		for _, terminal := range room.Terminals {
//			if true {
//				_ = terminal.WriteInterleavedFrame(channel, buf[:frameLen])
//				s.log.Debugf("push stream to %s --------> %d bytes", terminal.Addr().String(), frameLen)
//			}
//		}
//	}
//
//}
//
//func (s *KakaService) DESCRIBE(req rtsp.Request, res rtsp.Response) {
//	log.Debugf("describe request from: %s", req.URL().String())
//	id := parseRoomId(req.Path())
//	room, err := s.uc.GetRoom(context.Background(), id)
//	if err != nil {
//		return
//	}
//	if room.Source == nil {
//		return
//	}
//
//	res.SetHeader("Content-Base", req.URL().String())
//	res.SetHeader("Content-Type", "application/sdp")
//	res.SetBody(room.SDPRaw)
//
//}
//
//func (s *KakaService) PLAY(req rtsp.Request, res rtsp.Response, tx rtsp.Transport) error {
//	log.Debugf("play request from: %s", tx.Addr().String())
//	id := parseRoomId(req.Path())
//	room, err := s.uc.GetRoom(context.Background(), id)
//	if err != nil {
//		return err
//	}
//
//	// todo check the channel setup process.
//	res.SetHeader("Session", "12345678")
//	err = tx.SendResponse(res)
//	if err != nil {
//		return err
//	}
//	room.Terminals = append(room.Terminals, tx)
//	//buf := make([]byte, 2048)
//	//for {
//	//	_, err1 := tx.RawConn().Read(buf)
//	//	if err1 != nil {
//	//		if err == io.EOF {
//	//			return nil
//	//		}
//	//		return err1
//	//	}
//	//}
//	return nil
//}

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
