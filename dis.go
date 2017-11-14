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

type Disassembler struct{}

func (d *Disassembler) dis(mem []byte) instruction {
	ins := binary.BigEndian.Uint16(mem[:])

	switch (ins & 0xF000) >> 12 {
	case 0x0:
		switch ins {
		case 0x00E0:
			return instruction{"Clear display", nil, ins}
		case 0x00EE:
			return instruction{"Return", nil, ins}
		default:
			return instruction{"Call RCA 1802 %s", []string{SArgNNN(ins)}, ins}
		}
	case 0x1:
		return instruction{"Jump to %s", []string{SArgNNN(ins)}, ins}
	case 0x2:
		return instruction{"Call %s", []string{SArgNNN(ins)}, ins}
	case 0x3:
		return instruction{"Skip next if %s == %s", []string{SArgX(ins), SArgNN(ins)}, ins}
	case 0x4:
		return instruction{"Skip next if %s != %s", []string{SArgX(ins), SArgNN(ins)}, ins}
	case 0x5:
		return instruction{"Skip next if %s == %s", []string{SArgX(ins), SArgY(ins)}, ins}
	case 0x6:
		return instruction{"Set %s = %s", []string{SArgX(ins), SArgNN(ins)}, ins}
	case 0x7:
		return instruction{"Set %s += %s", []string{SArgX(ins), SArgNN(ins)}, ins}
	case 0x8:
		switch ins & 0xF {
		case 0x0:
			return instruction{"Set %s = %s", []string{SArgX(ins), SArgY(ins)}, ins}
		case 0x1:
			return instruction{"Set %s = %[1]s|%[2]s", []string{SArgX(ins), SArgY(ins)}, ins}
		case 0x2:
			return instruction{"Set %s = %[1]s&%[2]s", []string{SArgX(ins), SArgY(ins)}, ins}
		case 0x3:
			return instruction{"Set %s = %[1]s^%[2]s", []string{SArgX(ins), SArgY(ins)}, ins}
		case 0x4:
			return instruction{"Set %s += %s", []string{SArgX(ins), SArgY(ins)}, ins}
		case 0x5:
			return instruction{"Set %s -= %s", []string{SArgX(ins), SArgY(ins)}, ins}
		case 0x6:
			return instruction{"Set %s = %[2]s = %[1]s>>1", []string{SArgX(ins), SArgY(ins)}, ins}
		case 0x7:
			return instruction{"Set %s = %[2]s - %[1]s", []string{SArgX(ins), SArgY(ins)}, ins}
		default:
			return instruction{"ILLEGAL", nil, ins}
		}
	case 0x9:
		if ins&0xF != 0 {
			return instruction{"ILLEGAL", nil, ins}
		}
		return instruction{"Skip next if %s != %s", []string{SArgX(ins), SArgY(ins)}, ins}
	case 0xA:
		return instruction{"Set I = %s", []string{SArgNNN(ins)}, ins}
	case 0xB:
		return instruction{"Jump to V0 + %s", []string{SArgNNN(ins)}, ins}
	case 0xC:
		return instruction{"Set %s = rand() & %s", []string{SArgX(ins), SArgNN(ins)}, ins}
	case 0xD:
		return instruction{"Draw sprite at I at (%s, %s) height %s", []string{SArgX(ins), SArgY(ins), SArgN(ins)}, ins}
	case 0xE:
		switch ins & 0xFF {
		case 0x9E:
			return instruction{"Skip next if key() == %s", []string{SArgX(ins)}, ins}
		case 0xA1:
			return instruction{"Skip next if key() != %s", []string{SArgX(ins)}, ins}
		default:
			return instruction{"ILLEGAL", nil, ins}
		}
	case 0xF:
		switch ins & 0xFF {
		case 0x07:
			return instruction{"Set %s = get_delay()", []string{SArgX(ins)}, ins}
		case 0x0A:
			return instruction{"Set %s = get_key()", []string{SArgX(ins)}, ins}
		case 0x15:
			return instruction{"Set delay_timer = %s", []string{SArgX(ins)}, ins}
		case 0x18:
			return instruction{"Set sound_timer = %s", []string{SArgX(ins)}, ins}
		case 0x1E:
			return instruction{"Set I += %s", []string{SArgX(ins)}, ins}
		case 0x29:
			return instruction{"Set I = sprite_addr[%s]", []string{SArgX(ins)}, ins}
		case 0x33:
			return instruction{"Store BCD(%s) at I", []string{SArgX(ins)}, ins}
		case 0x55:
			return instruction{"Store V0 to %s at I", []string{SArgX(ins)}, ins}
		case 0x65:
			return instruction{"Load V0 to %s from I", []string{SArgX(ins)}, ins}
		default:
			return instruction{"ILLEGAL", nil, ins}
		}
	default:
		return instruction{"ILLEGAL", nil, ins}
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
