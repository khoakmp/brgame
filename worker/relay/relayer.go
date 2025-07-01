package relay

import (
	"container/ring"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/khoakmp/brgame/worker/constants"
	"github.com/khoakmp/brgame/worker/event"
	"github.com/khoakmp/brgame/worker/webrtc"
	"github.com/pion/rtp"
)

type Relayer struct {
	traceID            string
	translateKeyCodeFn func(peerOrder int, keycode int) int
	wineConn           *net.TCPConn
	wineLock           sync.Mutex
}

func NewRelayer(traceID string,
	videoChannel, audioChannel chan<- *rtp.Packet, inputChannel chan webrtc.Packet,
	videoListener, audioListener *net.UDPConn, inputListener *net.TCPListener,
	translateKeyCodeFn func(peerOrder int, keycode int) int,
) *Relayer {
	relayer := Relayer{
		traceID:            traceID,
		translateKeyCodeFn: translateKeyCodeFn,
	}

	startChan := make(chan struct{})
	var startOnce sync.Once

	go func() {
		log.Println("Wait for SyncInput connectTCP...")

		for {

			conn, err := inputListener.AcceptTCP()
			if err != nil {
				log.Printf("[%s] Failed to accept wine conn, %s\n", relayer.traceID, err)
				startOnce.Do(func() {
					close(startChan)
				})

				return
			}

			log.Println("One TCPconn established for sync input")

			conn.SetKeepAlive(true)
			conn.SetKeepAlivePeriod(time.Millisecond * 100)

			relayer.wineLock.Lock()
			if relayer.wineConn != nil {
				relayer.wineConn.Close()
			}
			relayer.wineConn = conn
			relayer.wineLock.Unlock()

			startOnce.Do(func() {
				close(startChan)
			})
		}
	}()

	go relayer.startRelay(startChan, videoListener, videoChannel, "video")
	go relayer.startRelay(startChan, audioListener, audioChannel, "audio")
	go relayer.handleInput(startChan, inputChannel)

	return &relayer
}

func (r *Relayer) startRelay(startChan chan struct{}, listener *net.UDPConn, channel chan<- *rtp.Packet, streamType string) {
	<-startChan
	log.Printf("[Session %s] Start Relaying %s from udp socket to buffer channel\n", r.traceID, streamType)

	buffers := ring.New(120)
	n := buffers.Len()
	for i := 0; i < n; i++ {
		buffers.Value = make([]byte, 1500)
		buffers = buffers.Next()
	}

	for {
		buf := buffers.Value.([]byte)
		buffers = buffers.Next()
		n, _, err := listener.ReadFrom(buf)
		if err != nil {
			log.Printf("[%s] Faield to read %s packet from udp listener -> stop relaying, %s\n", r.traceID, streamType, err)
			return
		}
		var packet rtp.Packet

		if err := packet.Unmarshal(buf[:n]); err != nil {
			continue
		}
		channel <- &packet
	}
}

func (r *Relayer) Close() {
	r.wineLock.Lock()
	defer r.wineLock.Unlock()
	log.Printf("[Session %s] Close Relayer\n", r.traceID)
	if r.wineConn != nil {
		r.wineConn.Close()
	}
}
func (r *Relayer) handleInput(startChan chan struct{}, inputChannel <-chan webrtc.Packet) {
	log.Printf("[Session %s] Relayer wait for start chan close \n", r.traceID)

	<-startChan
	log.Printf("[Session %s] Start relaying input to VM\n", r.traceID)
	for pkt := range inputChannel {
		switch pkt.Packet.Type {
		case constants.KEY_UP:
			r.handleKeyEvent(pkt.PeerOrder, pkt.Packet.Data, 0)
		case constants.KEY_DOWN:
			r.handleKeyEvent(pkt.PeerOrder, pkt.Packet.Data, 1)
		case constants.MOUSE_MOVE:
			r.handleMouseEvent(pkt.PeerOrder, pkt.Packet.Data, 0)
		case constants.MOUSE_DOWN:
			r.handleMouseEvent(pkt.PeerOrder, pkt.Packet.Data, 1)
		case constants.MOUSE_UP:
			r.handleMouseEvent(pkt.PeerOrder, pkt.Packet.Data, 2)
		}
	}
}

/* type keyEventPayload struct {
	Keycode byte `json:"keycode"`
} */

func (r *Relayer) handleKeyEvent(peerOrder int, data string, state int) {
	var event event.KeyEventPayload
	//log.Printf("[Session %s] Recv KeyEvent from peerOrder %d payload:%s\n", r.traceID, peerOrder, data)
	json.Unmarshal([]byte(data), &event)

	keycode := r.translateKeyCodeFn(peerOrder, event.Keycode)
	if keycode > 255 {
		return
	}

	if _, err := r.wineConn.Write([]byte(fmt.Sprintf("K%d,%d|", keycode, state))); err != nil {
		log.Printf("[%s] failed to send key event,%s\n", r.traceID, err)
	}

}

/*
type mouseEventPayload struct {
	IsLeft byte    `json:"isleft"`
	X      float32 `json:"x"`
	Y      float32 `json:"y"`
	Width  float32 `json:"width"`
	Height float32 `json:"height"`
} */

func (r *Relayer) handleMouseEvent(peerOrder int, data string, state byte) {
	var p event.MouseEventPayload
	//log.Printf("[Session %s] Recv MouseEvent from peerOrder %d payload:%s\n", r.traceID, peerOrder, data)

	if err := json.Unmarshal([]byte(data), &p); err != nil {
		log.Printf("[%s] Failed to parse mouse event payload: %s\n", r.traceID, err.Error())
		return
	}
	if peerOrder != 0 {
		return
	}
	p.X = p.X / p.Width
	p.Y = p.Y / p.Height

	cmd := fmt.Sprintf("M%d,%d,%f,%f,%f,%f|", p.IsLeft, state, p.X, p.Y, p.Width, p.Height)
	if _, err := r.wineConn.Write([]byte(cmd)); err != nil {
		log.Printf("[%s] failed to send key event,%s\n", r.traceID, err)
	}
}
