package chip8

type Keypad interface {
	Pressed(key uint8) bool
	WaitPress() uint8
}

type NoKeypad struct{}

func (k *NoKeypad) Pressed(key uint8) bool {
	return false
}

func (k *NoKeypad) WaitPress() uint8 {
	panic("Cannot get input from NoKeypad")
}
