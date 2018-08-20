package main

import (
	"fmt"
	"os"

	"github.com/Grazfather/chip8"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("usage: chip8 <filename>")
		os.Exit(1)
	}

	debugger := chip8.NewDebugger()
	debugger.Start(os.Args[1])
}
