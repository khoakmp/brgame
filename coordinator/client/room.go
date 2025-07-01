package client

import (
	"fmt"
	"sync"
	"time"
)

type WaitingSlot struct {
	expTime time.Time
	client  *Client
}

type RoomWaiting struct {
	slots *RingWait
	lock  sync.RWMutex
}

func (r *RoomWaiting) NumClientWaiting() int {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return r.slots.Len()
}

func (r *RoomWaiting) PrintClients(name string) {
	r.lock.RLock()
	txt := r.slots.GenInfor(name)
	r.lock.RUnlock()
	fmt.Print(txt)
}

func (r *RoomWaiting) releaseLoop() {
	ticker := time.NewTicker(time.Second * 2)
	for {
		<-ticker.C
		r.lock.Lock()
		current := time.Now()
		for i := 0; i < r.slots.Len(); i++ {
			slot := r.slots.Peek()
			if slot.expTime.After(current) {
				break
			}
			slot.client.WaitTimeout()
			r.slots.Pop()
		}
		r.lock.Unlock()
		//r.PrintClients("After Release")
	}
}

func (r *RoomWaiting) Process(client *Client, n int) []*Client {
	r.lock.Lock()
	defer r.lock.Unlock()
	if r.slots.Len() < n-1 {
		r.slots.Push(client, time.Now().Add(time.Second*10))
		return nil
	}
	var slots []WaitingSlot = make([]WaitingSlot, 0, r.slots.Len()+1)

	for !r.slots.IsEmpty() {
		slot := r.slots.Peek()
		r.slots.Pop()

		if slot.client.Alive() {
			slots = append(slots, slot)
		}
	}

	if len(slots) < n-1 {
		for _, slot := range slots {
			r.slots.Push(slot.client, slot.expTime)
		}
		r.slots.Push(client, time.Now().Add(time.Second*10))
		return nil
	}

	clients := make([]*Client, n)
	for i, slot := range slots {
		clients[i] = slot.client
	}

	clients[n-1] = client
	return clients
}

func NewRoomWaiting(size int) *RoomWaiting {
	room := &RoomWaiting{
		slots: NewRingWait(size),
		lock:  sync.RWMutex{},
	}
	go room.releaseLoop()
	return room
}

type HubRoomWaiting struct {
	rooms map[string]*RoomWaiting
	lock  sync.RWMutex
}

type RoomParam struct {
	AppName string
	NumSlot int
}

func NewHubRoomWaiting(params []RoomParam) *HubRoomWaiting {
	hub := &HubRoomWaiting{
		rooms: make(map[string]*RoomWaiting),
	}
	for _, p := range params {
		hub.rooms[p.AppName] = NewRoomWaiting(p.NumSlot)
	}
	return hub
}
func (h *HubRoomWaiting) GetRoom(appName string) (r *RoomWaiting, ok bool) {
	h.lock.RLock()
	defer h.lock.RUnlock()
	r, ok = h.rooms[appName]
	return
}

func (h *HubRoomWaiting) AddRoom(appName string, r *RoomWaiting) {
	h.lock.Lock()
	defer h.lock.Unlock()
	h.rooms[appName] = r
}
