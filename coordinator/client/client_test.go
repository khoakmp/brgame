package client

import (
	"fmt"
	"log"
	"net/http"
	"testing"

	"github.com/khoakmp/brgame/api"
	"github.com/khoakmp/brgame/coordinator/network"
	"github.com/khoakmp/brgame/coordinator/ws"
)

func CreateWsHandler() (http.Handler, <-chan network.Conn) {
	connChan := make(chan network.Conn)

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := ws.ServeReq(w, r, http.Header{})
		if err != nil {
			log.Println("Failed to create ws conn", err)
			return
		}
		connChan <- conn
	})
	return mux, connChan
}

type MockMessageHandler struct {
}

func (mh *MockMessageHandler) HandleClientMessage(msg *api.Message, cli *Client) {
	fmt.Println("Handling msg: ", msg.Type)
}

func TestClient(t *testing.T) {
	/* handler, connChan := CreateWsHandler()
	clientHub := NewHub()
	msgHandler := &MockMessageHandler{}
	go func() {
		http.ListenAndServe(":8080", handler)
	}()
	reqHeader := http.Header{}
	reqHeader.Add("Client_id", utils.RandString(6))
	reqHeader.Add("Role", "client")

	connCli, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/ws", reqHeader)
	if err != nil {
		log.Println("Failed to dial websocket:", err)
		return
	}
	go func() {
		connSrv := <-connChan
		clientID := utils.RandString(6)

		client := New(clientID, connSrv, msgHandler, clientHub)
		clientHub.AddClient(client)

	}()

	reqGameMsg := api.Message{
		SessionID: "",
	} */
	//connCli.WriteJSON()
}
