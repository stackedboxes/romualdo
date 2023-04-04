/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2023 Leandro Motta Barros                                     *
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
			cg.codeGenerator.csw.FirstChunk = cg.currentChunkIndex
		}

		// Leave the current chunk index invalid, as we are outside of any function.
		cg.currentChunkIndex = -1

	case *ast.Lecture:
		// Lectures are automatically "said". They live this double life of
		// being like both a literal and a statement.
		cg.emitConstant(bytecode.NewValueLecture(n.Text))
		cg.emitBytes(byte(bytecode.OpSay))

	default:
		cg.codeGenerator.ice("unknown node type: %T", n)
	}

	cg.codeGenerator.popFromNodeStack()
}

func (cg *codeGeneratorPassTwo) Event(node ast.Node, event int) {
	// Nothing for now
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
	if constantIndex >= bytecode.MaxConstants {
		cg.codeGenerator.error("Too many constants in one Storyworld, the maximum is %v.", bytecode.MaxConstants)
		return 0
	}

	return constantIndex
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
