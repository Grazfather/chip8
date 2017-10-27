package chip8

import (
	"math/rand"
)

// Opcode0NNN is an ignored opcode
func (c *Chip8) Opcode0NNN(ins uint16) {
	panic("unimplemented")
}

// Opcode00E0 clears the screen
func (c *Chip8) Opcode00E0(ins uint16) {
	c.screen.OnEachPixel(ClearPixel)
}

func (c *Chip8) Opcode00EE(ins uint16) {
	// TODO: The RA is actually the call address, which will get incremented
	c.pc = c.stack[c.sp]
	c.sp -= 2
}

// Opcode1NNN jumps to address NNN
func (c *Chip8) Opcode1NNN(ins uint16) {
	// TODO: PC target again. Compensating here with -2
	c.pc = ArgNNN(ins) - 2
}

// Opcode2NNN calls the subroutne at NNN
func (c *Chip8) Opcode2NNN(ins uint16) {
	// TODO: set PC target. For now we count on the loop incrementing past the call instruction
	c.sp += 2
	c.stack[c.sp] = c.pc
	c.pc = ArgNNN(ins) - 2
}

// Opcode3XNN Skips the next instruction if Vx equals NN
func (c *Chip8) Opcode3XNN(ins uint16) {
	if c.v[ArgX(ins)] == ArgNN(ins) {
		c.pc += 2
	}
}

func (c *Chip8) Opcode4XNN(ins uint16) {
	if c.v[ArgX(ins)] != ArgNN(ins) {
		c.pc += 2
	}
}

func (c *Chip8) Opcode5XY0(ins uint16) {
	if c.v[ArgX(ins)] == c.v[ArgY(ins)] {
		c.pc += 2
	}
}

func (c *Chip8) Opcode6XNN(ins uint16) {
	c.v[ArgX(ins)] = ArgNN(ins)
}

func (c *Chip8) Opcode7XNN(ins uint16) {
	c.v[ArgX(ins)] += ArgNN(ins)
}

func (c *Chip8) Opcode8XY0(ins uint16) {
	c.v[ArgX(ins)] = c.v[ArgY(ins)]
}

func (c *Chip8) Opcode8XY1(ins uint16) {
	c.v[ArgX(ins)] |= c.v[ArgY(ins)]
}

func (c *Chip8) Opcode8XY2(ins uint16) {
	c.v[ArgX(ins)] &= c.v[ArgY(ins)]
}

func (c *Chip8) Opcode8XY3(ins uint16) {
	c.v[ArgX(ins)] ^= c.v[ArgY(ins)]
}

func (c *Chip8) Opcode8XY4(ins uint16) {
	old := c.v[ArgX(ins)]
	c.v[ArgX(ins)] += c.v[ArgY(ins)]
	// Carry
	if old > c.v[ArgX(ins)] {
		c.v[VF] = 1
	} else {
		c.v[VF] = 0
	}
}

func (c *Chip8) Opcode8XY5(ins uint16) {
	old := c.v[ArgX(ins)]
	c.v[ArgX(ins)] -= c.v[ArgY(ins)]
	// Borrow
	if old < c.v[ArgX(ins)] {
		c.v[VF] = 0
	} else {
		c.v[VF] = 1
	}
}

func (c *Chip8) Opcode8XY6(ins uint16) {
	lsb := c.v[ArgY(ins)] & 1
	v := c.v[ArgY(ins)] >> 1
	c.v[ArgX(ins)] = v
	c.v[ArgY(ins)] = v
	c.v[VF] = lsb
}

func (c *Chip8) Opcode8XY7(ins uint16) {
	y := c.v[ArgY(ins)]
	x := c.v[ArgX(ins)]
	c.v[ArgX(ins)] = x - y
	// Borrow
	if y > x {
		c.v[VF] = 0
	} else {
		c.v[VF] = 1
	}
}

func (c *Chip8) Opcode8XYE(ins uint16) {
	msg := c.v[ArgY(ins)] >> 7 & 1
	v := c.v[ArgY(ins)] << 1
	c.v[ArgX(ins)] = v
	c.v[ArgY(ins)] = v
	c.v[VF] = msg
}

func (c *Chip8) Opcode9XY0(ins uint16) {
	if c.v[ArgX(ins)] != c.v[ArgY(ins)] {
		c.pc += 2
	}
}

func (c *Chip8) OpcodeANNN(ins uint16) {
	c.i = ArgNNN(ins)
}

