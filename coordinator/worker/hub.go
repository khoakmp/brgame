package worker

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
)

type Hub struct {
	workers map[string]*Worker
	lock    sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		workers: make(map[string]*Worker),
	}
}

func (h *Hub) NumActiveWorker() int {
	h.lock.RLock()
	defer h.lock.RUnlock()
	return len(h.workers)
}

func (h *Hub) PrintWorkers() {
	h.lock.RLock()
	wIDs := make([]string, len(h.workers))
	for wID := range h.workers {
		wIDs = append(wIDs, wID)
	}
	h.lock.RUnlock()
	fmt.Println(wIDs)
}
func (h *Hub) AddWorker(w *Worker) {
	h.lock.Lock()
	defer h.lock.Unlock()
	log.Println("Add worker", w.id)
	h.workers[w.id] = w
}

func (h *Hub) GetWorker(id string) (*Worker, bool) {
	h.lock.RLock()
	defer h.lock.RUnlock()
	w, ok := h.workers[id]
	return w, ok
}

func (h *Hub) RemoveWorker(id string) {
	h.lock.Lock()
	defer h.lock.Unlock()
	log.Println("Remove worker", id)
	delete(h.workers, id)
}

func (h *Hub) RandomWorker() *Worker {
	h.lock.RLock()
	defer h.lock.RUnlock()
	if len(h.workers) == 0 {
		return nil
	}
	keys := make([]string, 0, len(h.workers))
	for k := range h.workers {
		keys = append(keys, k)
	}
	idx := rand.Int() % len(keys)
	return h.workers[keys[idx]]
}
