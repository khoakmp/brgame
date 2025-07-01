package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"github.com/khoakmp/brgame/coordinator"
	"github.com/khoakmp/brgame/utils"
)

func runServer() {
	mux := http.NewServeMux()
	var upgrader websocket.Upgrader

	readLoop := func(conn *websocket.Conn, id string) {
		fmt.Println("Start Readloop for wsconn", id)
		for {
			msgType, buf, err := conn.ReadMessage()
			if err != nil {
				fmt.Println("failed to read msg:", err)
				return
			}
			fmt.Println("Recv msg type:", msgType)
			fmt.Println("msg:", string(buf))

		}
	}
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		id := r.Header["Id"][0]

		conn, err := upgrader.Upgrade(w, r, http.Header{})
		if err != nil {
			fmt.Println("Failed to create ws conn,", err)
			return
		}
		go readLoop(conn, id)

	})

	http.ListenAndServe(":8081", mux)
}
func runClient() {
	reqHeader := http.Header{}
	reqHeader.Add("Id", utils.RandString(6))
	conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:8081/ws", reqHeader)
	if err != nil {
		fmt.Println(err)
		return
	}
	conn.Close()
	fmt.Println("Connected WS ,local addr:", conn.LocalAddr().String())
	time.Sleep(time.Second)
}
func RunSeparate() {
	cmd := os.Args[1]
	switch cmd {
	case "s":
		runServer()
	case "c":
		runClient()
	}
}

func RunOne() {
	go func() {
		time.Sleep(time.Millisecond * 20)
		runClient()
	}()
	runServer()
}
func main() {
	coord := coordinator.New()
	coord.Run(":8080")
	//RunSeparate()
}
