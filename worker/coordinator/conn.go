package coordinator

import (
	"errors"

	"github.com/khoakmp/brgame/api"
)

type Conn interface {
	ReadMessage() (api.Message, error)
	WriteMessage(api.Message) error
	Close() error
}

var ErrInvalidProtocol = errors.New("invalid protocol")

func Connect(protocol, url, workerID string) (Conn, error) {
	switch protocol {
	case "ws":
		return ConnectWs(url, workerID)
	default:
		return nil, ErrInvalidProtocol
	}
}
