# CHIP8

My little implementation of [CHIP8](https://en.wikipedia.org/wiki/CHIP-8)

`go run github.com/Grazfather/chip8 <ROM>`

`chip8` package is provided with simple termbox-based keypad and display that
can be replaced with your own implementation.

_main.go_ will run the specified rom using a termbox-based ui, while _debug.go_ will open a simple debug repl.

![debug demo](static/demo.svg)
