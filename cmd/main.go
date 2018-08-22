package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jroimartin/gocui"

	"github.com/Grazfather/chip8"
)

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	var err error
	if maxY < chip8.SCREEN_HEIGHT || maxX < chip8.SCREEN_WIDTH {
		return fmt.Errorf("Cannot display if less than %d x %d! Resize your terminal! (^Q to quit)",
			chip8.SCREEN_WIDTH, chip8.SCREEN_HEIGHT)
	}
	left := (maxX - chip8.SCREEN_WIDTH) / 2
	_, err = g.SetView("display", left, 0, chip8.SCREEN_WIDTH+2+left, chip8.SCREEN_HEIGHT+2)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	}
	return nil
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("usage: chip8 <filename>")
		os.Exit(1)
	}

	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.SetManagerFunc(layout)
	// HACK: Need to call layout once to create the views
	layout(g)
	v, err := g.View("display")
	if err != nil {
		log.Panicln(err)
	}

	g.SetCurrentView(v.Name())
	k := chip8.NewGocuiKeypad(g, v)
	r := chip8.NewGocuiRenderer(v)
	c := chip8.NewChip8(r, k)
	c.Reset()

	if err := c.LoadBinary(os.Args[1]); err != nil {
		fmt.Printf("Error loading %s: %v\n", os.Args[1], err)
		os.Exit(1)
	}

	if err := g.SetKeybinding("", gocui.KeyCtrlQ, gocui.ModNone,
		func(g *gocui.Gui, v *gocui.View) error { return gocui.ErrQuit }); err != nil {
		log.Panicln(err)
	}

	go func() {
		go c.KeepTime()

		tick := time.Tick(2 * time.Millisecond)
	LOOP:
		for {
			select {
			case <-tick:
				err := c.RunOne()
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					break LOOP
				}
				if c.RenderFlag {
					g.Update(func(g *gocui.Gui) error {
						c.Render()
						return nil
					})
				}
			}
		}
	}()

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}
