package service

import (
	"context"
	"fmt"
	pb "github.com/ChinasMr/kaka/api/kaka/v1"
	"github.com/ChinasMr/kaka/internal/biz"
	"github.com/ChinasMr/kaka/pkg/log"
	"github.com/ChinasMr/kaka/pkg/transport/rtsp"
	"github.com/ChinasMr/kaka/pkg/transport/rtsp/header"
	"github.com/ChinasMr/kaka/pkg/transport/rtsp/method"
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

//func (s *KakaService) DESCRIBE(req *rtsp.Request, res *rtsp.Response) error {
//	log.Debugf("describe request input data: %+v", req)
//	res.SetContent(nil)
//	return nil
//}

func NewKakaService(logger log.Logger, useCase *biz.KakaUseCase) *KakaService {
	return &KakaService{
		log: log.NewHelper(logger),
		uc:  useCase,
	}
}

func (s *KakaService) Debug(ctx context.Context, req *pb.DebugRequest) (*pb.DebugReply, error) {
	s.log.Debugf("debug request incoming!")
	return &pb.DebugReply{
		Id:       "1",
		Name:     "2",
		Version:  "3",
		Metadata: map[string]string{},
	}, nil
}

func (s *KakaService) OPTIONS(_ rtsp.Request, res rtsp.Response, tx rtsp.Transport) error {
	s.log.Debugf("options request from %s", tx.Addr().String())
	methods := strings.Join([]string{
		method.DESCRIBE.String(),
		method.ANNOUNCE.String(),
		method.SETUP.String(),
		method.PLAY.String(),
		method.PAUSE.String(),
		method.RECORD.String(),
		method.TEARDOWN.String(),
	}, ", ")
	res.SetHeader(header.Public, methods)
	return nil
}

func (s *KakaService) ANNOUNCE(req rtsp.Request, _ rtsp.Response, tx rtsp.Transport) error {
	s.log.Debugf("announce request from %s", tx.Addr().String())
	if tx.Status() != status.STARTING {
		return fmt.Errorf("trans status error")
	}
	ct, ok := req.Header(header.ContentType)
	if !ok || len(ct) == 0 {
		return fmt.Errorf("can not get content type")
	}
	if ct[0] != "application/sdp" {
		return fmt.Errorf("content type error")
	}
	message, err := decodeSDP(req.Body())
	if err != nil {
		return fmt.Errorf("can not decode sdp message")
	}
	id := parseRoomId(req.Path())
	err = s.uc.SetRoomInput(context.Background(), id, &biz.Room{
		Source:    tx,
		Terminals: nil,
		SDP:       message,
		SDPRaw:    req.Body(),
	})
	if err != nil {
		return err
	}
	tx.SetStatus(status.ANNOUNCED)
	return nil
}

func (s *KakaService) SETUP(req rtsp.Request, res rtsp.Response, tx rtsp.Transport) error {
	s.log.Debugf("setup request from: %+v", tx.Addr().String())
	transports, ok := req.Transport()
	if !ok {
		return fmt.Errorf("err setup can not get transport")
	}
	ok = transports.Has("unicast")
	if !ok {
		return fmt.Errorf("err setup can not get unicast")
	}

	// record
	if tx.Status() == status.ANNOUNCED || tx.Status() == status.PRERECORD {
		ok = transports.Has("mode=record")
		if !ok {
			return fmt.Errorf("error setup can not get mode=record")
		}
		isUDP := transports.Has("RTP/AVP/UDP")
		isTCP := transports.Has("RTP/AVP/TCP")
		if isUDP == false && isTCP == false {
			return fmt.Errorf("err setup can not get RTP/AVP/UDP or RTP/AVP/TCP")
		}
		// tcp
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
			res.SetHeader("Session", "12345678")
			tx.SetStatus(status.PRERECORD)
			return nil
		}
		if isUDP {
			return nil
		}

	}

	// play
	if tx.Status() == status.STARTING || tx.Status() == status.PREPLAY {
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
			res.SetHeader("Session", "12345678")
			tx.SetStatus(status.PREPLAY)
			return nil
		}
	}

	return fmt.Errorf("status error")
}

func (s *KakaService) RECORD(req rtsp.Request, res rtsp.Response, tx rtsp.Transport) error {
	s.log.Debugf("record request from: %s", tx.Addr().String())
	id := parseRoomId(req.Path())
	room, err := s.uc.GetRoom(context.Background(), id)
	if err != nil {
		return err
	}
	if tx.Status() != status.PRERECORD {
		return fmt.Errorf("status error")
	}
	res.SetHeader("Session", "12345678")
	err = tx.SendResponse(res)
	if err != nil {
		return err
	}
	tx.SetStatus(status.RECORD)
	s.log.Debugf("------- Room %s start recording from %s -------", id, tx.Addr().String())
	defer func() {
		room.Source = nil
		s.log.Debugf("------- Room %s ended recording from %s -------", id, tx.Addr().String())
	}()
	buf := make([]byte, 2048)

	for {
		channel, frameLen, err1 := tx.ReadInterleavedFrame(buf)
		if err1 != nil {

			if err1 == io.EOF {
				return nil
			}
			return err1
		}
		for _, terminal := range room.Terminals {
			if terminal.Status() == status.PLAY {
				_ = terminal.WriteInterleavedFrame(channel, buf[:frameLen])
				s.log.Debugf("push stream to %s --------> %d bytes", terminal.Addr().String(), frameLen)
			}
		}
	}

}

func (s *KakaService) DESCRIBE(req rtsp.Request, res rtsp.Response, tx rtsp.Transport) error {
	log.Debugf("describe request from: %s", tx.Addr().String())
	if tx.Status() != status.STARTING {
		return fmt.Errorf("status error")
	}
	id := parseRoomId(req.Path())
	room, err := s.uc.GetRoom(context.Background(), id)
	if err != nil {
		return err
	}
	if room.Source == nil {
		return fmt.Errorf("nil room")
	}

	res.SetHeader("Content-Base", req.URL().String())
	res.SetHeader("Content-Type", "application/sdp")
	res.SetBody(room.SDPRaw)

	return nil
}

func (s *KakaService) PLAY(req rtsp.Request, res rtsp.Response, tx rtsp.Transport) error {
	log.Debugf("play request from: %s", tx.Addr().String())
	if tx.Status() != status.PREPLAY {
		return fmt.Errorf("status error")
	}
	id := parseRoomId(req.Path())
	room, err := s.uc.GetRoom(context.Background(), id)
	if err != nil {
		return err
	}

	// todo check the channel setup process.
	res.SetHeader("Session", "12345678")
	err = tx.SendResponse(res)
	if err != nil {
		return err
	}
	tx.SetStatus(status.PLAY)
	room.Terminals = append(room.Terminals, tx)
	buf := make([]byte, 2048)
	for {
		_, err1 := tx.RawConn().Read(buf)
		if err1 != nil {
			if err == io.EOF {
				return nil
			}
			return err1
		}
	}
}

func parseRoomId(p string) string {
	return strings.TrimLeft(p, "/")
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
