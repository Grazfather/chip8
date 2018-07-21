package chip8

import (
	"fmt"
	"os"
	"strconv"
)

func reset(d *Debugger, ops []string) {
	fmt.Println("Reseting CPU")
	d.c.Reset()
}

func context(d *Debugger, ops []string) {
	// Setting stop makes the debugger show the context
	stop = true
}

func printBreakpoints(breaks map[uint16]bool) {
	// TODO: Sort in any way?
	// TODO: Count and display # times hit
	for a, v := range breaks {
		if v {
			fmt.Printf(green("+ 0x%04X\n"), a)
		} else {
			fmt.Printf("- 0x%04X\n", a)
		}
	}
}

func breakpoints(d *Debugger, ops []string) {
	if (len(d.bps) == 0) && (len(d.tbps) == 0) {
		fmt.Println(white("No breakpoints"))
		return
	}
	if len(d.bps) > 0 {
		fmt.Println(white("Breakpoints"))
		printBreakpoints(d.bps)
	}
	if len(d.tbps) > 0 {
		fmt.Println(white("Temp Breakpoints"))
		printBreakpoints(d.tbps)
	}
}

func addBreak(d *Debugger, ops []string) {
	if len(ops) != 1 {
		fmt.Println("Usage: b <addr>")
		return
	}
	addr, err := parseAddr(ops[0])
	if err != nil {
		fmt.Println(err)
		return
	}
	addBreakpoint(d.bps, addr)
}

func addTBreak(d *Debugger, ops []string) {
	if len(ops) != 1 {
		fmt.Println("Usage: tb <addr>")
		return
	}
	addr, err := parseAddr(ops[0])
	if err != nil {
		fmt.Println(err)
		return
	}
	addBreakpoint(d.tbps, addr)
}

func addBreakpoint(breaks map[uint16]bool, addr uint16) {
	breaks[addr] = true
}

func disableBreak(d *Debugger, ops []string) {
	if len(ops) != 1 {
		fmt.Println("Usage: db <addr>")
		return
	}
	addr, err := parseAddr(ops[0])
	if err != nil {
		fmt.Println(err)
		return
	}
	disableBreakpoint(d.bps, addr)
}

func disableTBreak(d *Debugger, ops []string) {
	if len(ops) != 1 {
		fmt.Println("Usage: dtb <addr>")
		return
	}
	addr, err := parseAddr(ops[0])
	if err != nil {
		fmt.Println(err)
		return
	}
	disableBreakpoint(d.tbps, addr)
}

func disableBreakpoint(breaks map[uint16]bool, addr uint16) {
	breaks[addr] = false
}

func enableBreak(d *Debugger, ops []string) {
	if len(ops) != 1 {
		fmt.Println("Usage: db <addr>")
		return
	}
	addr, err := parseAddr(ops[0])
	if err != nil {
		fmt.Println(err)
		return
	}
	enableBreakpoint(d.bps, addr)
}

func enableTBreak(d *Debugger, ops []string) {
	if len(ops) != 1 {
		fmt.Println("Usage: tdb <addr>")
		return
	}
	addr, err := parseAddr(ops[0])
	if err != nil {
		fmt.Println(err)
		return
	}
	enableBreakpoint(d.tbps, addr)
}

func enableBreakpoint(breaks map[uint16]bool, addr uint16) {
	breaks[addr] = true
}

func removeBreak(d *Debugger, ops []string) {
	if len(ops) != 1 {
		fmt.Println("Usage: rb <addr>")
		return
	}
	addr, err := parseAddr(ops[0])
	if err != nil {
		fmt.Println(err)
		return
	}
	removeBreakpoint(d.bps, addr)
}

func removeTBreak(d *Debugger, ops []string) {
	if len(ops) != 1 {
		fmt.Println("Usage: trb <addr>")
		return
	}
	addr, err := parseAddr(ops[0])
	if err != nil {
		fmt.Println(err)
		return
	}
	removeBreakpoint(d.bps, addr)
}

func removeBreakpoint(breaks map[uint16]bool, addr uint16) {
	delete(breaks, addr)
}

func run(d *Debugger, ops []string) {
	fmt.Println("Running")
	stop = false
	stopped = false
	first := true // To allow us to run while on a bp
	for stop == false {
		select {
		case <-tick:
			if v, ok := d.bps[d.c.pc]; ok && v && !first {
				fmt.Printf(red("Hit breakpoint 0x%04X\n"), d.c.pc)
				stop = true
				continue
			}
			if v, ok := d.tbps[d.c.pc]; ok && v && !first {
				fmt.Printf(red("Hit temp breakpoint 0x%04X\n"), d.c.pc)
				removeBreakpoint(d.tbps, d.c.pc)
				stop = true
				continue
			}
			first = false
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
}

func step(d *Debugger, ops []string) {
	err := d.c.RunOne()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	stop = true
}

func examine(d *Debugger, ops []string) {
	var count uint16
	if len(ops) == 0 {
		fmt.Println("usage: x [COUNT] ADDR")
		return
	}
	if len(ops) == 1 {
		count = 1
	} else {
		var err error
		count, err = parseAddr(ops[0])
		if err != nil {
			fmt.Println(err)
			return
		}
		ops = ops[1:]
	}
	addr, err := parseAddr(ops[0])
	if err != nil {
		fmt.Println(err)
		return
	}
	var i uint16
	if addr+count > MAX_MEM_ADDRESS {
		count = MAX_MEM_ADDRESS - addr
	}
	for ; count > 16; count -= 16 {
		fmt.Printf(white("%#04x: ")+"% x\n", addr+i, d.c.mem[addr+i:addr+i+16])
		i += 16
	}
	if count != 0 {
		fmt.Printf(white("%#04x: ")+"% x\n", addr+i, d.c.mem[addr+i:addr+i+count])
	}
}

func edit(d *Debugger, ops []string) {
	if len(ops) != 2 {
		fmt.Println("usage: e ADDR value")
		return
	}
	addr, err := parseAddr(ops[0])
	if err != nil {
		fmt.Println(err)
		return
	}
	v, err := strconv.ParseUint(ops[1], 0, 8)
	if err != nil {
		fmt.Println(err)
		return
	}
	d.c.mem[addr] = byte(v)
}

func quit(d *Debugger, ops []string) {
	fmt.Println("goodbye.")
	os.Exit(0)
}
