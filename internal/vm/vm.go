package vm

import (
	"fmt"

	"github.com/JosueMolinaMorales/orionlang/internal/code"
	"github.com/JosueMolinaMorales/orionlang/internal/compiler"
	"github.com/JosueMolinaMorales/orionlang/internal/object"
)

const StackSize = 2046

var (
	True  = &object.Boolean{Value: true}
	False = &object.Boolean{Value: false}
)

// VM represents a virtual machine that executes bytecode instructions.
type VM struct {
	constants    []object.Object
	instructions code.Instructions

	stack []object.Object
	sp    int // Always points to the next value. Top of stack is stack[sp-1]
}

// New creates a new instance of the VM with the given bytecode.
// It initializes the VM's instructions, constants, stack, and stack pointer.
func New(bytecode *compiler.Bytecode) *VM {
	return &VM{
		instructions: bytecode.Instructions,
		constants:    bytecode.Constants,
		stack:        make([]object.Object, StackSize),
		sp:           0,
	}
}

// StackTop returns the top element of the stack.
// If the stack is empty, it returns nil.
func (vm *VM) StackTop() object.Object {
	if vm.sp == 0 {
		return nil
	}
	return vm.stack[vm.sp-1]
}

// Run executes the instructions stored in the VM.
// It iterates over each instruction, fetches the current instruction,
// and performs the corresponding operation based on the OpCode.
// If an error occurs during execution, it is returned.
func (vm *VM) Run() error {
	for ip := 0; ip < len(vm.instructions); ip++ {
		// Fetch the current instruction and turn the byte into an OpCode
		op := code.Opcode(vm.instructions[ip])

		switch op {
		case code.OpConstant:
			constIndex := code.ReadUInt16(vm.instructions[ip+1:])
			ip += 2

			err := vm.push(vm.constants[constIndex])
			if err != nil {
				return err
			}
		case code.OpPop:
			vm.pop()
		case code.OpAdd, code.OpDivide, code.OpMultiply, code.OpSubtract:
			err := vm.executeBinaryOperation(op)
			if err != nil {
				return err
			}
		case code.OpTrue:
			err := vm.push(True)
			if err != nil {
				return err
			}
		case code.OpFalse:
			err := vm.push(False)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (vm *VM) LastPoppedStackElem() object.Object {
	return vm.stack[vm.sp]
}

func (vm *VM) executeBinaryOperation(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	leftType := left.Type()
	rightType := right.Type()

	if leftType != object.INTEGER_OBJ || rightType != object.INTEGER_OBJ {
		return fmt.Errorf("unsupported types for binary operation: %s %s", leftType, rightType)
	}

	return vm.executeBinaryIntegerOperation(op, left, right)
}

func (vm *VM) executeBinaryIntegerOperation(op code.Opcode, left, right object.Object) error {
	rightValue := right.(*object.Integer).Value
	leftValue := left.(*object.Integer).Value

	var result int64

	switch op {
	case code.OpAdd:
		result = leftValue + rightValue
	case code.OpSubtract:
		result = leftValue - rightValue
	case code.OpMultiply:
		result = leftValue * rightValue
	case code.OpDivide:
		result = leftValue / rightValue
	default:
		return fmt.Errorf("unknown integer operator: %d", op)
	}

	return vm.push(&object.Integer{Value: result})
}

// push pushes the given object onto the stack.
// It returns an error if the stack is already full.
func (vm *VM) push(o object.Object) error {
	// check to see if the stackpointer went over the limit
	if vm.sp >= StackSize {
		return fmt.Errorf("stack overflow")
	}

	vm.stack[vm.sp] = o
	vm.sp++

	return nil
}

func (vm *VM) pop() object.Object {
	o := vm.stack[vm.sp-1]
	vm.sp--
	return o
}
