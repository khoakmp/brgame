export type InputPacket = {
  type : string;
  data: string;
}

export type MouseEventPayload = {
  isleft: number;
  x: number;
  y: number;
  w: number;
  h: number
}

export type KeyEventPayload = {
  keycode: number;
}
export enum EventType {
  KEY_UP     = "KEYUP",
  KEY_DOWN   = "KEYDOWN",
  MOUSE_UP   = "MOUSEUP",
  MOUSE_DOWN = "MOUSEDOWN",
  MOUSE_MOVE = "MOUSEMOVE"
}