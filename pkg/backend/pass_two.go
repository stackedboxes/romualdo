/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2025 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package backend

import (
	"github.com/stackedboxes/romualdo/pkg/ast"
	"github.com/stackedboxes/romualdo/pkg/bytecode"
)

// codeGeneratorPassTwo does the actual bytecode generation. It fills in the
// Chunks with bytecode.
//
// This implements the ast.Visitor interface.
type codeGeneratorPassTwo struct {
	codeGenerator *codeGenerator

	// currentChunkIndex contains the index of the chunk we are currently
	// generating code for.
	currentChunkIndex int
}

//
// The ast.Visitor interface
//

func (cg *codeGeneratorPassTwo) Enter(node ast.Node) {
	cg.codeGenerator.pushIntoNodeStack(node)

	switch n := node.(type) {
	case *ast.Block:
		cg.codeGenerator.beginScope()

	case *ast.ProcedureDecl:
		cg.currentChunkIndex = n.ChunkIndex

	default:
		// nothing
	}
}

func (cg *codeGeneratorPassTwo) Leave(node ast.Node) {
	defer cg.codeGenerator.popFromNodeStack()

	switch n := node.(type) {
	case *ast.Storyworld:
		break

	case *ast.Block:
		cg.codeGenerator.endScope()

	case *ast.ProcedureDecl:
		// No need to worry about duplicate `main`s: the semantic checker
		// already verified this.
		//
		// TODO: Not yet! There's no semantic checker for now.
		if n.Name == "main" && n.Package == "/" {
			cg.codeGenerator.csw.InitialChunk = cg.currentChunkIndex
		}

		// Leave the current chunk index invalid, as we are outside of any function.
		cg.currentChunkIndex = -1

	case *ast.Say:
		// This is a no-op. The actual `say`ing is done by the Lecture nodes
		// within the Say node.

	case *ast.Lecture:
		// Lectures are automatically "said". They live this double life of
		// being like both a literal and a statement.
		cg.emitConstant(bytecode.NewValueLecture(n.Text))
		cg.emitBytes(byte(bytecode.OpSay))

	case *ast.BoolLiteral:
		if n.Value {
			cg.emitBytes(byte(bytecode.OpTrue))
		} else {
			cg.emitBytes(byte(bytecode.OpFalse))
		}

	case *ast.StringLiteral:
		cg.emitConstant(bytecode.NewValueString(n.Value))

	case *ast.IfStmt:
		break

	case *ast.Listen:
		cg.emitBytes(byte(bytecode.OpListen))

	case *ast.ExpressionStmt:
		// A call to a void Procedure is still an expression statement for
		// grammar purposes -- but one that does not push anything into the
		// stack. Don't try to pop what isn't there.
		if n.Expr.Type() != ast.TypeVoid {
			cg.emitBytes(byte(bytecode.OpPop))
		}

	case *ast.Binary:
		switch n.Operator {
		case "!=":
			cg.emitBytes(byte(bytecode.OpNotEqual))
		case "==":
			cg.emitBytes(byte(bytecode.OpEqual))
		default:
			cg.codeGenerator.ice("unknown binary operator: %v", n.Operator)
		}

	case *ast.Curlies:
		// The Curlies expression value shall be on the stack now.
		cg.emitBytes(byte(bytecode.OpToLecture))
		cg.emitBytes(byte(bytecode.OpSay))

	default:
		cg.codeGenerator.ice("unknown node type: %T", n)
	}
}

