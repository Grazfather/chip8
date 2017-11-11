package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/nsf/termbox-go"

	"github.com/Grazfather/chip8"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("usage: chip8 <filename>")
		os.Exit(1)
	}

	d, err := chip8.NewTerminal()
	if err != nil {
		fmt.Printf("Error setting up renderer: %v\n", err)
		os.Exit(1)
	}
	defer termbox.Close()

	quit := make(chan termbox.Event)
	k := chip8.NewTermKeypad(quit)

	c := chip8.NewChip8(d, k)
	c.Reset()

	if err := c.LoadBinary(os.Args[1]); err != nil {
		fmt.Printf("Error loading %s: %v\n", os.Args[1], err)
		os.Exit(1)
	}

	go c.KeepTime()

	tick := time.Tick(2 * time.Millisecond)
	exit := make(chan os.Signal)
	signal.Notify(exit, os.Interrupt)

LOOP:
	for {
		select {
		case <-tick:
			err := c.RunOne()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				break LOOP
			}
		case <-exit:
			break LOOP
		case <-quit:
			break LOOP
		}
	}
	time.Sleep(1 * time.Second)
}
