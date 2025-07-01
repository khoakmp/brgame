package event

type MouseEventPayload struct {
	IsLeft byte    `json:"isleft"`
	X      float32 `json:"x"`
	Y      float32 `json:"y"`
	Width  float32 `json:"w"`
	Height float32 `json:"h"`
}

type KeyEventPayload struct {
	Keycode int `json:"keycode"`
}
