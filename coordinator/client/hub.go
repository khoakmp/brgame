package client

import (
	"fmt"
	"log"
	"sync"

	"github.com/khoakmp/brgame/api"
)

type Hub struct {
	clients map[string]*Client
	lock    sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[string]*Client),
	}
}

func (h *Hub) NumClient() int {
	h.lock.RLock()
	defer h.lock.RUnlock()
	return len(h.clients)
}

func (h *Hub) PrintClients() {
	h.lock.RLock()
	ids := make([]string, 0, len(h.clients))
	for id := range h.clients {
		ids = append(ids, id)
	}
	fmt.Println("ClientIDs:", ids)
}
func (h *Hub) AddClient(c *Client) {
	h.lock.Lock()
	defer h.lock.Unlock()
	log.Println("Add Client", c.id)
	h.clients[c.id] = c
}

func (h *Hub) RemoveClient(clientID string) {
	h.lock.Lock()
	defer h.lock.Unlock()
	log.Println("Remove Client", clientID)
	delete(h.clients, clientID)
}

func (h *Hub) GetClient(id string) (*Client, bool) {
	h.lock.RLock()
	defer h.lock.RUnlock()
	client, ok := h.clients[id]
	return client, ok
}

func (h *Hub) ForwardMessage(msg *api.Message) {
	h.lock.RLock()
	clients := make([]*Client, 0, len(msg.ReceiverIDs))
	for _, clientID := range msg.ReceiverIDs {
		if client, ok := h.clients[clientID]; ok {
			clients = append(clients, client)
		}
	}
	h.lock.RUnlock()
	for _, client := range clients {
		client.SendMessage(msg)
	}
}
