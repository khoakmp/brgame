package worker

import (
	"errors"
	"log"
	"net"
	"sync/atomic"
	"time"

	"github.com/khoakmp/brgame/api"
	"github.com/khoakmp/brgame/coordinator/network"
)

type MessageHandler interface {
	HandleWorkerMessage(msg *api.Message, worker *Worker)
}

type Worker struct {
	id             string
	conn           network.Conn
	hub            *Hub
	outMessageChan chan *api.Message
	handler        MessageHandler
	exitChan       chan struct{}
	exitFlag       uint32
	lastContact    int64
}

func (w *Worker) setLastContact(val int64) {
	atomic.StoreInt64(&w.lastContact, val)
}

func (w *Worker) ID() string {
	return w.id
}

func (w *Worker) readLoop() {
	for {
		msg, err := w.conn.ReadMessage()
		w.setLastContact(time.Now().UnixMicro())

		if err != nil {
			log.Printf("[Worker %s] Failed to read message %s\n", w.id, err)
			break
		}

		w.handler.HandleWorkerMessage(&msg, w)
		// after process -> change coordinator app state, it may not need to send response message
		// back to worker
	}
	w.Close()
}

func (w *Worker) Close() {
	if !atomic.CompareAndSwapUint32(&w.exitFlag, 0, 1) {
		return
	}
	log.Printf("[Worker %s] Closing\n", w.id)
	w.conn.Close()
	close(w.exitChan)
	w.hub.RemoveWorker(w.id)
}

func (w *Worker) Exiting() bool {
	return atomic.LoadUint32(&w.exitFlag) == 1
}

func (w *Worker) writeLoop() {
	//log.Printf("[Worker %s] Start WriteLoop\n", w.id)

	for {
		select {
		case <-w.exitChan:
			log.Printf("[Worker %s] Exit WriteLoop\n", w.id)
			return
		case msg := <-w.outMessageChan:
			if err := w.conn.WriteMessage(msg); err != nil {
				log.Printf("[Worker %s] Failed to send message\n", w.id)
				if errors.Is(err, net.ErrClosed) {
					return
				}
			}
		}
	}
}

func (w *Worker) SendMessage(msg *api.Message) {
	w.outMessageChan <- msg
}

func New(id string, conn network.Conn, hub *Hub, handler MessageHandler) *Worker {
	w := &Worker{
		id:             id,
		conn:           conn,
		hub:            hub,
		outMessageChan: make(chan *api.Message),
		handler:        handler,
		exitChan:       make(chan struct{}),
		exitFlag:       0,
	}
	w.setLastContact(time.Now().UnixMicro())
	go w.readLoop()
	go w.writeLoop()

	return w
}
