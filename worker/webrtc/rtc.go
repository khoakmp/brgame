package webrtc

type InputPacket struct {
	Type string `json:"type"`
	Data string `json:"data"`
}

type MouseEventPayload struct {
	IsLeft byte    `json:"isleft"`
	X      float32 `json:"x"`
	Y      float32 `json:"y"`
	Width  float32 `json:"w"`
	Height float32 `json:"h"`
}

type KeyEventPayload struct {
	KeyCode byte `json:"keycode"`
}
