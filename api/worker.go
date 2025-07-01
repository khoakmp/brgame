package api

import (
	"bytes"
	"fmt"
	"log"
	"strings"
)

type StartSessionPayload struct {
	ClientIDs []string `json:"client_ids"`
	AppName   string   `json:"app_name"`
}

func (p *StartSessionPayload) Print() {
	buffer := bytes.NewBuffer(nil)
	buffer.WriteString("StartSessionPayload:\nClientIDs:")
	buffer.WriteString(strings.Join(p.ClientIDs, ","))
	buffer.WriteString(fmt.Sprintf("\nAppName:%s\n", p.AppName))
	log.Println(buffer.String())
}
