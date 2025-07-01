package game

var TranslateKeyFnMap map[string]func(peerOrder int, keycode int) int

func init() {
	TranslateKeyFnMap = make(map[string]func(peerOrder int, keycode int) int)
	TranslateKeyFnMap["br2"] = func(peerOrder int, keycode int) int {
		// TODO:
		if peerOrder == 0 {
			return keycode
		}
		if keycode > 36 && keycode < 41 {
			return keycode
		}

		switch keycode {
		case 65:
			return 77
		case 83:
			return 86
		case 90:
			return 66
		case 88:
			return 78
		}
		return 256
	}
}
