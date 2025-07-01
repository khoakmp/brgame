package coordinator

import (
	"encoding/json"

	"github.com/khoakmp/brgame/api"
)

type MessageBuilder struct {
}

var MsgBuilder MessageBuilder

func (mb *MessageBuilder) SessionCreated(workerID, sessionID string, playerIDs []string) *api.Message {
	payload := api.SessionCreatedPayload{
		WorkerID:  workerID,
		SessionID: sessionID,
		PlayerIDs: playerIDs,
	}
	buf, _ := json.Marshal(payload)

	return &api.Message{
		SessionID:   sessionID,
		SenderID:    "coordinator_",
		ReceiverIDs: nil,
		Type:        api.MessageSessionCreated,
		Payload:     string(buf),
	}
}

func (mb *MessageBuilder) WorkerNotFound() *api.Message {
	return &api.Message{
		Type:     api.MessageWorkerNotFound,
		SenderID: "coordinator_",
		Payload:  "serivce unavailable",
	}
}

func (mb *MessageBuilder) StartSession(sessionID, appName string, clientIDs []string, workerID string) *api.Message {
	payload := api.StartSessionPayload{
		ClientIDs: clientIDs,
		AppName:   appName,
	}

	buf, _ := json.Marshal(payload)
	startMessage := api.Message{
		SessionID:   sessionID,
		SenderID:    "coordinator_",
		ReceiverIDs: nil,
		Type:        api.MessageStartSession,
		Payload:     string(buf),
	}
	return &startMessage
}
