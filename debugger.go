package chip8

import (
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/jroimartin/gocui"
)

var yellow = color.New(color.FgYellow).SprintFunc()
var red = color.New(color.FgRed).SprintFunc()
var blue = color.New(color.FgBlue).SprintFunc()
var green = color.New(color.FgGreen).SprintFunc()
var cyan = color.New(color.FgCyan).SprintFunc()
var white = color.New(color.FgWhite, color.Bold).SprintFunc()

// TODO: Make part of debugger
var stop bool
var stopped bool
var first bool
var tick = time.Tick(2 * time.Millisecond)

func parseAddr(s string) (uint16, error) {
	addr, err := strconv.ParseUint(s, 0, 16)
	if err != nil {
		return 0, fmt.Errorf("couldn't parse address from %s", s)
	}
	if int(addr) >= 0x1000 {
		return 0, fmt.Errorf("addr out of range")
	}
	return uint16(addr), nil
}

type Debugger struct {
	c    *Chip8
	bps  map[uint16]bool
	tbps map[uint16]bool
	dis  Disassembler
	ui   *ui
	last string
}

type ui struct {
	*gocui.Gui
	displayView *gocui.View
	debugView   *gocui.View
	promptView  *gocui.View
}

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
		g.SetKeybinding("display", key, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
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

func NewDebugger() *Debugger {
	return &Debugger{
		c:    nil,
		bps:  make(map[uint16]bool),
		tbps: make(map[uint16]bool),
		dis:  Disassembler{},
	}
}

// The promptEditor adds readline keys and assumes one line
type promptEditor struct {
	gocui.Editor
}

func (e promptEditor) Edit(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	switch {
	case key == gocui.KeyEnter:
		return
	case key == gocui.KeyCtrlU:
		ox, _ := v.Cursor()
		for ox > 0 {
			v.EditDelete(true)
			ox, _ = v.Cursor()
		}
		return
	case key == gocui.KeyCtrlD:
		key = gocui.KeyDelete
	case key == gocui.KeyCtrlB:
		key = gocui.KeyArrowLeft
	case key == gocui.KeyCtrlF:
		key = gocui.KeyArrowRight
		fallthrough
	case key == gocui.KeyArrowRight:
		ox, _ := v.Cursor()
		if ox >= len(v.Buffer())-1 {
			return
		}
	case key == gocui.KeyHome || key == gocui.KeyArrowUp || key == gocui.KeyCtrlA:
		v.SetCursor(0, 0)
		v.SetOrigin(0, 0)
		return
	case key == gocui.KeyEnd || key == gocui.KeyArrowDown || key == gocui.KeyCtrlE:
		width, _ := v.Size()
		lineWidth := len(v.Buffer()) - 1
		if lineWidth > width {
			v.SetOrigin(lineWidth-width, 0)
			lineWidth = width - 1
		}
		v.SetCursor(lineWidth, 0)
		return
	}
	e.Editor.Edit(v, key, ch, mod)
}

func (ui *ui) layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	var err error
	if maxY < SCREEN_HEIGHT || maxX < SCREEN_WIDTH {
		return fmt.Errorf("Cannot display if less than %d x %d! Resize your terminal! (^Q to quit)",
			SCREEN_WIDTH, SCREEN_HEIGHT)
	}
	// TODO: Choose vertical or horizontal layout if only one would work
	left := (maxX - SCREEN_WIDTH) / 2
	ui.displayView, err = g.SetView("display", left, 0, SCREEN_WIDTH+2+left, SCREEN_HEIGHT+2)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	}
	ui.displayView.Title = "display"

	ui.debugView, err = g.SetView("debug", -1, SCREEN_HEIGHT+3, maxX, maxY-2)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	}
	ui.debugView.Title = "debug"
	ui.debugView.Wrap = false
	ui.debugView.Autoscroll = true
	ui.promptView, err = g.SetView("prompt", -1, maxY-2, maxX, maxY)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	}
	ui.promptView.Title = "prompt"
	ui.promptView.Wrap = false
	ui.promptView.Editable = true
	ui.promptView.Editor = &promptEditor{gocui.DefaultEditor}
	ui.promptView.Autoscroll = true
	return nil
}

