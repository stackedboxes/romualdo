/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2023 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package vm

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/stackedboxes/romualdo/pkg/bytecode"
	"github.com/stackedboxes/romualdo/pkg/errs"
)

// VM is a Romualdo Virtual Machine.
type VM struct {
	// Set DebugTraceExecution to true to make the VM disassemble the code as it
	// runs through it.
	DebugTraceExecution bool

	// out is where the VM sends its output.
	out io.Writer

	// csw is the compiled storyworld we are executing.
	csw *bytecode.CompiledStoryworld

	// debugInfo contains the debug information corresponding to csw.
	// TODO: Make this optional. If nil, issue less friendly error messages,
	// etc.
	debugInfo *bytecode.DebugInfo

	// stack is the VM stack, used for storing values during interpretation.
	stack *Stack

	// frames is the stack of call frames. It has one entry for every function
	// that has started running bit hasn't returned yet.
	frames []*callFrame

	// The current call frame (the one on top of VM.frames).
	frame *callFrame
}

// New returns a new Virtual Machine. out is where the VM sends its output.
func New(out io.Writer) *VM {
	return &VM{
		stack: &Stack{},
		out:   out,
	}
}

// currentChunk returns the chunk currently being executed.
func (vm *VM) currentChunk() *bytecode.Chunk {
	return vm.csw.Chunks[vm.frame.proc.ChunkIndex]
}

// Interpret interprets a given compiled Storyworld.
// TODO: DebugInfo should be optional.
func (vm *VM) Interpret(csw *bytecode.CompiledStoryworld, di *bytecode.DebugInfo) (err errs.Error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(*errs.Runtime); ok {
				err = e
				return
			}
			err = errs.NewICE("Unexpected error type: %T", r)
			return
		}
	}()

	vm.csw = csw
	vm.debugInfo = di

	// Normal Procedure calls start by pushing the callable thing. Here we have
	// an implicit call to the initial Procedure, so we push it. This keeps this
	// implicit call consistent with calls made by the user, and avoid having to
	// treat it as a special case elsewhere.
	vm.push(bytecode.NewValueProcedure(csw.InitialChunk))
	proc := bytecode.Procedure{ChunkIndex: csw.InitialChunk}
	vm.callProcedure(proc, 0)
	vm.frame = vm.frames[0]

	r := vm.run()

	// TODO: This will be true once we have proper procedure calls and returns.
	//       Right now, the main procedure will reman on the stack.
	// if vm.stack.size() != 0 {
	// 	vm.runtimeError("Stack size should be zero after execution, was %v.", vm.stack.size())
	// }

	return r
}

// run runs the code loaded into vm.
func (vm *VM) run() errs.Error {
	for {
		// TODO: Temporary hack to detect the of end of a program. Eventually,
		// this will be done by the return instruction.
		if vm.frame.ip >= len(vm.currentChunk().Code) {
			return nil
		}

		if vm.DebugTraceExecution {
			fmt.Print("Stack: ")

			for _, v := range vm.stack.data {
				fmt.Printf("[ %v ]", v.DebugString(vm.debugInfo))
			}

			fmt.Print("\n")

			chunkIndex := vm.frame.proc.ChunkIndex
			vm.csw.DisassembleInstruction(vm.currentChunk(), os.Stdout, vm.frame.ip, vm.debugInfo, chunkIndex)
		}

		currentChunk := vm.currentChunk()
		instruction := currentChunk.Code[vm.frame.ip]
		vm.frame.ip++

		switch bytecode.OpCode(instruction) {
		case bytecode.OpNop:
			break

		case bytecode.OpConstant:
			constant := vm.readConstant()
			vm.push(constant)

		case bytecode.OpSay:
			value := vm.pop()
			if !value.IsLecture() {
				vm.runtimeError("Expected a Lecture, got %T", value.Value)
			}
			fmt.Fprintf(vm.out, "%v", value.AsLecture().Text)

		case bytecode.OpListen:
			options := vm.pop()
			fmt.Fprintf(vm.out, "==> %v\n", options.AsString())

			// TODO: Don't read from stdin, need to be more versatile for testing
			// and real use.

			// TODO: Implement proper return to driver program and stuff.
			fmt.Fprint(vm.out, "> ")
			scanner := bufio.NewScanner(os.Stdin)
			scanner.Scan()
			choice := scanner.Text()
			fmt.Fprintf(vm.out, "USER INPUT: %v", choice)
			vm.push(bytecode.NewValueString(choice))

		case bytecode.OpPop:
			vm.pop()

		default:
			vm.runtimeError("Unexpected instruction: %v", instruction)
		}
	}
}

// readConstant reads a 32-bit constant index from the chunk bytecode and
// returns the corresponding constant value.
func (vm *VM) readConstant() bytecode.Value {
	chunk := vm.currentChunk()
	index := bytecode.DecodeUInt31(chunk.Code[vm.frame.ip:])
	constant := vm.csw.Constants[index]
	vm.frame.ip += 4
	return constant
}

// push pushes a value into the VM stack.
func (vm *VM) push(value bytecode.Value) {
	vm.stack.push(value)
}

// top returns the value on the top of the VM stack (without removing it).
// Panics on underflow.
func (vm *VM) top() bytecode.Value {
	return vm.stack.top()
}

// pop pops a value from the VM stack and returns it. Panics on underflow.
func (vm *VM) pop() bytecode.Value {
	return vm.stack.pop()
}

// peek returns a value on the stack that is a given distance from the top.
// Passing 0 means "give me the value on the top of the stack". The stack is not
// changed at all.
func (vm *VM) peek(distance int) bytecode.Value {
	return vm.stack.peek(distance)
}

// callProcedure calls Procedure proc. Assumes that the function and its arguments
// were pushed into the stack. Pushes a new frame into vm.frames.
func (vm *VM) callProcedure(proc bytecode.Procedure, argCount int) {
	vm.frames = append(vm.frames, &callFrame{
		proc:  proc,
		stack: vm.stack.createView(argCount + 1), // "+1" is the callee, which is on the stack
	})
}

// runtimeError stops the execution and reports a runtime error with a given
// message and fmt.Printf-like arguments.
func (vm *VM) runtimeError(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", a...)

	stackTrace := strings.Builder{}
	for i := len(vm.frames) - 1; i >= 0; i-- {
		frame := vm.frames[i]
		proc := frame.proc
		instructionOffset := frame.ip - 1
		chunkIndex := proc.ChunkIndex
		lineNumber := vm.debugInfo.ChunksLines[chunkIndex][instructionOffset]
		functionName := vm.debugInfo.ChunksNames[chunkIndex]
		stackTrace.WriteString(fmt.Sprintf("[line %v] in %v\n", lineNumber, functionName))
	}

	stackTrace.WriteRune('\n')
	panic(errs.NewRuntime(stackTrace.String()))
}

// callFrame contains the information needed at runtime about an ongoing
// Procedure call.
type callFrame struct {
	// proc is the Procedure running.
	proc bytecode.Procedure

	// ip is the instruction pointer, which points to the next instruction to be
	// executed (it's an index into proc's chunk).
	ip int

	// stack is a read/write view into the VM stack, and represents the stack
	// that this Procedure can use.
	stack *StackView
}
