package code

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// Instructions represents a sequence of bytes that define a set of instructions.
type Instructions []byte

// String returns a string representation of the Instructions.
// It iterates over the Instructions and formats each instruction
// along with its operands. If an error occurs during the lookup
// or reading of operands, it appends an error message to the output.
func (ins Instructions) String() string {
	var out bytes.Buffer

	i := 0
	for i < len(ins) {
		def, err := Lookup(ins[i])
		if err != nil {
			fmt.Fprintf(&out, "ERROR: %s\n", err)
			continue
		}

		operands, read := ReadOperands(def, ins[i+1:])
		fmt.Fprintf(&out, "%04d %s\n", i, ins.fmtInstruction(def, operands))

		i += 1 + read
	}

	return out.String()
}

// fmtInstruction formats an instruction based on the given definition and operands.
// It returns the formatted instruction as a string.
func (ins Instructions) fmtInstruction(def *Definition, operands []int) string {
	operandCount := len(def.OperandWidths)

	if len(operands) != operandCount {
		return fmt.Sprintf("ERROR: operand len %d does not match defined %d\n", len(operands), operandCount)
	}

	switch operandCount {
	case 0:
		return def.Name
	case 1:
		return fmt.Sprintf("%s %d", def.Name, operands[0])
	}

	return fmt.Sprintf("ERROR: unhandled operandCount for %s\n", def.Name)
}

type Opcode byte

const (
	// OpConstant has one operand: the number we previously assigned to the constants
	// When the VM executes OpConstant it retrieves the constant using the operand
	// as an index and pushes it on to the stack
	OpConstant Opcode = iota
	// OpAdd has no operands. It adds the two numbers on top of the stack and adds
	// the result back on to the stack
	OpAdd
	// OpMultiply has no operands. It multiplies the top two numbers on the stacks
	// and adds the result to the stack
	OpMultiply
	// OpDivide has no operands. It divides the top two numbers on the stack and
	// adds the results to the stack
	OpDivide
	// OpSubtract has no operands. It divides the top two numbers on the stack and
	// adds the result to the stack
	OpSubtract
	// OpPop tells the VM when to pop the topmost element off the stack
	OpPop
	// OpTrue represents the true boolean literal
	OpTrue
	// OpFalse represents the false boolean literal
	OpFalse
	// OpEqual represents the == comparison operator
	OpEqual
	// OpNotEqual represents the != comparison operator
	OpNotEqual
	// OpGreaterThan represents the > comparison operator
	// There is no representation of < comparison operator because we can just flip < to >
	// e.g. 4 < 2 --> 2 > 4
	OpGreaterThan
	// OpMinus represents the - negate operator. Negating the integer thats
	// on the top of the stack
	OpMinus
	// OpBang represents the ! negate operator. Negating the boolean thats
	// on top of the stack
	OpBang
	// OpJumpNotTruthy is used to jump when a given condition resolves to a non-truthy value
	// This opcode expects an argument to where to jump to if the condition is falsy
	OpJumpNotTruthy
	// OpJump is used when jumping out of a conditional body when it results in a truthy value.
	// This opcode expects an argument to where to jump to
	OpJump
	// OpNull represents a Null value and tells the vm to insert a null value
	OpNull
	// OpGetGlobal represents retrieving the value of a variable
	OpGetGlobal
	// OpSetGloabl represents setting the value of a global variable
	OpSetGlobal
	// OpArray represents an array. It tells the VM how to build the array.
	// It has 1 argument, an integer that represents how big the array is
	OpArray
	// OpHash represents a hash table in orionlang. It tells the VM how to build the hash table.
	// It has 1 argument, an integer that represents the number of keys and values sitting on the stack
	OpHash
	// OpIndex represents indexing an array, hash, etc. There need to be two values sitting on the top
	// of the stack: the object to be indexed and above that, the object serving as the index
	OpIndex
	// OpCall represents a call to a function. This tells the VM to execute the function on top
	// of the stack
	OpCall
	// OpReturnValue represents returning from a function and tells the VM to return a value
	OpReturnValue
	// OpReturn represents returning from a function with no value. The VM returns null.
	OpReturn
	// OpGetLocal represents retrieving the value of a local variable
	OpGetLocal
	// OpSetLocal represents setting the value of a local variable
	OpSetLocal
	// OpGetBuiltin represents retrieving a builtin function to call
	OpGetBuiltin
)