func (d *Debugger) quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func (d *Debugger) halt(g *gocui.Gui, v *gocui.View) error {
	if stopped {
		g.Update(func(g *gocui.Gui) error {
			d.Printf("Already stopped. Press Ctrl-Q or q to quit\n")
			return nil
		})
		return nil
	}
	g.Update(func(g *gocui.Gui) error {
		d.Println("Received halt")
		return nil
	})
	stop = true
	return nil
}

func (d *Debugger) cont() {
	stop = false
	stopped = false
	first = true
	d.ui.Cursor = false
	d.ui.SetCurrentView("display")
}

func (d *Debugger) RunOne() {
	err := d.c.RunOne()
	stopped = false
	if err != nil {
		d.Println(err)
	}
	stop = true
}
func (ui *ui) swapFocus(g *gocui.Gui, v *gocui.View) error {
	currentView := g.CurrentView()
	if currentView == nil {
		if _, err := g.SetCurrentView("prompt"); err != nil {
			return err
		}
		return nil
	}
	if currentView.Name() == "prompt" {
		if _, err := g.SetCurrentView("display"); err != nil {
			return err
		}
		g.Cursor = false
	} else {
		if _, err := g.SetCurrentView("prompt"); err != nil {
			return err
		}
	}
	return nil
}

func (d *Debugger) Println(a ...interface{}) {
	fmt.Fprintln(d.ui.debugView, a...)
}

func (d *Debugger) Printf(format string, a ...interface{}) {
	fmt.Fprintf(d.ui.debugView, format, a...)
}

func (d *Debugger) printContext() error {
	d.ui.Update(func(g *gocui.Gui) error {
		d.printState()
		return nil
	})
	return nil
}

func (d *Debugger) cleanPrompt() {
	v := d.ui.promptView
	d.ui.SetCurrentView("prompt")
	v.Clear()
	// TODO: Get the y coord
	d.ui.Cursor = true
	v.SetCursor(0, 0)
}

func (d *Debugger) Start(rom string) {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	d.ui = &ui{g, nil, nil, nil}
	g.SetManagerFunc(d.ui.layout)
	// HACK: layout needs to have been called so grab handles to views
	d.ui.layout(g)
	g.SetCurrentView("display")

	k := NewGocuiKeypad(g, d.ui.displayView)
	r := NewGocuiRenderer(d.ui.displayView)
	d.c = NewChip8(r, k)
	d.c.Reset()

	if err := d.c.LoadBinary(rom); err != nil {
		fmt.Printf("Error loading %s: %v\n", os.Args[1], err)
		os.Exit(1)
	}

	if err := g.SetKeybinding("", gocui.KeyCtrlQ, gocui.ModNone, d.quit); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, d.halt); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("", gocui.KeyTab, gocui.ModNone, d.ui.swapFocus); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("prompt", gocui.KeyEnter, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		err := d.Handle(strings.Trim(v.Buffer(), "\n"))
		d.cleanPrompt()
		return err
	}); err != nil {
		log.Panicln(err)
	}

	go d.run()

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}

func (d *Debugger) run() {
	go d.c.KeepTime()
	d.printContext()

	first = true // To allow us to run while on a bp
LOOP:
	for {
		var err error
		select {
		case <-tick:
			if stopped {
				continue
			}
			if v, ok := d.bps[d.c.pc]; ok && v && !first {
				d.ui.Update(func(g *gocui.Gui) error {
					d.Printf(red("Hit breakpoint at 0x%04X\n"), d.c.pc)
					return nil
				})
				stop = true
			}
			if v, ok := d.tbps[d.c.pc]; ok && v && !first {
				d.ui.Update(func(g *gocui.Gui) error {
					d.Printf(red("Hit temp breakpoint at 0x%04X\n"), d.c.pc)
					return nil
				})
				removeBreakpoint(d.tbps, d.c.pc)
				stop = true
			}
			first = false
			if !stop {
				err = d.c.RunOne()
			}
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				break LOOP
			}
			if stop {
				stopped = true
				d.printContext()
				d.cleanPrompt()
			}
		case <-d.c.Renderch:
			d.ui.Update(func(g *gocui.Gui) error {
				d.c.Render()
				return nil
			})
		}
	}
}

