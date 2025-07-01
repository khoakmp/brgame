package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/khoakmp/brgame/api"
	"github.com/khoakmp/brgame/worker/config"
	"github.com/khoakmp/brgame/worker/coordinator"
	"github.com/khoakmp/brgame/worker/session"
	"github.com/khoakmp/brgame/worker/utils"
)

func main() {

	config.AppConfig = config.Configuration{
		VideoBufferSize:  400,
		AudioBufferSize:  400,
		InputChannelSize: 100,
		VideoClockRate:   90,
	}
	var pid int = os.Getpid()
	fmt.Println("Process ID:", pid)

	var exitFlag int32
	var coordConn coordinator.Conn = nil
	var connLock sync.RWMutex

	go func() {
		r := bufio.NewReader(os.Stdin)
		for {
			cmd, _, err := r.ReadLine()
			if err != nil {
				return
			}

			switch {
			case bytes.Equal(cmd, []byte("p")):

			case bytes.Equal(cmd, []byte("s")):
				var memstats runtime.MemStats
				runtime.ReadMemStats(&memstats)
				fmt.Println("Heap Alloc:", memstats.HeapAlloc, "numGC:", memstats.NumGC, "heapObjects:", memstats.HeapObjects)
			case bytes.Equal(cmd, []byte("q")):
				atomic.StoreInt32(&exitFlag, 1)
				connLock.RLock()
				if coordConn != nil {
					coordConn.Close()
				}
				connLock.RUnlock()
				return
			}
		}

	}()
	workerID := utils.RandString(6)
	fmt.Println(workerID)
	sessionHub := session.NewHub()

	for atomic.LoadInt32(&exitFlag) == 0 {
		conn, err := coordinator.Connect("ws", "ws://localhost:8080/ws", workerID)
		if err != nil {
			log.Printf("[W %s] Failed to connect Coordinator, %s\n", workerID, err)
			time.Sleep(time.Second * 2)
			continue
		}
		connLock.Lock()
		coordConn = conn
		connLock.Unlock()
		if atomic.LoadInt32(&exitFlag) == 1 {
			return
		}
		readloop(workerID, coordConn, sessionHub)
	}
}

type StartSessionMultiPeer struct {
	ClientIDs []string `json:"client_ids"`
	AppName   string   `json:"app_name"`
}

func readloop(workerID string, coordConn coordinator.Conn, sessionHub *session.Hub) {
	for {
		msg, err := coordConn.ReadMessage()
		if err != nil {
			log.Printf("[W %s] Failed to read message,%s\n", workerID, err)
			return
		}
		if len(msg.SessionID) > 0 {
			s, ok := sessionHub.GetSession(msg.SessionID)
			if ok {
				s.Receive(msg)
				continue
			}
			switch msg.Type {
			case api.MessageStartSession:

				var payload api.StartSessionPayload
				json.Unmarshal([]byte(msg.Payload), &payload)
				log.Println("Recv StartSession Message, sessionID:", msg.SessionID)
				payload.Print()

				s, err := session.NewSessionMultiPeer(workerID, payload.AppName, msg.SessionID, payload.ClientIDs, sessionHub, coordConn)
				if err != nil {
					log.Printf("[W %s] Failed to create Session, %s\n", workerID, err)
					// TODO: send failed message back to coordinator un?
					continue
				}
				sessionHub.AddSession(s)
			}
		}
	}
}