type Definition struct {
	// Name helps to make an Opcode readable
	Name string
	// OperandWidths contains the number of bytes each operand takes up
	OperandWidths []int
}

// definitions holds all opcode definitions
var definitions = map[Opcode]*Definition{
	// OpConstant has only an operand that is two bytes wide, which makes it an uint16
	// which limits its maximum value to 65536
	OpConstant:      {"OpConstant", []int{2}},
	OpAdd:           {"OpAdd", []int{}},
	OpPop:           {"OpPop", []int{}},
	OpMultiply:      {"OpMultiply", []int{}},
	OpDivide:        {"OpDivide", []int{}},
	OpSubtract:      {"OpSubtract", []int{}},
	OpFalse:         {"OpFalse", []int{}},
	OpTrue:          {"OpTrue", []int{}},
	OpEqual:         {"OpEqual", []int{}},
	OpNotEqual:      {"OpNotEqual", []int{}},
	OpGreaterThan:   {"OpGreaterThan", []int{}},
	OpMinus:         {"OpMinus", []int{}},
	OpBang:          {"OpBang", []int{}},
	OpJumpNotTruthy: {"OpJumpNotTruthy", []int{2}},
	OpJump:          {"OpJump", []int{2}},
	OpNull:          {"OpNull", []int{}},
	OpSetGlobal:     {"OpSetGlobal", []int{2}},
	OpGetGlobal:     {"OpGetGlobal", []int{2}},
	OpArray:         {"OpArray", []int{2}},
	OpHash:          {"OpHash", []int{2}},
	OpIndex:         {"OpIndex", []int{}},
	OpCall:          {"OpCall", []int{1}},
	OpReturnValue:   {"OpReturnValue", []int{}},
	OpReturn:        {"OpReturn", []int{}},
	OpSetLocal:      {"OpSetLocal", []int{1}},
	OpGetLocal:      {"OpGetLocal", []int{1}},
	OpGetBuiltin:    {"OpGetBuiltin", []int{1}},
}

// Lookup looksup an opcode and returns its definition if found. otherwise, returns an error.
func Lookup(op byte) (*Definition, error) {
	def, ok := definitions[Opcode(op)]
	if !ok {
		return nil, fmt.Errorf("opcode %d undefined", op)
	}
	return def, nil
}

// Make creates an instruction byte slice based on the given opcode and operands.
// It returns the instruction byte slice.
// If the opcode is not found in the definitions, it returns an empty byte slice.
// The length of the instruction is determined by adding the length of each operand.
// The instruction byte slice is created with the opcode as the first byte.
// For each operand, the width of the operand is determined and the value is converted to BigEndian byte.
// The offset is increased based on the width of the operand.
func Make(op Opcode, operands ...int) []byte {
	// Find the opcode definition
	def, ok := definitions[op]
	if !ok {
		return []byte{}
	}

	// Get the length of the instruction by adding the length of each operand
	instructionLen := 1
	for _, w := range def.OperandWidths {
		instructionLen += w
	}

	// Create the instruction slice
	instruction := make([]byte, instructionLen)
	// Add the opcode first
	instruction[0] = byte(op)

	// For each operand, get the width of the current operands position
	// convert to BigEndian byte, and increase the offset
	offset := 1
	for i, o := range operands {
		width := def.OperandWidths[i]
		switch width {
		case 2:
			binary.BigEndian.PutUint16(instruction[offset:], uint16(o))
		case 1:
			instruction[offset] = byte(o)
		}
		offset += width
	}

	return instruction
}

// ReadOperands reads the operands from the given instructions based on the provided definition.
// It returns the operands as a slice of integers and the offset after reading the operands.
func ReadOperands(def *Definition, ins Instructions) ([]int, int) {
	operands := make([]int, len(def.OperandWidths))
	offset := 0

	for i, width := range def.OperandWidths {
		switch width {
		case 2:
			operands[i] = int(ReadUInt16(ins[offset:]))
		case 1:
			operands[i] = int(ReadUInt8(ins[offset:]))
		}
		offset += width
	}

	return operands, offset
}

// ReadUInt16 reads a uint16 value from the given byte slice.
// It assumes that the byte slice is in big-endian byte order.
func ReadUInt16(ins Instructions) uint16 {
	return binary.BigEndian.Uint16(ins)
}

// ReadUInt8 reads in a uint8 value from the given byte slice.
func ReadUInt8(ins Instructions) uint8 {
	return uint8(ins[0])
}
