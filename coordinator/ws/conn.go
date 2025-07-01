package ws

import (
	"log"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/khoakmp/brgame/api"
)

type WsConn struct {
	conn *websocket.Conn
}

var upgrader websocket.Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true }, // TODO: modify later
}

func ServeReq(w http.ResponseWriter, r *http.Request, respHeader http.Header) (*WsConn, error) {
	conn, err := upgrader.Upgrade(w, r, respHeader)
	if err != nil {
		log.Printf("Failed to create ws conn, %s\n", err)
		return nil, err
	}

	tcpConn := conn.UnderlyingConn().(*net.TCPConn)
	tcpConn.SetKeepAlive(true)
	tcpConn.SetKeepAlivePeriod(time.Second * 2)

	ws := WsConn{
		conn: conn,
	}
	return &ws, nil
}

func (w *WsConn) ReadMessage() (msg api.Message, err error) {
	err = w.conn.ReadJSON(&msg)
	return
}

func (w *WsConn) WriteMessage(msg *api.Message) error {
	return w.conn.WriteJSON(msg)
}

func (w *WsConn) Close() error {
	return w.conn.Close()
}
