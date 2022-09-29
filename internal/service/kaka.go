package service

import (
	"context"
	pb "github.com/ChinasMr/kaka/api/kaka/v1"
	"github.com/ChinasMr/kaka/pkg/log"
	"github.com/ChinasMr/kaka/pkg/transport/rtsp"
)

type KakaService struct {
	pb.UnimplementedKakaServer
	rtsp.UnimplementedServerHandler
	log *log.Helper
}

func (s *KakaService) DESCRIBE(req *rtsp.Request, res *rtsp.Response) error {
	log.Debugf("describe request input data: %+v", req)
	res.SetContent(nil)
	return nil
}

func NewKakaService(logger log.Logger) *KakaService {
	return &KakaService{
		log: log.NewHelper(logger),
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
