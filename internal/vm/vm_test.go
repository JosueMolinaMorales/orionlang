package vm

import (
	"fmt"
	"testing"

	"github.com/JosueMolinaMorales/orionlang/internal/ast"
	"github.com/JosueMolinaMorales/orionlang/internal/compiler"
	"github.com/JosueMolinaMorales/orionlang/internal/lexer"
	"github.com/JosueMolinaMorales/orionlang/internal/object"
	"github.com/JosueMolinaMorales/orionlang/internal/parser"
)

type vmTestCase struct {
	input    string
	expected interface{}
}

func runVmTests(t *testing.T, tests []vmTestCase) {
	t.Helper()

	for _, tt := range tests {
		program := parse(tt.input)

		comp := compiler.New()
		err := comp.Compile(program)
		if err != nil {
			t.Fatalf("compiler error: %s", err)
		}

		vm := New(comp.Bytecode())
		err = vm.Run()
		if err != nil {
			t.Fatalf("vm error: %s", err)
		}

		stackElem := vm.LastPoppedStackElem()
		testExpectedObject(t, tt.expected, stackElem)
	}
}

func TestBooleanExpressions(t *testing.T) {
	tests := []vmTestCase{
		{"true", true},
		{"false", false},
		{"1 < 2", true},
		{"1 > 2", false},
		{"1 < 1", false},
		{"1 > 1", false},
		{"1 == 1", true},
		{"1 != 1", false},
		{"1 == 2", false},
		{"1 != 2", true},
		{"true == true", true},
		{"false == false", true},
		{"true == false", false},
		{"true != false", true},
		{"false != true", true},
		{"(1 < 2) == true", true},
		{"(1 < 2) == false", false},
		{"(1 > 2) == true", false},
		{"(1 > 2) == false", true},
		{"!true", false},
		{"!false", true},
		{"!5", false},
		{"!!true", true},
		{"!!false", false},
		{"!!5", true},
	}

	runVmTests(t, tests)
}

func TestIntegerArithmetic(t *testing.T) {
	tests := []vmTestCase{
		{"1", 1},
		{"2", 2},
		{"1 + 2", 3},
		{"1 - 2", -1},
		{"1 * 2", 2},
		{"2 / 1", 2},
		{"50 / 2 * 2 + 10 - 5", 55},
		{"10 + 20 * 30", 610},         // test order of operations
		{"(10 + 20) * 30", 900},       // test parentheses
		{"20 / 5", 4},                 // simple division
		{"5 / 2", 2},                  // integer division
		{"10 * 0", 0},                 // multiplication by zero
		{"0 * 10", 0},                 // zero multiplied by any number
		{"0 + 0", 0},                  // zero addition
		{"0 - 0", 0},                  // zero subtraction
		{"(10 / 2) * 3", 15},          // division followed by multiplication
		{"10 / (2 * 2)", 2},           // division with multiplication in parentheses
		{"(10 + 20) / (2 * 5)", 3},    // combined operations with parentheses
		{"1 + (2 * (2 + 3))", 11},     // nested parentheses
		{"10 - (5 + 5)", 0},           // subtraction with parentheses
		{"100 - ((10 - 5) * 10)", 50}, // nested operations with parentheses
		{"-5", -5},
		{"-10", -10},
		{"-50 + 100 + -50", 0},
		{"(5 + 10 * 2 + 15 / 3) * 2 + -10", 50},
	}

	runVmTests(t, tests)
}

func testExpectedObject(t *testing.T, expected interface{}, actual object.Object) {
	t.Helper()

	switch expected := expected.(type) {
	case int:
		err := testIntegerObject(int64(expected), actual)
		if err != nil {
			t.Errorf("testIntegerObject failed: %s", err)
		}
	case bool:
		err := testBooleanObject(bool(expected), actual)
		if err != nil {
			t.Errorf("testBooleanObject failed: %s", err)
		}
	}
}

func parse(input string) *ast.Program {
	l := lexer.New(input)
	p := parser.New(l)
	return p.ParseProgram()
}

func testBooleanObject(expected bool, actual object.Object) error {
	result, ok := actual.(*object.Boolean)
	if !ok {
		return fmt.Errorf("object is not Boolean. got=%T (%+v)", actual, actual)
	}

	if result.Value != expected {
		return fmt.Errorf("object has wrong value. got=%t, want=%t", result.Value, expected)
	}

	return nil
}

func testIntegerObject(expected int64, actual object.Object) error {
	result, ok := actual.(*object.Integer)
	if !ok {
		return fmt.Errorf("object is not Integer. got=%T (%+v)", actual, actual)
	}

	if result.Value != expected {
		return fmt.Errorf("object has wrong value. got=%d, want=%d", result.Value, expected)
	}

	return nil
}
