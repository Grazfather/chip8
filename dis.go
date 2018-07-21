package chip8

import (
	"encoding/binary"
	"fmt"
)

type instruction struct {
	name string
	args []string
	op   uint16
}

func (i instruction) String() string {
	// Go is retarded https://golang.org/doc/faq#convert_slice_of_interface
	// this would be too nice: return fmt.Sprintf(i.name, i.args...)
	s := make([]interface{}, len(i.args))
	for j, v := range i.args {
		s[j] = v
	}
	return fmt.Sprintf(i.name, s...)
}

func (i instruction) isCall() bool {
	return (i.op & 0xF000) == 0x2000
}

func (i instruction) callTarget() uint16 {
	return ArgNNN(i.op) - 2
}

type Disassembler struct{}

func (d *Disassembler) dis(mem []byte) instruction {
	ins := binary.BigEndian.Uint16(mem[:])

	switch (ins & 0xF000) >> 12 {
	case 0x0:
		switch ins {
		case 0x00E0:
			return instruction{"CLS", nil, ins}
		case 0x00EE:
			return instruction{"RET", nil, ins}
		default:
			return instruction{"SYS %s", []string{SArgNNN(ins)}, ins}
		}
	case 0x1:
		return instruction{"JP %s", []string{SArgNNN(ins)}, ins}
	case 0x2:
		return instruction{"CALL %s", []string{SArgNNN(ins)}, ins}
	case 0x3:
		return instruction{"SE %s, %s", []string{SArgX(ins), SArgNN(ins)}, ins}
	case 0x4:
		return instruction{"SNE %s, %s", []string{SArgX(ins), SArgNN(ins)}, ins}
	case 0x5:
		return instruction{"SE %s, %s", []string{SArgX(ins), SArgY(ins)}, ins}
	case 0x6:
		return instruction{"LD %s, %s", []string{SArgX(ins), SArgNN(ins)}, ins}
	case 0x7:
		return instruction{"ADD %s, %s", []string{SArgX(ins), SArgNN(ins)}, ins}
	case 0x8:
		switch ins & 0xF {
		case 0x0:
			return instruction{"LD %s, %s", []string{SArgX(ins), SArgY(ins)}, ins}
		case 0x1:
			return instruction{"OR %s, %s", []string{SArgX(ins), SArgY(ins)}, ins}
		case 0x2:
			return instruction{"AND %s, %s", []string{SArgX(ins), SArgY(ins)}, ins}
		case 0x3:
			return instruction{"XOR %s, %s", []string{SArgX(ins), SArgY(ins)}, ins}
		case 0x4:
			return instruction{"ADD %s, %s", []string{SArgX(ins), SArgY(ins)}, ins}
		case 0x5:
			return instruction{"SUB %s, %s", []string{SArgX(ins), SArgY(ins)}, ins}
		case 0x6:
			return instruction{"SHR %s, %s", []string{SArgX(ins), SArgY(ins)}, ins}
		case 0x7:
			return instruction{"SUBN %s %s", []string{SArgX(ins), SArgY(ins)}, ins}
		case 0xE:
			return instruction{"SHL %s, %s", []string{SArgX(ins), SArgY(ins)}, ins}
		default:
			return instruction{"<ILL>", nil, ins}
		}
	case 0x9:
		if ins&0xF != 0 {
			return instruction{"<ILL>", nil, ins}
		}
		return instruction{"SNE %s, %s", []string{SArgX(ins), SArgY(ins)}, ins}
	case 0xA:
		return instruction{"LD I, %s", []string{SArgNNN(ins)}, ins}
	case 0xB:
		return instruction{"JP V0, ADDR", []string{SArgNNN(ins)}, ins}
	case 0xC:
		return instruction{"RND, %s, %s", []string{SArgX(ins), SArgNN(ins)}, ins}
	case 0xD:
		return instruction{"DRW %s, %s, %s", []string{SArgX(ins), SArgY(ins), SArgN(ins)}, ins}
	case 0xE:
		switch ins & 0xFF {
		case 0x9E:
			return instruction{"SKP %s", []string{SArgX(ins)}, ins}
		case 0xA1:
			return instruction{"SKNP %s", []string{SArgX(ins)}, ins}
		default:
			return instruction{"<ILL>", nil, ins}
		}
	case 0xF:
		switch ins & 0xFF {
		case 0x07:
			return instruction{"LD %s, DT", []string{SArgX(ins)}, ins}
		case 0x0A:
			return instruction{"LD %s, K", []string{SArgX(ins)}, ins}
		case 0x15:
			return instruction{"LD DT, %s", []string{SArgX(ins)}, ins}
		case 0x18:
			return instruction{"LD ST, %s", []string{SArgX(ins)}, ins}
		case 0x1E:
			return instruction{"ADD I, %s", []string{SArgX(ins)}, ins}
		case 0x29:
			return instruction{"LD F, %s", []string{SArgX(ins)}, ins}
		case 0x33:
			return instruction{"LD B, %s", []string{SArgX(ins)}, ins}
		case 0x55:
			return instruction{"LD [I], %s", []string{SArgX(ins)}, ins}
		case 0x65:
			return instruction{"LD %s, [I]", []string{SArgX(ins)}, ins}
		default:
			return instruction{"<ILL>", nil, ins}
		}
	default:
		return instruction{"<ILL>", nil, ins}
	}
}

func SArgX(ins uint16) string {
	return fmt.Sprintf("V%01X", uint8(ins>>8)&0xF)
}

func SArgY(ins uint16) string {
	return fmt.Sprintf("V%01X", uint8(ins>>4)&0xF)
}

func SArgN(ins uint16) string {
	return fmt.Sprintf("0x%01X", uint8(ins)&0xF)
}

func SArgNN(ins uint16) string {
	return fmt.Sprintf("0x%02X", uint8(ins)&0xFF)
}

func SArgNNN(ins uint16) string {
	return fmt.Sprintf("0x%03X", ins&0xFFF)
}
