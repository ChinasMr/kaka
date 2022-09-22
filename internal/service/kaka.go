package service

import (
	"context"
	pb "github.com/ChinasMr/kaka/api/kaka/v1"
	"github.com/ChinasMr/kaka/pkg/log"
)

type KakaService struct {
	pb.UnimplementedKakaServer
	log *log.Helper
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
