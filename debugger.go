package chip8

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"
)

const PROMPT = "\033[31m>>> \033[0m"

// TODO: Make part of debugger
var stopch chan os.Signal
var stop bool
var tick = time.Tick(2 * time.Millisecond)

type Debugger struct {
	c       *Chip8
	verbose bool
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
			fmt.Println("Got signal", s)
			stop = true
		}
	}()

	for {
		fmt.Printf(PROMPT)
		line, err := reader.ReadString('\n')
		line = strings.TrimSpace(line)
		if err != nil {
			fmt.Println(err)
			panic(err)
		}
		err = d.Handle(line)
		if err != nil {
			fmt.Println(err)
		}
		stop = false
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
				if d.verbose {
					fmt.Printf("%v\n", d.c)
				}
			case <-stopch:
				fmt.Println("Got Ctrl-C")
				stop = true
			}
		}
		fmt.Printf("%v\n", d.c)
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
