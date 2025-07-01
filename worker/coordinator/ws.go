package coordinator

import (
	"errors"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/khoakmp/brgame/api"
)

type WsConn struct {
	conn      *websocket.Conn
	writeLock sync.Mutex
}

func (ws *WsConn) ReadMessage() (msg api.Message, err error) {
	err = ws.conn.ReadJSON(&msg)
	return
}

func (ws *WsConn) WriteMessage(msg api.Message) error {
	ws.writeLock.Lock()
	defer ws.writeLock.Unlock()
	return ws.conn.WriteJSON(msg)
}

func (ws *WsConn) Close() error {
	return ws.conn.Close()
}

var ErrConnectWsFailed = errors.New("connect ws server failed")

func ConnectWs(url string, workerID string) (*WsConn, error) {
	var header http.Header = make(http.Header)

	header.Add("Client_id", workerID)
	header.Add("Role", "worker")

	c, _, err := websocket.DefaultDialer.Dial(url, header)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, ErrConnectWsFailed
	}
	wsconn := &WsConn{
		conn: c,
	}
	return wsconn, nil
}
