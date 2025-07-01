package network

import "github.com/khoakmp/brgame/api"

type Conn interface {
	ReadMessage() (api.Message, error)
	WriteMessage(*api.Message) error
	Close() error
}
