package biz

import (
	"context"
	"github.com/ChinasMr/kaka/pkg/log"
	"github.com/google/uuid"
)

type Room struct {
	Id string
}

type RoomRepo interface {
	Create(ctx context.Context, room *Room) (*Room, error)
	Get(ctx context.Context, id string) (*Room, error)
	Delete(ctx context.Context, id string) error
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
