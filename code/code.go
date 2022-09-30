package code

import (
	"encoding/binary"
	"fmt"
)

const (
	OpConstant Opcode = iota
)

type Instructions []byte

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

func Make(op Opcode, operands ...int) []byte {
	def, ok := definitions[op]
	if !ok {
		return []byte{}
	}

	instructionsLen := 1
	for _, w := range def.OperandWiths {
		instructionsLen += w
	}
	instructions := make([]byte, instructionsLen)
	instructions[0] = byte(op)
	offset := 1

	for i, o := range operands {
		width := def.OperandWiths[i]
		switch width {
		case 2:
			binary.BigEndian.PutUint16(instructions[offset:], uint16(o))
		}
		offset += width
	}

	return instructions
}

func (instructions Instructions) String() string {
	// var out = bytes.Buffer{}
	// offset := 0
	// for j, _ := range instructions {
	// 	fmt.Print(j)
	// 	def, _ := definitions[Opcode(instructions[offset])]
	// 	out.WriteString(def.Name)
	// 	for i, _ := range def.OperandWiths {
	// 		width := def.OperandWiths[i]
	// 		switch width {
	// 		case 2:
	// 			out.WriteString(" " + fmt.Sprint((binary.BigEndian.Uint16(instructions[offset : offset+width]))))
	// 		}
	// 		offset += width
	// 	}
	// }
	// return out.String()
	return ""
}

func ReadOperands(def *Definition, ins Instructions) ([]int, int) {
	offset := 0
	result := make([]int, len(def.OperandWiths))
	for i, width := range def.OperandWiths {
		operand := binary.BigEndian.Uint16(ins[offset:])
		switch width {
		case 2:
			result[i] = int(operand)
		}
		offset += width
	}
	return result, offset
}
