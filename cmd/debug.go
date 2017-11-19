package main

import (
	"fmt"
	"os"
	"time"

	"github.com/Grazfather/chip8"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("usage: chip8 <filename>")
		os.Exit(1)
	}

	k := &chip8.NoKeypad{}
	d := &chip8.NullDisplay{}
	c := chip8.NewChip8(d, k)
	c.Reset()

	if err := c.LoadBinary(os.Args[1]); err != nil {
		fmt.Printf("Error loading %s: %v\n", os.Args[1], err)
		os.Exit(1)
	}

	go c.KeepTime()

	debugger := chip8.NewDebugger(c)
	debugger.Start()

	time.Sleep(1 * time.Second)
}
