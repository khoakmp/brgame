package api

type Message struct {
	SessionID   string   `json:"session_id"`
	SenderID    string   `json:"sender_id"`
	ReceiverIDs []string `json:"receiver_ids"`
	Type        string   `json:"type"`
	Payload     string   `json:"payload"`
}
