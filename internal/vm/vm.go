package vm

import (
	"fmt"

	"github.com/JosueMolinaMorales/orionlang/internal/code"
	"github.com/JosueMolinaMorales/orionlang/internal/compiler"
	"github.com/JosueMolinaMorales/orionlang/internal/object"
)

const (
	StackSize   = 2046
	GlobalsSize = 65536
	MaxFrames   = 1024
)

var (
	True  = &object.Boolean{Value: true}
	False = &object.Boolean{Value: false}
	Null  = &object.Null{}
)

// VM represents a virtual machine that executes bytecode instructions.
type VM struct {
	constants []object.Object

	stack   []object.Object
	sp      int // Always points to the next value. Top of stack is stack[sp-1]
	globals []object.Object

	frames      []*Frame
	framesIndex int
}

// New creates a new instance of the VM with the given bytecode.
// It initializes the VM's instructions, constants, stack, and stack pointer.
func New(bytecode *compiler.Bytecode) *VM {
	mainFn := &object.CompiledFunction{Instructions: bytecode.Instructions}
	mainFrame := NewFrame(mainFn)

	frames := make([]*Frame, MaxFrames)
	frames[0] = mainFrame

	return &VM{
		constants: bytecode.Constants,

		stack: make([]object.Object, StackSize),
		sp:    0,

		globals: make([]object.Object, GlobalsSize),

		frames:      frames,
		framesIndex: 1,
	}
}

