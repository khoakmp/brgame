package coordinator

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/khoakmp/brgame/api"
	"github.com/khoakmp/brgame/coordinator/client"
	"github.com/khoakmp/brgame/coordinator/worker"
	"github.com/khoakmp/brgame/coordinator/ws"
	"github.com/khoakmp/brgame/utils"
)

type Coordinator struct {
	workers *worker.Hub
	clients *client.Hub
	rooms   *client.HubRoomWaiting
}

func (c *Coordinator) HandleClientMessage(msg *api.Message, cli *client.Client) {
	if len(msg.ReceiverIDs) > 0 {
		w, ok := c.workers.GetWorker(msg.ReceiverIDs[0])

		if !ok {
			msg := MsgBuilder.WorkerNotFound()
			cli.SendMessage(msg)

			return
		}
		w.SendMessage(msg)
		return
	}

	switch msg.Type {
	case api.MessageRequestGame:
		var req api.RequestGamePayload
		json.Unmarshal([]byte(msg.Payload), &req)
		if req.Mode == api.ModeMulti {
			c.handleGameMulti(req.AppName, cli)
			return
		}
		c.handleGameSingle(req.AppName, cli)
		// TODO: handle other case
	}

}

func (c *Coordinator) handleGameSingle(appName string, cli *client.Client) {
	sessionID := utils.RandString(6)
	wrk := c.workers.RandomWorker()
	if wrk == nil {
		msg := MsgBuilder.WorkerNotFound()
		cli.SendMessage(msg)
		return
	}
	//wrk.StartSession(sessionID, appName, []string{cli.ID()})

	startMsg := MsgBuilder.StartSession(sessionID, appName, []string{cli.ID()}, wrk.ID())
	wrk.SendMessage(startMsg)

	msg := MsgBuilder.SessionCreated(wrk.ID(), sessionID, []string{cli.ID()})
	//msg := c.MsgSessionCreated(wrk.ID(), sessionID, []string{cli.ID()})
	cli.SendMessage(msg)
}

func (c *Coordinator) handleGameMulti(appName string, cli *client.Client) {
	room, ok := c.rooms.GetRoom(appName)
	if !ok {
		log.Println("Not found room", appName)
		return
	}

	clients := room.Process(cli, 2)
	// TODO: remove later
	room.PrintClients(appName)

	if clients == nil {
		fmt.Printf("[Client %s] in waiting room app[%s]\n", cli.ID(), appName)
		return
	}

	wrk := c.workers.RandomWorker()
	var clientMsg *api.Message
	if wrk == nil {
		clientMsg = MsgBuilder.WorkerNotFound()
	} else {
		sessionID := utils.RandString(6)
		var clientIDs []string = make([]string, len(clients))
		for i, c := range clients {
			clientIDs[i] = c.ID()
		}

		clientMsg = MsgBuilder.SessionCreated(wrk.ID(), sessionID, clientIDs)
		startMessage := MsgBuilder.StartSession(sessionID, appName, clientIDs, wrk.ID())
		wrk.SendMessage(startMessage)
	}

	for _, c := range clients {
		c.SendMessage(clientMsg)
	}
}

func (c *Coordinator) HandleWorkerMessage(msg *api.Message, worker *worker.Worker) {
	if len(msg.ReceiverIDs) > 0 {
		recvs := strings.Join(msg.ReceiverIDs, ",")
		log.Printf("[Worker %s] Forward message to clients: %s\n", worker.ID(), recvs)

		c.clients.ForwardMessage(msg)
	}
}

func (c *Coordinator) Run(addr string) {
	c.runHttp(addr)
}

func (c *Coordinator) Workers() *worker.Hub {
	return c.workers
}

func New() *Coordinator {
	var params []client.RoomParam = make([]client.RoomParam, 1)
	params[0] = client.RoomParam{
		AppName: "bloody_roar_2",
		NumSlot: 2,
	}
	c := Coordinator{
		workers: worker.NewHub(),
		clients: client.NewHub(),
		rooms:   client.NewHubRoomWaiting(params),
	}
	return &c
}
func getRoleAndID(r *http.Request) (string, string) {
	roles := r.Header["Role"]
	if len(roles) > 0 && len(r.Header["Client_id"]) > 0 {
		return roles[0], r.Header["Client_id"][0]
	}
	q := r.URL.Query()
	return q.Get("role"), q.Get("client_id")
}

func (c *Coordinator) runHttp(addr string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		role, clientID := getRoleAndID(r)
		if role == "" || clientID == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		/* roles := r.Header["Role"]

		if len(roles) == 0 || len(r.Header["Client_id"]) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			return
		} */

		//clientID := r.Header["Client_id"][0]
		conn, err := ws.ServeReq(w, r, http.Header{})
		if err != nil {
			return
		}

		switch role {
		case api.RoleWorker:
			worker := worker.New(clientID, conn, c.workers, c)
			c.workers.AddWorker(worker)

		case api.RoleClient:
			client := client.New(clientID, conn, c, c.clients)
			c.clients.AddClient(client)
		}
	})
	http.ListenAndServe(addr, mux)
}
