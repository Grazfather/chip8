package chip8

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
)

var yellow = color.New(color.FgYellow).SprintFunc()
var red = color.New(color.FgRed).SprintFunc()
var blue = color.New(color.FgBlue).SprintFunc()
var green = color.New(color.FgGreen).SprintFunc()
var cyan = color.New(color.FgCyan).SprintFunc()
var white = color.New(color.FgWhite, color.Bold).SprintFunc()

var PROMPT = red(">>> ")

// TODO: Make part of debugger
var stopch chan os.Signal
var stop bool
var stopped bool
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
}

func NewDebugger(c *Chip8) *Debugger {
	return &Debugger{
		c:    c,
		bps:  make(map[uint16]bool),
		tbps: make(map[uint16]bool),
		dis:  Disassembler{}}
}

func (d *Debugger) Start() {
	reader := bufio.NewReader(os.Stdin)

	stopch := make(chan os.Signal, 1)
	signal.Notify(stopch, os.Interrupt)
	go func() {
		for {
			s := <-stopch
			if stopped {
				fmt.Printf("\nAlready stopped. Press Ctrl-D to or q to quit\n")
				fmt.Printf(PROMPT)
				continue
			}
			if !stop {
				fmt.Println("Got ", s)
				stop = true
			}
		}
	}()

	var last string
	stop = true
	for {
		if stop {
			d.PrintState()
			stop = false
			stopped = true
		}
		fmt.Printf(PROMPT)
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Printf("\n")
				return
			}
			fmt.Println(err)
		}
		line = strings.TrimSpace(line)

		// A blank line means repeat the last
		if line == "" {
			line = last
		}
		last = line
		err = d.Handle(line)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func (d *Debugger) PrintAsm(addr, count int) {
}
func (d *Debugger) PrintState() {
	fmt.Println(green("-- ") + yellow("Registers") + green(" --"))
	fmt.Printf("PC: "+white("0x%04X")+" I: "+white("0x%04X\n"), d.c.pc, d.c.i)
	fmt.Printf("Delay: "+white("0x%02X")+" Sound: "+white("0x%02X\n"), d.c.delay, d.c.sound)
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			fmt.Printf("V%X: "+white("%02X")+", ", i*4+j, d.c.v[i*4+j])
		}
		fmt.Printf("\n")
	}
	fmt.Println(green("-- ") + yellow("Assembly") + green(" --"))
	// Print a few instructions back
	for i := uint16(4); i > 0; i -= 2 {
		addr := d.c.pc - i
		if addr < d.c.pc {
			fmt.Printf("0x%04X %04X %s\n",
				addr,
				binary.BigEndian.Uint16(d.c.mem[addr:]),
				d.dis.dis(d.c.mem[addr:]))
		}
	}
	// Print current instruction
	ins := d.dis.dis(d.c.mem[d.c.pc:])
	fmt.Printf(white("0x%04X")+green(" %04X ")+blue("%s\n"),
		d.c.pc,
		binary.BigEndian.Uint16(d.c.mem[d.c.pc:]),
		ins)
	// If we're on a call, peek at its dest
	i := uint16(2)
	if ins.isCall() {
		addr := ins.callTarget()
		fmt.Printf("⤷  0x%04X"+green(" %04X ")+cyan("%s\n"),
			addr+i,
			binary.BigEndian.Uint16(d.c.mem[addr+i:]),
			d.dis.dis(d.c.mem[addr+i:]))
		i += 2
		for ; i < 8 && int(addr+i) < len(d.c.mem); i += 2 {
			fmt.Printf("   0x%04X"+green(" %04X ")+cyan("%s\n"),
				addr+i,
				binary.BigEndian.Uint16(d.c.mem[addr+i:]),
				d.dis.dis(d.c.mem[addr+i:]))
		}
	}
	// Print a few instructions forward
	for ; i < 16 && int(d.c.pc+i) < len(d.c.mem); i += 2 {
		addr := d.c.pc + i
		fmt.Printf("0x%04X"+green(" %04X ")+cyan("%s\n"),
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
	ops := strings.Split(line, " ")
	cmd := ops[0]
	ops = ops[1:]
	if f, ok := commands[cmd]; ok {
		f(d, ops)
	} else {
		return fmt.Errorf("illegal command: '%s'", cmd)
	}
	return nil
}
