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
)

// DisassembleChunk disassembles a whole chunk and writes the output to out.
// debugInfo is optional: if not nil, it will be used for better disassembly.
func (csw *CompiledStoryworld) DisassembleChunk(chunk *Chunk, out io.Writer, debugInfo *DebugInfo, chunkIndex int) {
	for offset := 0; offset < len(chunk.Code); {
		offset = csw.DisassembleInstruction(chunk, out, offset, debugInfo, chunkIndex)
	}
}

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

	if offset > 0 && lines[offset] == lines[offset-1] {
		fmt.Fprintf(out, "    | ")
	} else {
		fmt.Fprintf(out, "%5d ", lines[offset])
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

	case OpListen:
		return csw.disassembleSimpleInstruction(out, "LISTEN", offset)

	case OpPop:
		return csw.disassembleSimpleInstruction(out, "POP", offset)

	case OpTrue:
		return csw.disassembleSimpleInstruction(out, "TRUE", offset)

	case OpFalse:
		return csw.disassembleSimpleInstruction(out, "FALSE", offset)

	case OpJump:
		return csw.disassembleInt32Instruction(chunk, out, "JUMP", offset)

	case OpJumpIfFalse:
		return csw.disassembleInt32Instruction(chunk, out, "JUMP_IF_FALSE", offset)

	case OpEqual:
		return csw.disassembleSimpleInstruction(out, "EQUAL", offset)

	case OpNotEqual:
		return csw.disassembleSimpleInstruction(out, "NOT_EQUAL", offset)

	case OpToString:
		return csw.disassembleSimpleInstruction(out, "TO_STRING", offset)

	case OpToLecture:
		return csw.disassembleSimpleInstruction(out, "TO_LECTURE", offset)

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
	fmt.Fprintf(out, "%-16s %4d %v\n", name, index, csw.Constants[index].DebugString(di))
	return offset + 5
}

// disassembleInt32Instruction disassembles an instruction that has a single
// int32 operand.
func (csw *CompiledStoryworld) disassembleInt32Instruction(chunk *Chunk, out io.Writer, name string, offset int) int {
	operand := DecodeInt32(chunk.Code[offset+1:])
	fmt.Fprintf(out, "%-16s %4d\n", name, operand)
	return offset + 5
}
