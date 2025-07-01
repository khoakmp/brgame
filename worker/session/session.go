package session

import (
	"github.com/khoakmp/brgame/api"
)

type Session interface {
	ID() string
	Stats() Stats
	Receive(api.Message)
}