func (c *Chip8) OpcodeBNNN(ins uint16) {
	// TODO: Want to skip the +=2 at the end of the loop
	c.pc = (uint16(c.v[V0]) + ArgNNN(ins) - 2) & 0xFFF
}

// OpcodeCXNN sets Vx to the result of rand()&NN
func (c *Chip8) OpcodeCXNN(ins uint16) {
	c.v[ArgX(ins)] = uint8(rand.Uint32()) & ArgNN(ins)
}

// OpcodeDXYN draws a sprite I to Vx, Vy with width 8 height N
func (c *Chip8) OpcodeDXYN(ins uint16) {
	x := c.v[ArgX(ins)]
	y := c.v[ArgY(ins)]
	height := ArgN(ins)
	collision := false
	for j := uint8(0); j < height; j++ {
		row := c.mem[c.i+uint16(j)]
		for i := uint8(0); i < 8; i++ {
			color := byte(0)
			if (row & 0x80) == 0x80 {
				color = 1
			}
			if c.screen.Toggle(int(x+i), int(y+j), color) {
				collision = true
			}
			row <<= 1
		}
	}
	if collision {
		c.v[VF] = 1
	} else {
		c.v[VF] = 0
	}
	c.Render(c.screen)
}

// OpcodeEX9E skips the next instruction if key Vx is pressed
func (c *Chip8) OpcodeEX9E(ins uint16) {
	if c.keypad.Pressed(c.v[ArgX(ins)]) {
		c.pc += 2
	}
}

// OpcodeEXA1 skips the next instruction if key Vx is not pressed
func (c *Chip8) OpcodeEXA1(ins uint16) {
	if !c.keypad.Pressed(c.v[ArgX(ins)]) {
		c.pc += 2
	}
}

// OpcodeFX07 sets Vx to the value of the delay timer
func (c *Chip8) OpcodeFX07(ins uint16) {
	c.v[ArgX(ins)] = c.delay
}

// OpcodeFX0A halts until a key is pressed, and stores it in Vx
func (c *Chip8) OpcodeFX0A(ins uint16) {
	c.v[ArgX(ins)] = c.keypad.WaitPress()
}

// OpcodeFX15 sets the delay timer to Vx
func (c *Chip8) OpcodeFX15(ins uint16) {
	c.delay = c.v[ArgX(ins)]
}

// OpcodeFX18 sets the sound timer to Vx
func (c *Chip8) OpcodeFX18(ins uint16) {
	c.sound = c.v[ArgX(ins)]
}

// OpcodeFX1E adds Vx to I
func (c *Chip8) OpcodeFX1E(ins uint16) {
	c.i = (c.i + uint16(c.v[ArgX(ins)])) & 0xFFF
}

// OpcodeFX29 sets I to point to sprite for digit Vx
func (c *Chip8) OpcodeFX29(ins uint16) {
	c.i = 5 * uint16(c.v[ArgX(ins)])
}

// OpcodeFX33 stores the BCD representation of Vx into memory at I
func (c *Chip8) OpcodeFX33(ins uint16) {
	v := c.v[ArgX(ins)]
	c.mem[c.i+2] = v % 10
	v /= 10
	c.mem[c.i+1] = v % 10
	v /= 10
	c.mem[c.i] = v % 10
}

// OpcodeFX55 Stores V[0-X] inclusive in memory starting at address I
func (c *Chip8) OpcodeFX55(ins uint16) {
	x := ArgX(ins)
	for r := uint8(0); r < x; r++ {
		c.mem[c.i] = c.v[r]
		c.i++
	}
}

// OpcodeFX65 Loads V[0-X] inclusive from memory starting at address I
func (c *Chip8) OpcodeFX65(ins uint16) {
	x := ArgX(ins)
	for r := uint8(0); r < x; r++ {
		c.v[r] = c.mem[c.i]
		c.i++
	}
}

// TODO: Return a reference we can write to
func ArgX(ins uint16) uint8 {
	return uint8(ins>>8) & 0xF
}

func ArgY(ins uint16) uint8 {
	return uint8(ins>>4) & 0xF
}

func ArgN(ins uint16) uint8 {
	return uint8(ins) & 0xF
}

func ArgNN(ins uint16) uint8 {
	return uint8(ins) & 0xFF
}

func ArgNNN(ins uint16) uint16 {
	return ins & 0xFFF
}
