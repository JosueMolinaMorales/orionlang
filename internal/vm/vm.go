package vm

import (
	"fmt"

	"github.com/JosueMolinaMorales/orionlang/internal/code"
	"github.com/JosueMolinaMorales/orionlang/internal/compiler"
	"github.com/JosueMolinaMorales/orionlang/internal/object"
)

const StackSize = 2046

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
		case code.OpAdd:
			right := vm.pop()
			left := vm.pop()
			leftValue := left.(*object.Integer).Value
			rightValue := right.(*object.Integer).Value
			result := leftValue + rightValue
			vm.push(&object.Integer{Value: result})
		case code.OpPop:
			vm.pop()
		}
	}

	return nil
}

func (vm *VM) LastPoppedStackElem() object.Object {
	return vm.stack[vm.sp]
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
