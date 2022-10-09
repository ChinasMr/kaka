package biz

import (
	"context"
	"github.com/ChinasMr/kaka/pkg/log"
	"github.com/google/uuid"
	"gortc.io/sdp"
	"sync"
)

type Channel struct {
	Id        string
	Terminals TerminalsOperator
	SDP       *sdp.Message
	RawSDP    []byte
	mu        sync.Mutex
	Input     chan []byte
}

type ChannelRepo interface {
	Create(ctx context.Context, id string) (*Channel, error)
	Get(ctx context.Context, id string) (*Channel, error)
	Delete(ctx context.Context, id string) error
}

type KakaUseCase struct {
	log     *log.Helper
	channel ChannelRepo
}

func NewKakaUseCase(logger log.Logger, repo ChannelRepo) *KakaUseCase {
	return &KakaUseCase{
		log:     log.NewHelper(logger),
		channel: repo,
	}
}

func (uc *KakaUseCase) CreateChannel(ctx context.Context) (*Channel, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}
	return uc.channel.Create(ctx, id.String())
}

func (uc *KakaUseCase) GetChannel(ctx context.Context, id string) (*Channel, error) {
	return uc.channel.Get(ctx, id)
}

func (uc *KakaUseCase) SetChannelPresentationDescription(ctx context.Context, id string, sdp *sdp.Message, raw []byte) error {
	p, err := uc.channel.Get(ctx, id)
	if err != nil {
		return err
	}
	// todo check the source state.
	p.mu.Lock()
	p.SDP = sdp
	p.RawSDP = raw
	p.mu.Unlock()
	return nil
}
