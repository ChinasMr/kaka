package biz

import (
	"context"
	"fmt"
	"github.com/ChinasMr/kaka/pkg/log"
	"github.com/ChinasMr/kaka/pkg/transport/rtsp"
	"github.com/google/uuid"
	"gortc.io/sdp"
	"sync"
)

type Room struct {
	Id        string
	Source    rtsp.Transport
	Terminals []rtsp.Transport
	SDP       *sdp.Message
	SDPRaw    []byte
	mu        sync.Mutex
}

type RoomRepo interface {
	Create(ctx context.Context, room *Room) (*Room, error)
	Get(ctx context.Context, id string) (*Room, error)
	Delete(ctx context.Context, id string) error
	SetRoomInput(ctx context.Context, id string, room *Room) error
}

type KakaUseCase struct {
	log  *log.Helper
	room RoomRepo
}

func NewKakaUseCase(logger log.Logger, repo RoomRepo) *KakaUseCase {
	return &KakaUseCase{
		log:  log.NewHelper(logger),
		room: repo,
	}
}

func (uc *KakaUseCase) CreateRoom(ctx context.Context, room *Room) (*Room, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}
	nr := &Room{
		Id: id.String(),
	}
	return uc.room.Create(ctx, nr)
}

func (uc *KakaUseCase) GetRoom(ctx context.Context, id string) (*Room, error) {
	return uc.room.Get(ctx, id)
}

func (uc *KakaUseCase) SetRoomInput(ctx context.Context, id string, room *Room) error {
	p, err := uc.room.Get(ctx, id)
	if err != nil {
		return err
	}
	if p.Source != nil {
		return fmt.Errorf("this room already has a source")
	}
	return uc.room.SetRoomInput(ctx, id, room)
}
