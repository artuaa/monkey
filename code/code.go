package code

import (
	"encoding/binary"
	"fmt"
)

const (
	OpConstant Opcode = iota
)

type Instruction []byte

type Opcode byte

type Definition struct {
	Name         string
	OperandWiths []int
}

var definitions = map[Opcode]*Definition{
	OpConstant: {"OpConstant", []int{2}},
}

func Lookup(op byte) (*Definition, error) {
	def, ok := definitions[Opcode(op)]
	if !ok {
		return nil, fmt.Errorf("opcode %d undefined", op)
	}
	return def, nil
}

func Make(op Opcode, operands... int) []byte {
    def, ok := definitions[op]
	if !ok{
		return []byte{}
	}

	instructionsLen := 1
	for _, w :=range def.OperandWiths{
		instructionsLen +=w
	}
	instructions := make([]byte, instructionsLen)
	instructions[0] = byte(op)
    offset := 1

	for i, o := range operands {
		width := def.OperandWiths[i]
		switch width{
			case 2:
			binary.BigEndian.PutUint16(instructions[offset:], uint16(o))
		}
		offset += width
	}

	return instructions
}
