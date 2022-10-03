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
	log.Debugf("options request from %s", tx.Addr().String())
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
	log.Debugf("announce request from %s", tx.Addr().String())
	if tx.Status() != status.STARING {
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
	log.Debugf("setup request from: %+v", tx.Addr().String())
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

	return nil
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
