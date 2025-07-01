export enum MessageType {
  RequestGame = "req_game",
  WaitTimeout       = "wait_timeout",
  SessionCreated    = "session_created",
  WorkerNotFound    = "worker_not_found",
  RequestGameFailed = "request_game_failed",
  SDP          = "sdp",
  ICECandidate = "ice_candidate"
}
/* 
const MessageRequestGame = "req_game"
const MessageWaitTimeout       = "wait_timeout"
const	MessageSessionCreated    = "session_created"
const	MessageWorkerNotFound    = "worker_not_found"
const	MessageRequestGameFailed = "request_game_failed"

const MessageSDP          = "sdp"
const	MessageICECandidate = "ice_candidate" */

export type WsMessage = {
  session_id: string;
  sender_id: string;
  receiver_ids: string[];
  type: string;
  payload: string;
}
export type RequestGamePayload = {
  app_name :string;
  mode :string;
}

export type SessionCreatedPayload = {
  worker_id :string;
  session_id:string;
  player_ids: string[];
}

