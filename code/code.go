package code

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

const (
	OpConstant Opcode = iota
	OpAdd
)

type Instructions []byte

type Opcode byte

type Definition struct {
	Name         string
	OperandWiths []int
}

var definitions = map[Opcode]*Definition{
	OpConstant: {"OpConstant", []int{2}},
	OpAdd:      {"OpAdd", []int{}},
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
	var out = bytes.Buffer{}
	i := 0
	for i < len(instructions) {
		def, err := Lookup(instructions[i])
		if err != nil {
			fmt.Fprintf(&out, "ERROR: %s\n", err)
		}
		operands, read := ReadOperands(def, instructions[i+1:])
		out.WriteString(fmt.Sprintf("%04d ", i))
		switch len(operands) {
		case 0:
			out.WriteString(def.Name + "\n\t")
		case 1:
			out.WriteString(fmt.Sprintf("%s %d\n\t", def.Name, operands[0]))
		default:
			out.WriteString(fmt.Sprintf("ERROR: unhandled operandCount for %s\n", def.Name))
		}
		i += 1 + read
	}
	return out.String()
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