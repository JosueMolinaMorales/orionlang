package vm

import (
	"github.com/JosueMolinaMorales/orionlang/internal/code"
	"github.com/JosueMolinaMorales/orionlang/internal/object"
)

// Frame represents a frame in the call stack
type Frame struct {
	// fn represents the function referenced by the frame
	fn *object.CompiledFunction
	// ip is the insturction pointer in this frame, for this function
	ip int
	// basePointer is the pointer that points to the bottom of the stack of the current
	// call frame
	basePointer int
}

// NewFrame creates a new frame with the given compiled function and base pointer.
func NewFrame(fn *object.CompiledFunction, basePointer int) *Frame {
	return &Frame{fn: fn, ip: -1, basePointer: basePointer}
}

// Instructions returns the instructions associated with the frame.
func (f *Frame) Instructions() code.Instructions {
	return f.fn.Instructions
}