// NewWithGlobalsStore creates a new VM with the given bytecode and global store.
// It initializes a new VM, sets the bytecode, and assigns the global store to the VM's globals field.
// Returns the newly created VM.
func NewWithGlobalsStore(bytecode *compiler.Bytecode, s []object.Object) *VM {
	vm := New(bytecode)
	vm.globals = s
	return vm
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
	var ip int
	var ins code.Instructions
	var op code.Opcode

	for vm.currentFrame().ip < len(vm.currentFrame().Instructions())-1 {
		vm.currentFrame().ip++

		ip = vm.currentFrame().ip
		ins = vm.currentFrame().Instructions()
		op = code.Opcode(ins[ip])

		switch op {
		case code.OpConstant:
			constIndex := code.ReadUInt16(ins[ip+1:])
			vm.currentFrame().ip += 2
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
		case code.OpEqual, code.OpGreaterThan, code.OpNotEqual:
			err := vm.executeComparisonOperation(op)
			if err != nil {
				return nil
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
		case code.OpBang:
			err := vm.executeBangOperator()
			if err != nil {
				return nil
			}
		case code.OpMinus:
			err := vm.executeMinusOperator()
			if err != nil {
				return err
			}
		case code.OpJump:
			pos := int(code.ReadUInt16(ins[ip+1:]))
			vm.currentFrame().ip = pos - 1
		case code.OpJumpNotTruthy:
			pos := int(code.ReadUInt16(ins[ip+1:]))
			vm.currentFrame().ip += 2

			condition := vm.pop()
			if !isTruthy(condition) {
				vm.currentFrame().ip = pos - 1
			}
		case code.OpNull:
			err := vm.push(Null)
			if err != nil {
				return err
			}
		case code.OpSetGlobal:
			globalIndex := code.ReadUInt16(ins[ip+1:])
			vm.currentFrame().ip += 2

			vm.globals[globalIndex] = vm.pop()
		case code.OpGetGlobal:
			globalIndex := code.ReadUInt16(ins[ip+1:])
			vm.currentFrame().ip += 2

			err := vm.push(vm.globals[globalIndex])
			if err != nil {
				return err
			}
		case code.OpArray:
			numElements := int(code.ReadUInt16(ins[ip+1:]))
			vm.currentFrame().ip += 2

			array := vm.buildArray(vm.sp-numElements, vm.sp)
			vm.sp = vm.sp - numElements

			err := vm.push(array)
			if err != nil {
				return err
			}
		case code.OpHash:
			numElements := int(code.ReadUInt16(ins[ip+1:]))
			vm.currentFrame().ip += 2

			hash, err := vm.buildHash(vm.sp-numElements, vm.sp)
			if err != nil {
				return err
			}
			vm.sp = vm.sp - numElements
			err = vm.push(hash)
			if err != nil {
				return err
			}
		case code.OpIndex:
			index := vm.pop()
			left := vm.pop()

			err := vm.executeIndexExpression(left, index)
			if err != nil {
				return err
			}
		case code.OpCall:
			fn, ok := vm.stack[vm.sp-1].(*object.CompiledFunction)
			if !ok {
				return fmt.Errorf("calling non-function")
			}
			frame := NewFrame(fn)
			vm.pushFrame(frame)
		case code.OpReturnValue:
			returnValue := vm.pop()

			vm.popFrame()
			vm.pop()

			err := vm.push(returnValue)
			if err != nil {
				return err
			}
		case code.OpReturn:
			vm.popFrame()
			vm.pop()

			err := vm.push(Null)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// LastPoppedStackElem returns the last element popped from the stack.
func (vm *VM) LastPoppedStackElem() object.Object {
	return vm.stack[vm.sp]
}

// currentFrame returns the current frame in the VM's call stack.
func (vm *VM) currentFrame() *Frame {
	return vm.frames[vm.framesIndex-1]
}

// pushFrame pushes a new frame onto the VM's call stack.
func (vm *VM) pushFrame(f *Frame) {
	vm.frames[vm.framesIndex] = f
	vm.framesIndex++
}

// popFrame pops the top frame from the VM's call stack.
func (vm *VM) popFrame() *Frame {
	vm.framesIndex--
	return vm.frames[vm.framesIndex]
}

// buildHash builds a hash object from a range of stack elements.
// It takes the start and end indices of the range and returns the resulting hash object and an error, if any.
func (vm *VM) buildHash(startIndex, endIndex int) (object.Object, error) {
	hashedPairs := make(map[object.HashKey]object.HashPair)

	for i := startIndex; i < endIndex; i += 2 {
		key := vm.stack[i]
		value := vm.stack[i+1]

		pair := object.HashPair{Key: key, Value: value}

		hashKey, ok := key.(object.Hashable)
		if !ok {
			return nil, fmt.Errorf("unusable has hash key: %s", key.Type())
		}

		hashedPairs[hashKey.HashKey()] = pair
	}

	return &object.Hash{Pairs: hashedPairs}, nil
}

// buildArray builds an array object from the elements on the VM stack.
// It takes the start and end indices of the elements to include in the array.
// It returns a pointer to the created array object.
func (vm *VM) buildArray(startIndex, endIndex int) object.Object {
	elements := make([]object.Object, endIndex-startIndex)

	for i := startIndex; i < endIndex; i++ {
		elements[i-startIndex] = vm.stack[i]
	}

	return &object.Array{Elements: elements}
}

// executeIndexExpression executes the index expression for the given left and index objects.
// It supports indexing on arrays and hashes.
// If the index operator is not supported for the given left object, it returns an error.
func (vm *VM) executeIndexExpression(left, index object.Object) error {
	switch {
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		return vm.executeArrayIndex(left, index)
	case left.Type() == object.HASH_OBJ:
		return vm.executeHashIndex(left, index)
	default:
		return fmt.Errorf("index operator not supported: %s", left.Type())
	}
}

// executeArrayIndex executes the array index operation on the virtual machine.
// It takes an array object and an index object as arguments and returns an error.
// If the index is out of range, it pushes null onto the stack. Otherwise, it pushes
// the element at the specified index onto the stack.
func (vm *VM) executeArrayIndex(array, index object.Object) error {
	arrayObject := array.(*object.Array)
	i := index.(*object.Integer).Value
	max := int64(len(arrayObject.Elements) - 1)

	if i < 0 || i > max {
		return vm.push(Null)
	}

	return vm.push(arrayObject.Elements[i])
}

// executeHashIndex executes the hash index operation on the virtual machine.
// It takes a hash object and an index object as parameters and returns an error.
// If the index object is not usable as a hash key, it returns an error.
// If the hash does not contain the specified key, it pushes null onto the stack.
// Otherwise, it pushes the value associated with the key onto the stack.
func (vm *VM) executeHashIndex(hash, index object.Object) error {
	hashObject := hash.(*object.Hash)

	key, ok := index.(object.Hashable)
	if !ok {
		return fmt.Errorf("unusable as hash key: %s", index.Type())
	}

	pair, ok := hashObject.Pairs[key.HashKey()]
	if !ok {
		return vm.push(Null)
	}

	return vm.push(pair.Value)
}

// executeMinusOperator performs the execution of the minus operator in the virtual machine.
// It pops an operand from the stack and checks if it is of type INTEGER_OBJ.
// If the operand is not an integer, it returns an error.
// Otherwise, it negates the value of the integer and pushes the result back onto the stack.
func (vm *VM) executeMinusOperator() error {
	operand := vm.pop()

	if operand.Type() != object.INTEGER_OBJ {
		return fmt.Errorf("unsupported type for negation: %s", operand.Type())
	}

	value := operand.(*object.Integer).Value
	return vm.push(&object.Integer{Value: -value})
}

// executeBangOperator performs the logical negation operation on the top value of the stack.
// If the top value is True, it pushes False onto the stack.
// If the top value is False or Null, it pushes True onto the stack.
// For any other value, it pushes False onto the stack.
func (vm *VM) executeBangOperator() error {
	operand := vm.pop()

	switch operand {
	case True:
		return vm.push(False)
	case False:
		return vm.push(True)
	case Null:
		return vm.push(True)
	default:
		return vm.push(False)
	}
}

// executeComparisonOperation executes a comparison operation on the top two values on the VM's stack.
// It takes an opcode as a parameter and returns an error if the operation fails.
// The method first pops the top two values from the stack and checks their types.
// If both values are integers, it calls the executeIntegerComparison method to perform the comparison.
// If the opcode is OpEqual, it pushes the result of the comparison (true or false) onto the stack.
// If the opcode is OpNotEqual, it pushes the negation of the comparison result onto the stack.
// If the opcode is not recognized, it returns an error with an unknown operator message.
func (vm *VM) executeComparisonOperation(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	if left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ {
		return vm.executeIntegerComparison(op, left, right)
	}

	switch op {
	case code.OpEqual:
		return vm.push(nativeBoolToBooleanObject(right == left))
	case code.OpNotEqual:
		return vm.push(nativeBoolToBooleanObject(right != left))
	default:
		return fmt.Errorf("unknown operator: %d (%s %s)", op, left.Type(), right.Type())
	}
}

// executeIntegerComparison performs a comparison operation on two integer values.
// It takes an opcode, left operand, and right operand as arguments.
// The function returns an error if the operator is unknown.
func (vm *VM) executeIntegerComparison(op code.Opcode, left, right object.Object) error {
	leftValue := left.(*object.Integer).Value
	rightValue := right.(*object.Integer).Value

	switch op {
	case code.OpEqual:
		return vm.push(nativeBoolToBooleanObject(rightValue == leftValue))
	case code.OpNotEqual:
		return vm.push(nativeBoolToBooleanObject(rightValue != leftValue))
	case code.OpGreaterThan:
		return vm.push(nativeBoolToBooleanObject(leftValue > rightValue))
	default:
		return fmt.Errorf("unknown operator: %d", op)
	}
}

// executeBinaryOperation executes a binary operation on the top two values on the VM's stack.
// It takes an opcode as a parameter and returns an error if the operation is not supported for the given types.
// The method first pops the top two values from the stack and determines their types.
// It then performs the binary operation based on the types of the values.
// If the types are not supported for the operation, it returns an error.
// Supported types for binary operations are integer and string.
// For integer types, it calls the executeBinaryIntegerOperation method.
// For string types, it calls the executeBinaryStringOperation method.
// If the types are not supported, it returns an error indicating the unsupported types.
func (vm *VM) executeBinaryOperation(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	leftType := left.Type()
	rightType := right.Type()

	switch {
	case leftType == object.INTEGER_OBJ && rightType == object.INTEGER_OBJ:
		return vm.executeBinaryIntegerOperation(op, left, right)
	case leftType == object.STRING_OBJ && rightType == object.STRING_OBJ:
		return vm.executeBinaryStringOperation(op, left, right)
	default:
		return fmt.Errorf("unsupported types for binary operation: %s %s", leftType, rightType)
	}
}

// executeBinaryStringOperation performs a binary string operation on the given operands.
// It concatenates the values of the left and right string objects and pushes the result onto the VM stack.
// The only supported operator for string operations is the addition operator (+).
// If the operator is not addition, it returns an error indicating an unknown string operator.
func (vm *VM) executeBinaryStringOperation(op code.Opcode, left, right object.Object) error {
	if op != code.OpAdd {
		return fmt.Errorf("unknown string operator: %d", op)
	}

	leftValue := left.(*object.String).Value
	rightValue := right.(*object.String).Value

	return vm.push(&object.String{Value: leftValue + rightValue})
}

// executeBinaryIntegerOperation executes a binary integer operation on the virtual machine.
// It takes an opcode, left operand, and right operand as arguments and returns an error if any.
// The function performs the specified operation on the integer values and pushes the result onto the stack.
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

// pop removes and returns the top element from the stack.
func (vm *VM) pop() object.Object {
	o := vm.stack[vm.sp-1]
	vm.sp--
	return o
}

// nativeBoolToBooleanObject converts a native bool value to a Boolean object.
// If the input is true, it returns the True object. Otherwise, it returns the False object.
func nativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return True
	}
	return False
}

// isTruthy checks if the given object is considered truthy.
// It returns true if the object is a non-null boolean with a value of true,
// and false otherwise.
func isTruthy(obj object.Object) bool {
	switch obj := obj.(type) {
	case *object.Boolean:
		return obj.Value
	case *object.Null:
		return false
	default:
		return true
	}
}
