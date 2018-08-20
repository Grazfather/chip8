package chip8

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"os"
	"time"
)

const (
	V0 = iota
	V1
	V2
	V3
	V4
	V5
	V6
	V7
	V8
	V9
	V10
	V11
	V12
	V13
	V14
	V15
	VF = V15
)

const (
	ErrIllegal = "Illegal Instruction! %04X"
)

func IllegalInstruction(opcode uint16) error {
	return fmt.Errorf(ErrIllegal, opcode)
}

const MAX_MEM_ADDRESS = 0x1000

type Chip8 struct {
	mem    [0x1000]byte
	v      [16]byte
	stack  [24]uint16
	sp     int
	i      uint16
	pc     uint16
	delay  uint8
	sound  uint8
	screen IterableImage
	Renderer
	keypad   Keypad
	Renderch chan bool
	timer    *time.Ticker
	r        *rand.Rand
}

// TODO Implement incrementing I, PC behaviour (halt on overflow, or wrap
// around?)

func NewChip8(r Renderer, k Keypad) *Chip8 {
	return &Chip8{
		screen:   &myScreen{},
		Renderer: r,
		keypad:   k,
		Renderch: make(chan bool, 10),
		timer:    time.NewTicker(17 * time.Millisecond),
		r:        rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (c *Chip8) Render() {
	c.Renderer.Render(c.screen)
}

func (c *Chip8) String() string {
	return fmt.Sprintf("PC:0x%04X I:0x%04X regs:% X", c.pc, c.i, c.v)
}

func (c *Chip8) LoadBinary(filename string) (err error) {
	f, err := os.Open(filename)
	if err != nil {
		return
	}
	defer f.Close()

	n, err := f.Read(c.mem[0x200:])
	_ = n
	if err != nil {
		return
	}
	return
}

func (c *Chip8) Reset() {
	for i := 0; i < len(c.v); i++ {
		c.v[i] = 0
	}
	for i := 0; i < len(c.mem); i++ {
		c.mem[i] = 0
	}

	c.pc = 0x200
	c.i = 0
	c.sp = 0

	c.screen = &myScreen{}
	c.screen.OnEachPixel(ClearPixel)
	copy(c.mem[:], font)
}

func (c *Chip8) RunOne() error {

	// TODO: Rate limit with a timer (select on it and a stop chan)
	ins := binary.BigEndian.Uint16(c.mem[c.pc:])
	switch (ins & 0xF000) >> 12 {
	case 0x0:
		switch ins {
		case 0x00E0:
			c.Opcode00E0(ins)
		case 0x00EE:
			c.Opcode00EE(ins)
		default:
			c.Opcode0NNN(ins)
		}
	case 0x1:
		c.Opcode1NNN(ins)
	case 0x2:
		c.Opcode2NNN(ins)
	case 0x3:
		c.Opcode3XNN(ins)
	case 0x4:
		c.Opcode4XNN(ins)
	case 0x5:
		c.Opcode5XY0(ins)
	case 0x6:
		c.Opcode6XNN(ins)
	case 0x7:
		c.Opcode7XNN(ins)
	case 0x8:
		switch ins & 0xF {
		case 0x0:
			c.Opcode8XY0(ins)
		case 0x1:
			c.Opcode8XY1(ins)
		case 0x2:
			c.Opcode8XY2(ins)
		case 0x3:
			c.Opcode8XY3(ins)
		case 0x4:
			c.Opcode8XY4(ins)
		case 0x5:
			c.Opcode8XY5(ins)
		case 0x6:
			c.Opcode8XY6(ins)
		case 0x7:
			c.Opcode8XY7(ins)
		case 0xE:
			c.Opcode8XYE(ins)
		default:
			return IllegalInstruction(ins)
		}
	case 0x9:
		if ins&0xF != 0 {
			return IllegalInstruction(ins)
		}
		c.Opcode9XY0(ins)
	case 0xA:
		c.OpcodeANNN(ins)
	case 0xB:
		c.OpcodeBNNN(ins)
	case 0xC:
		c.OpcodeCXNN(ins)
	case 0xD:
		c.OpcodeDXYN(ins)
	case 0xE:
		switch ins & 0xFF {
		case 0x9E:
			c.OpcodeEX9E(ins)
		case 0xA1:
			c.OpcodeEXA1(ins)
		default:
			return IllegalInstruction(ins)
		}
	case 0xF:
		switch ins & 0xFF {
		case 0x07:
			c.OpcodeFX07(ins)
		case 0x0A:
			c.OpcodeFX0A(ins)
		case 0x15:
			c.OpcodeFX15(ins)
		case 0x18:
			c.OpcodeFX18(ins)
		case 0x1E:
			c.OpcodeFX1E(ins)
		case 0x29:
			c.OpcodeFX29(ins)
		case 0x33:
			c.OpcodeFX33(ins)
		case 0x55:
			c.OpcodeFX55(ins)
		case 0x65:
			c.OpcodeFX65(ins)
		default:
			return IllegalInstruction(ins)
		}
	default:
		return IllegalInstruction(ins)
	}
	c.pc += 2

	return nil
}

func (c *Chip8) KeepTime() {
	for {
		select {
		case <-c.timer.C:
			if c.delay != 0 {
				c.delay--
			}
			if c.sound != 0 {
				c.sound--
				if c.sound == 0 {
					fmt.Printf("\a")
				}
			}
		}
	}
}
