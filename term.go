package chip8

import (
	"time"

	"github.com/nsf/termbox-go"
)

type Terminal struct {
	fg, bg termbox.Attribute
}

func NewTerminal() (*Terminal, error) {
	return &Terminal{termbox.ColorWhite, termbox.ColorBlack}, initTerm()
}

func (d *Terminal) Render(i IterableImage) {
	i.OnEachPixel(func(x, y int, i WriteableImage) {
		c := i.At(x, y)
		if c != 0 {
			termbox.SetCell(x, y, '\u2588', d.fg, d.bg)
		} else {
			termbox.SetCell(x, y, ' ', d.fg, d.bg)
		}
	})

	termbox.Flush()
}

func initTerm() error {
	if err := termbox.Init(); err != nil {
		return err
	}
	termbox.HideCursor()
	if err := termbox.Clear(0, 0); err != nil {
		return err
	}

	return termbox.Flush()
}

type TermKeypad struct {
	state       [16]bool
	keyMap      map[rune]byte
	wantnext    bool
	key         chan byte
	keyUpTimers map[rune]*time.Timer
}

func (k *TermKeypad) receiveEvents(event chan<- termbox.Event) {
	for {
		e := termbox.PollEvent()
		if e.Type == termbox.EventInterrupt {
			event <- e
		} else if e.Type != termbox.EventKey {
			return
		}
		if v, ok := k.keyMap[e.Ch]; ok == true {
			// HACK: reset the timer for that key. On expiration mark as up.
			k.keyUpTimers[e.Ch].Reset(100 * time.Millisecond)
			k.state[v] = true
			// If the emulator is in WaitPress then send it the key
			if k.wantnext == true {
				k.key <- v
			}
		} else if e.Ch == '`' {
			event <- e
		}
	}
}

func NewTermKeypad(event chan<- termbox.Event) *TermKeypad {
	// Map left side of the keyboard, 1234qwerasdfzxcv to their keyboard
	k := &TermKeypad{
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
		c := code
		k.keyUpTimers[key] = time.AfterFunc(0, func() {
			k.state[c] = false
		})
	}
	go k.receiveEvents(event)
	return k
}

func (k *TermKeypad) Pressed(key uint8) bool {
	return k.state[key]
}

func (k *TermKeypad) WaitPress() uint8 {
	k.wantnext = true
	v := <-k.key
	k.wantnext = false
	return v
}
