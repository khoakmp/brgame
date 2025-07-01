package coordinator

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/khoakmp/brgame/api"
	"github.com/khoakmp/brgame/utils"
)

var appName = "bloody-roar-2"

func joinAsClient() {
	reqHeader := http.Header{}
	clientID := utils.RandString(6)
	reqHeader.Add("Client_id", clientID)
	reqHeader.Add("Role", "client")
	fmt.Printf("[CLIENT %s] Try connecting:\n", clientID)

	conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/ws", reqHeader)
	if err != nil {
		fmt.Printf("[CLIENT %s] Failed to dial ws %s\n", clientID, err)
		return
	}

	defer conn.Close()

	payload := api.RequestGamePayload{
		AppName: appName,
		Mode:    "multi",
	}
	buf, _ := json.Marshal(&payload)
	reqGameMsg := api.Message{
		SessionID:   "",
		SenderID:    clientID,
		ReceiverIDs: nil,
		Type:        api.MessageRequestGame,
		Payload:     string(buf),
	}

	err = conn.WriteJSON(&reqGameMsg)
	if err != nil {
		fmt.Printf("[CLIENT %s] Failed to write ReqGame message %s\n", clientID, err)
	}

	/* if r, ok := coord.rooms.GetRoom(appName); ok {
		r.PrintClients(appName)
	} */
	var msg api.Message
	if err := conn.ReadJSON(&msg); err != nil {
		fmt.Printf("client %s Failed to read message %s\n", clientID, err)
		return
	}
	fmt.Printf("[CLIENT %s] recv msg Type %s,payload %s\n", clientID, msg.Type, msg.Payload)

	/* if r, ok := coord.rooms.GetRoom(appName); ok {
		r.PrintClients(appName)
	} */
}

func joinAsWorker() {
	clientID := utils.RandString(6)

	defer func() {
		fmt.Printf("[WORKER %s] Closed\n", clientID)
	}()

	reqHeader := http.Header{}
	reqHeader.Add("Client_id", clientID)
	reqHeader.Add("Role", "worker")
	fmt.Printf("[WORKER %s] Joining as worker ID\n", clientID)
	conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/ws", reqHeader)
	if err != nil {
		fmt.Printf("[WORKER %s]Failed to dial ws %s\n", clientID, err)
		return
	}
	defer conn.Close()

	fmt.Println(conn.LocalAddr().String())
	for {
		var msg api.Message
		err := conn.ReadJSON(&msg)
		if err != nil {

			fmt.Printf("[WORKER %s]Failed to read message %s\n", clientID, err)
			return
		}

		fmt.Printf("[WORKER %s] recv msg type: %s, payload: %s\n", clientID, msg.Type, msg.Payload)

	}
}

func TestCoord(t *testing.T) {
	coord := New()
	go joinAsWorker()
	numClient := 10
	for i := 0; i < numClient; i++ {
		go joinAsClient()
	}
	//go printStats(coord)
	coord.Run(":8080")
}
