package api

// These message type for client
const (
	MessageRequestGame       = "req_game"
	MessageWaitTimeout       = "wait_timeout"
	MessageSessionCreated    = "session_created"
	MessageWorkerNotFound    = "worker_not_found"
	MessageRequestGameFailed = "request_game_failed"
)

const (
	MessageSDP          = "sdp"
	MessageICECandidate = "ice_candidate"
)

const (
	MessageStartSession = "start_session"
	MessageHearBeat     = "hb"
)
const (
	ModeMulti  = "multi"
	ModeSingle = "single"
)

const (
	RoleWorker = "worker"
	RoleClient = "client"
)
