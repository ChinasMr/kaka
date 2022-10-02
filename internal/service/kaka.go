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

func (s *KakaService) ANNOUNCE(req rtsp.Request, res rtsp.Response, tx rtsp.Transport) error {
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
	fmt.Println(message)

	return nil
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
