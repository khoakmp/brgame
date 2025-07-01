package api

type RequestGamePayload struct {
	AppName string `json:"app_name"`
	Mode    string `json:"mode"` // single or multi
}

type SessionCreatedPayload struct {
	WorkerID  string   `json:"worker_id"`
	SessionID string   `json:"session_id"`
	PlayerIDs []string `json:"player_ids"`
}
