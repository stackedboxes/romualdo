/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2023 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package bytecode

import (
	"fmt"
	"io"
	"strings"
)

// DisassembleInstruction disassembles the instruction at a given offset of
// chunk and returns the offset of the next instruction to disassemble. Output
// is written to out. chunkIndex is the index of the current chunk. debugInfo is
// optional: if not nil, it will be used for better disassembly.
func (csw *CompiledStoryworld) DisassembleInstruction(chunk *Chunk, out io.Writer, offset int, debugInfo *DebugInfo, chunkIndex int) int {
	// Offset
	fmt.Fprintf(out, "%05v ", offset)

	// Source file and line
	var lines []int = nil
	if debugInfo != nil {
		lines = debugInfo.ChunksLines[chunkIndex]
	}
	sourceFile := ""
	if debugInfo != nil {
		sourceFile = debugInfo.ChunksSourceFiles[chunkIndex]
	}

	if offset > 0 && lines[offset] == lines[offset-1] {
		blank := strings.Repeat(" ", len(sourceFile)+1)
		fmt.Fprintf(out, "%v    | ", blank)
	} else {
		fmt.Fprintf(out, "%v:%5d ", sourceFile, lines[offset])
	}

	// Instruction
	instruction := OpCode(chunk.Code[offset])

	switch instruction {
	case OpNop:
		return csw.disassembleSimpleInstruction(out, "NOP", offset)

	case OpConstant:
		return csw.disassembleConstantInstruction(chunk, out, "CONSTANT", offset, debugInfo)

	case OpSay:
		return csw.disassembleSimpleInstruction(out, "SAY", offset)

	default:
		fmt.Fprintf(out, "Unknown opcode %d\n", instruction)
		return offset + 1
	}
}

// disassembleSimpleInstruction disassembles a simple instruction at a given
// offset. name is the instruction name, and the output is written to out.
// Returns the offset to the next instruction.
//
// A simple instruction is one composed of a single byte (just the opcode, no
// operands).
func (csw *CompiledStoryworld) disassembleSimpleInstruction(out io.Writer, name string, offset int) int {
	fmt.Fprintf(out, "%v\n", name)
	return offset + 1
}

// disassembleConstantInstruction disassembles a OpConstant instruction at a
// given offset. name is the instruction name, and the output is written to out.
// Returns the offset to the next instruction.
func (csw *CompiledStoryworld) disassembleConstantInstruction(chunk *Chunk, out io.Writer, name string, offset int, di *DebugInfo) int {
	index := DecodeUInt31(chunk.Code[offset+1:])
	fmt.Fprintf(out, "%-16s %4d '%v'\n", name, index, csw.Constants[index].DebugString(di))
	return offset + 5
}