func (d *Debugger) PrintAsm(addr, count int) {
}

func (d *Debugger) printState() {
	d.Println(green("-- ") + yellow("Registers") + green(" --"))
	d.Printf("PC: "+white("0x%04X")+" I: "+white("0x%04X\n"), d.c.pc, d.c.i)
	d.Printf("Delay: "+white("0x%02X")+" Sound: "+white("0x%02X\n"), d.c.delay, d.c.sound)
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			d.Printf("V%X: "+white("%02X")+", ", i*4+j, d.c.v[i*4+j])
		}
		d.Printf("\n")
	}
	d.Println(green("-- ") + yellow("Assembly") + green(" --"))
	// Print a few instructions back
	for i := uint16(4); i > 0; i -= 2 {
		addr := d.c.pc - i
		if addr < d.c.pc {
			d.Printf("0x%04X %04X %s\n",
				addr,
				binary.BigEndian.Uint16(d.c.mem[addr:]),
				d.dis.dis(d.c.mem[addr:]))
		}
	}
	// Print current instruction
	ins := d.dis.dis(d.c.mem[d.c.pc:])
	d.Printf(white("0x%04X")+green(" %04X ")+blue("%s\n"),
		d.c.pc,
		binary.BigEndian.Uint16(d.c.mem[d.c.pc:]),
		ins)
	// If we're on a call, peek at its dest
	i := uint16(2)
	if ins.isCall() {
		addr := ins.callTarget()
		d.Printf("⤷  0x%04X"+green(" %04X ")+cyan("%s\n"),
			addr+i,
			binary.BigEndian.Uint16(d.c.mem[addr+i:]),
			d.dis.dis(d.c.mem[addr+i:]))
		i += 2
		for ; i < 8 && int(addr+i) < len(d.c.mem); i += 2 {
			d.Printf("   0x%04X"+green(" %04X ")+cyan("%s\n"),
				addr+i,
				binary.BigEndian.Uint16(d.c.mem[addr+i:]),
				d.dis.dis(d.c.mem[addr+i:]))
		}
	}
	// Print a few instructions forward
	for ; i < 16 && int(d.c.pc+i) < len(d.c.mem); i += 2 {
		addr := d.c.pc + i
		d.Printf("0x%04X"+green(" %04X ")+cyan("%s\n"),
			addr,
			binary.BigEndian.Uint16(d.c.mem[addr:]),
			d.dis.dis(d.c.mem[addr:]))
	}

}

var commands = map[string]func(*Debugger, []string){
	"reset": reset,
	"ctx":   context,
	"ib":    breakpoints,
	"b":     addBreak,
	"tb":    addTBreak,
	"db":    disableBreak,
	"dtb":   disableTBreak,
	"eb":    enableBreak,
	"etb":   enableTBreak,
	"rb":    removeBreak,
	"rtb":   removeTBreak,
	"c":     cont,
	"s":     step,
	"si":    step,
	"n":     next,
	"ni":    next,
	"x":     examine,
	"e":     edit,
	"q":     quit,
}

func (d *Debugger) Handle(line string) error {
	if line == "" && d.last != "" {
		line = d.last
	}
	ops := strings.Split(line, " ")
	cmd := ops[0]
	ops = ops[1:]
	if f, ok := commands[cmd]; ok {
		f(d, ops)
		d.last = line
	} else {
		d.Printf("illegal command: '%s'\n", cmd)
	}
	return nil
}
