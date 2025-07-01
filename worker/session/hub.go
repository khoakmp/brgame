package session

import (
	"log"
	"sync"
)

type Hub struct {
	sesions map[string]Session
	lock    sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		sesions: make(map[string]Session),
		lock:    sync.RWMutex{},
	}
}

func (h *Hub) AddSession(s Session) {
	h.lock.Lock()
	defer h.lock.Unlock()
	log.Println("Add Session", s.ID())
	h.sesions[s.ID()] = s
}

func (h *Hub) GetSession(sessionID string) (Session, bool) {
	h.lock.RLock()
	defer h.lock.RUnlock()
	s, ok := h.sesions[sessionID]
	return s, ok
}
func (h *Hub) RemoveSession(sessionID string) {
	h.lock.Lock()
	defer h.lock.Unlock()
	log.Println("Remove Session", sessionID)
	delete(h.sesions, sessionID)
}
