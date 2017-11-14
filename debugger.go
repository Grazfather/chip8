package chip8

import (
	"bufio"
	"encoding/binary"
	"fmt"
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

//const PROMPT = "\033[31m>>> \033[0m"

// TODO: Make part of debugger
var stopch chan os.Signal
var stop bool
var tick = time.Tick(2 * time.Millisecond)

type Debugger struct {
	c *Chip8
}

func NewDebugger(c *Chip8) *Debugger {
	return &Debugger{c: c}
}

func (d *Debugger) Start() {
	reader := bufio.NewReader(os.Stdin)

	stopch := make(chan os.Signal, 1)
	signal.Notify(stopch, os.Interrupt)
	go func() {
		for {
			s := <-stopch
			if !stop {
				fmt.Println("Got ", s)
				stop = true
			}
		}
	}()

	var last string
	for {
		d.PrintState()
		fmt.Printf(PROMPT)
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(err)
			panic(err)
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
			fmt.Printf("V%02d: "+white("%02X")+", ", i*4+j, d.c.v[i*4+j])
		}
		fmt.Printf("\n")
	}
	fmt.Println(green("-- ") + yellow("Assembly") + green(" --"))
	dis := &Disassembler{}
	for i := uint16(4); i > 0; i -= 2 {
		fmt.Printf("0x%04X %04X %s\n",
			d.c.pc-i,
			binary.BigEndian.Uint16(d.c.mem[d.c.pc-i:]),
			dis.dis(d.c.mem[d.c.pc-i:]))
	}
	fmt.Printf(white("0x%04X")+green(" %04X ")+blue("%s\n"),
		d.c.pc,
		binary.BigEndian.Uint16(d.c.mem[d.c.pc:]),
		dis.dis(d.c.mem[d.c.pc:]))
	for i := uint16(2); i < 16; i += 2 {
		fmt.Printf("0x%04X"+green(" %04X ")+cyan("%s\n"),
			d.c.pc+i,
			binary.BigEndian.Uint16(d.c.mem[d.c.pc+i:]),
			dis.dis(d.c.mem[d.c.pc+i:]))
	}

}

var commands = map[string]func(*Debugger, []string){
	"reset": func(d *Debugger, ops []string) {
		fmt.Println("Reseting CPU")
		d.c.Reset()
	},
	"r": func(d *Debugger, ops []string) {
		fmt.Println("Running")
		for stop == false {
			select {
			case <-tick:
				err := d.c.RunOne()
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					stop = true
				}
			case <-stopch:
				fmt.Println("Got Ctrl-C")
				stop = true
			}
		}
	},
	"s": func(d *Debugger, ops []string) {
		err := d.c.RunOne()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			stop = true
		}
	},
	"x": func(d *Debugger, ops []string) {
		if len(ops) != 1 {
			fmt.Println("usage: x ADDR")
		}
		addr, err := strconv.ParseUint(ops[0], 0, 16)
		if err != nil {
			fmt.Println("Couldn't parse address from", ops[0])
		}
		if int(addr) >= len(d.c.mem) {
			fmt.Println("Addr out of range")
		}
		fmt.Printf("%#04x: %02x\n", addr, d.c.mem[addr])
	},
	"q": func(d *Debugger, ops []string) {
		fmt.Println("goodbye.")
		os.Exit(0)
	},
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
