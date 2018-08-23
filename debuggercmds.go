package chip8

import (
	"os"
	"strconv"
)

func reset(d *Debugger, ops []string) {
	d.Println("Reseting CPU")
	d.c.Reset()
	if err := d.c.LoadBinary(d.rom); err != nil {
		panic("Cannot reload rom")
	}
	d.stopped = false // Redraws context
}

func context(d *Debugger, ops []string) {
	// Setting stop makes the debugger show the context
	d.stop = true
	d.stopped = false // Redraws context
}

func (d *Debugger) printBreakpoints(breaks map[uint16]Breakpoint) {
	// TODO: Sort in any way?
	for a, v := range breaks {
		if v.enabled {
			d.Printf(green("+ 0x%04X ")+"(hit %d times)\n", a, v.timesHit)
		} else {
			d.Printf("- 0x%04X (hit %d times)\n", a, v.timesHit)
		}
	}
}

func breakpoints(d *Debugger, ops []string) {
	if (len(d.bps) == 0) && (len(d.tbps) == 0) {
		d.Println(white("No breakpoints"))
		return
	}
	if len(d.bps) > 0 {
		d.Println(white("Breakpoints"))
		d.printBreakpoints(d.bps)
	}
	if len(d.tbps) > 0 {
		d.Println(white("Temp Breakpoints"))
		d.printBreakpoints(d.tbps)
	}
}

func addBreak(d *Debugger, ops []string) {
	if len(ops) != 1 {
		d.Println("Usage: b <addr>")
		return
	}
	addr, err := parseAddr(ops[0])
	if err != nil {
		d.Println(err)
		return
	}
	addBreakpoint(d.bps, addr)
	d.Printf("Added bp at 0x%04X\n", addr)
}

func addTBreak(d *Debugger, ops []string) {
	if len(ops) != 1 {
		d.Println("Usage: tb <addr>")
		return
	}
	addr, err := parseAddr(ops[0])
	if err != nil {
		d.Println(err)
		return
	}
	addBreakpoint(d.tbps, addr)
	d.Printf("Added bp at 0x%04X\n", addr)
}

func addBreakpoint(breaks map[uint16]Breakpoint, addr uint16) {
	breaks[addr] = Breakpoint{true, 0}
}

func disableBreak(d *Debugger, ops []string) {
	if len(ops) != 1 {
		d.Println("Usage: db <addr>")
		return
	}
	addr, err := parseAddr(ops[0])
	if err != nil {
		d.Println(err)
		return
	}
	disableBreakpoint(d.bps, addr)
	d.Printf("Disabled bp at 0x%04X\n", addr)
}

func disableTBreak(d *Debugger, ops []string) {
	if len(ops) != 1 {
		d.Println("Usage: dtb <addr>")
		return
	}
	addr, err := parseAddr(ops[0])
	if err != nil {
		d.Println(err)
		return
	}
	disableBreakpoint(d.tbps, addr)
	d.Printf("Disabled bp at 0x%04X\n", addr)
}

func disableBreakpoint(breaks map[uint16]Breakpoint, addr uint16) {
	breaks[addr] = Breakpoint{false, breaks[addr].timesHit}
}

func enableBreak(d *Debugger, ops []string) {
	if len(ops) != 1 {
		d.Println("Usage: db <addr>")
		return
	}
	addr, err := parseAddr(ops[0])
	if err != nil {
		d.Println(err)
		return
	}
	enableBreakpoint(d.bps, addr)
	d.Printf("Enabled bp at 0x%04X\n", addr)
}

func enableTBreak(d *Debugger, ops []string) {
	if len(ops) != 1 {
		d.Println("Usage: tdb <addr>")
		return
	}
	addr, err := parseAddr(ops[0])
	if err != nil {
		d.Println(err)
		return
	}
	enableBreakpoint(d.tbps, addr)
	d.Printf("Enabled bp at 0x%04X\n", addr)
}

func enableBreakpoint(breaks map[uint16]Breakpoint, addr uint16) {
	breaks[addr] = Breakpoint{true, breaks[addr].timesHit}
}

func removeBreak(d *Debugger, ops []string) {
	if len(ops) != 1 {
		d.Println("Usage: rb <addr>")
		return
	}
	addr, err := parseAddr(ops[0])
	if err != nil {
		d.Println(err)
		return
	}
	removeBreakpoint(d.bps, addr)
	d.Printf("Removed bp at 0x%04X\n", addr)
}

func removeTBreak(d *Debugger, ops []string) {
	if len(ops) != 1 {
		d.Println("Usage: trb <addr>")
		return
	}
	addr, err := parseAddr(ops[0])
	if err != nil {
		d.Println(err)
		return
	}
	removeBreakpoint(d.bps, addr)
	d.Printf("Removed bp at 0x%04X\n", addr)
}

func removeBreakpoint(breaks map[uint16]Breakpoint, addr uint16) {
	delete(breaks, addr)
}

func cont(d *Debugger, ops []string) {
	d.cont()
}

func step(d *Debugger, ops []string) {
	d.StepOne()
}

func next(d *Debugger, ops []string) {
	// Next is just like step, except for if we're on a call instruction,
	// we stop after the call finishes.
	if ins := d.dis.dis(d.c.mem[d.c.pc : d.c.pc+2]); ins.isCall() {
		addBreakpoint(d.tbps, d.c.pc+2)
		cont(d, nil)
	} else {
		d.StepOne()
	}
}

func examine(d *Debugger, ops []string) {
	var count uint16
	if len(ops) == 0 {
		d.Println("usage: x [COUNT] ADDR")
		return
	}
	if len(ops) == 1 {
		count = 1
	} else {
		var err error
		count, err = parseAddr(ops[0])
		if err != nil {
			d.Println(err)
			return
		}
		ops = ops[1:]
	}
	addr, err := parseAddr(ops[0])
	if err != nil {
		d.Println(err)
		return
	}
	var i uint16
	if addr+count > MAX_MEM_ADDRESS {
		count = MAX_MEM_ADDRESS - addr
	}
	for ; count > 16; count -= 16 {
		d.Printf(white("%#04x: ")+"% x\n", addr+i, d.c.mem[addr+i:addr+i+16])
		i += 16
	}
	if count != 0 {
		d.Printf(white("%#04x: ")+"% x\n", addr+i, d.c.mem[addr+i:addr+i+count])
	}
}

func edit(d *Debugger, ops []string) {
	if len(ops) != 2 {
		d.Println("usage: e ADDR value")
		return
	}
	addr, err := parseAddr(ops[0])
	if err != nil {
		d.Println(err)
		return
	}
	v, err := strconv.ParseUint(ops[1], 0, 8)
	if err != nil {
		d.Println(err)
		return
	}
	d.c.mem[addr] = byte(v)
}

func quit(d *Debugger, ops []string) {
	d.Println("goodbye.")
	os.Exit(0)
}