func (cg *codeGeneratorPassTwo) Event(node ast.Node, event ast.EventType) {
	switch n := node.(type) {
	case *ast.IfStmt:
		switch event {
		case ast.EventAfterIfCondition:
			// Right after evaluating the condition, we need to jump over the
			// "then" block if the condition is false. We don't know yet the
			// jump offset, so we emit a jump instruction with a dummy offset
			// of 0 we'll patch later in the EventAfterThenBlock event.
			n.IfJumpAddress = len(cg.currentChunk().Code)
			cg.emitBytes(byte(bytecode.OpJumpIfFalse), 0x00, 0x00, 0x00, 0x00)

		case ast.EventAfterThenBlock:
			// At this point we just finished generating the code for the "then"
			// block. We therefore know how large this code is, and therefore
			// this is the moment to patch the jump instruction we emitted
			// in the EventAfterIfCondition with the correct offset.
			addressToPatch := n.IfJumpAddress
			code := cg.currentChunk().Code
			jumpOffset := len(code) - addressToPatch
			cg.patchJump(addressToPatch, jumpOffset)

		case ast.EventBeforeElse:
			// We unconditionally jump over the "else" block as if the "if"
			// condition was true (if it was false, we'd jump over this
			// unconditional jump right into the actual code for the "else"
			// block).
			//
			// We do the same as in EventAfterIfCondition: emit a jump
			// instruction with a dummy offset of 0 that we'll patch later in
			// the EventAfterElse event.
			n.ElseJumpAddress = len(cg.currentChunk().Code)
			cg.emitBytes(byte(bytecode.OpJump), 0x00, 0x00, 0x00, 0x00)

			// Because we had an "else" block, we needed to emit that
			// unconditional jump. So now we need to re-patch the jump
			// instruction at IfJumpAddress to take into account this
			// additional unconditional jump.
			addressToPatch := n.IfJumpAddress
			code := cg.currentChunk().Code
			currentOffset := bytecode.DecodeInt32(code[addressToPatch+1:])
			jumpOffset := currentOffset + 5
			cg.patchJump(addressToPatch, jumpOffset)

		case ast.EventAfterElse:
			// At this point we just finished generating the code for the "else"
			// block. We therefore need to patch the jump instruction we emitted
			// in the EventBeforeElse with the correct offset. (Just like we did
			// in EventAfterThenBlock.)
			addressToPatch := n.ElseJumpAddress
			code := cg.currentChunk().Code
			jumpOffset := len(code) - addressToPatch
			cg.patchJump(addressToPatch, jumpOffset)

		default:
			cg.codeGenerator.ice("Unexpected event while generating code for 'if' statement: %v", event)
		}
	}
}

//
// Actual code generation
//

// emitBytes writes one or more bytes to the bytecode chunk being generated.
func (cg *codeGeneratorPassTwo) emitBytes(bytes ...byte) {
	for _, b := range bytes {
		chunk := cg.currentChunk()
		chunk.Code = append(chunk.Code, b)
		lines := cg.currentLines()
		*lines = append(*lines, cg.codeGenerator.currentLine())
	}
}

// emitConstant emits the bytecode for a constant having a given value.
func (cg *codeGeneratorPassTwo) emitConstant(value bytecode.Value) {
	// TODO: Not handling globals yet
	//
	// if cg.codeGenerator.isInsideGlobalsBlock() {
	// 	// Globals are initialized directly from the initializer value from the
	// 	// AST. No need to push the initializer value to the stack.
	// 	return
	// }

	constantIndex := cg.makeConstant(value)
	operandStart := len(cg.currentChunk().Code) + 1
	cg.emitBytes(byte(bytecode.OpConstant), 0, 0, 0, 0)
	bytecode.EncodeUInt31(cg.currentChunk().Code[operandStart:], constantIndex)
}

// makeConstant adds value to the pool of constants and returns the index in
// which it was added. If there is already a constant with this value, its index
// is returned (hey, we don't need duplicate constants, right? They are
// constant, after all!)
func (cg *codeGeneratorPassTwo) makeConstant(value bytecode.Value) int {
	if i := cg.codeGenerator.csw.SearchConstant(value); i >= 0 {
		return i
	}

	constantIndex := cg.codeGenerator.csw.AddConstant(value)
	if constantIndex >= int(bytecode.MaxConstants) {
		cg.codeGenerator.error("Too many constants in one Storyworld, the maximum is %v.", bytecode.MaxConstants)
		return 0
	}

	return constantIndex
}

// patchJump patches a jump instruction, that is to say, sets the operand of the
// jump instruction at addressToPatch to jumpOffset.
func (cg *codeGeneratorPassTwo) patchJump(addressToPatch, jumpOffset int) {
	bytecode.EncodeInt32(cg.currentChunk().Code[addressToPatch+1:], jumpOffset)
}

//
// Helpers
//

// currentLines returns the current array mapping instructions to source code
// lines.
//
// TODO: Returning a pointer to a slice is ugly as hell, and leads to even
// uglier client code.
func (cg *codeGeneratorPassTwo) currentLines() *[]int {
	return &cg.codeGenerator.debugInfo.ChunksLines[cg.currentChunkIndex]
}

// currentChunk returns the current chunk we are compiling into.
func (cg *codeGeneratorPassTwo) currentChunk() *bytecode.Chunk {
	return cg.codeGenerator.csw.Chunks[cg.currentChunkIndex]
}
