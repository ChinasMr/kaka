package data

import (
	"context"
	"fmt"
	"github.com/ChinasMr/kaka/internal/biz"
	"github.com/ChinasMr/kaka/pkg/log"
	"sync"
)

var defaultRoom = "live"

type roomRepo struct {
	log   *log.Helper
	rooms map[string]*biz.Room
	rwm   sync.RWMutex
}

func (r *roomRepo) SetRoomInput(_ context.Context, id string, room *biz.Room) error {
	r.rwm.Lock()
	defer r.rwm.Unlock()
	ro, ok := r.rooms[id]
	if !ok {
		return fmt.Errorf("room not found")
	}
	ro.Source = room.Source
	ro.SDP = room.SDP
	ro.SDPRaw = room.SDPRaw
	return nil
}

func (r *roomRepo) Get(_ context.Context, id string) (*biz.Room, error) {
	r.rwm.RLock()
	defer r.rwm.RUnlock()
	rv, ok := r.rooms[id]
	if !ok {
		return nil, fmt.Errorf("room not found")
	}
	return rv, nil
}

func (r *roomRepo) Delete(_ context.Context, id string) error {
	r.rwm.Lock()
	defer r.rwm.Unlock()
	delete(r.rooms, id)
	return nil
}

func (r *roomRepo) Create(_ context.Context, room *biz.Room) (*biz.Room, error) {
	r.rwm.Lock()
	defer r.rwm.Unlock()
	r.rooms[room.Id] = room
	return room, nil
}

func NewRoomRepo(logger log.Logger) biz.RoomRepo {
	return &roomRepo{
		log: log.NewHelper(logger),
		rooms: map[string]*biz.Room{
			defaultRoom: {Id: defaultRoom},
		},
		rwm: sync.RWMutex{},
	}
}
