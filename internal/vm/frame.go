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
}

func NewFrame(fn *object.CompiledFunction) *Frame {
	return &Frame{fn: fn, ip: -1}
}

func (f *Frame) Instructions() code.Instructions {
	return f.fn.Instructions
}
