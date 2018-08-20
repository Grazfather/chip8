package chip8

import (
	"time"

	"github.com/jroimartin/gocui"
)

type gocuiKeypad struct {
	state       [16]bool
	keyMap      map[rune]byte
	wantnext    bool
	key         chan byte
	keyUpTimers map[rune]*time.Timer
}

func NewGocuiKeypad(g *gocui.Gui, view *gocui.View) *gocuiKeypad {
	k := &gocuiKeypad{
		keyMap: map[rune]byte{
			'1': 0x1,
			'2': 0x2,
			'3': 0x3,
			'4': 0xC,
			'q': 0x4,
			'w': 0x5,
			'e': 0x6,
			'r': 0xD,
			'a': 0x7,
			's': 0x8,
			'd': 0x9,
			'f': 0xE,
			'z': 0xA,
			'x': 0x0,
			'c': 0xB,
			'v': 0xF,
		},
	}

	k.key = make(chan uint8)
	k.keyUpTimers = make(map[rune]*time.Timer)

	for key, code := range k.keyMap {
		key := key
		c := code
		k.keyUpTimers[key] = time.AfterFunc(0, func() {
			k.state[c] = false
		})
		g.SetKeybinding(view.Name(), key, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
			// HACK: reset the timer for that key. On expiration mark as up.
			k.keyUpTimers[key].Reset(100 * time.Millisecond)
			k.state[c] = true
			// If the emulator is in WaitPress then send it the key
			if k.wantnext == true {
				k.key <- c
			}
			return nil
		})
	}
	return k
}

func (k *gocuiKeypad) Pressed(key uint8) bool {
	return k.state[key]
}

func (k *gocuiKeypad) WaitPress() uint8 {
	k.wantnext = true
	v := <-k.key
	k.wantnext = false
	return v
}

type gocuiRenderer struct {
	*gocui.View
}

func NewGocuiRenderer(view *gocui.View) *gocuiRenderer {
	return &gocuiRenderer{view}
}

func (d *gocuiRenderer) Render(i IterableImage) {
	i.OnEachPixel(func(x, y int, i WriteableImage) {
		d.SetCursor(x, y)
		d.EditDelete(false)

		c := i.At(x, y)
		if c != 0 {
			d.EditWrite('\u2588')
		} else {
			d.EditWrite(' ')
		}
	})
}
