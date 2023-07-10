/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2023 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package bytecode

// OpCode is an opcode in the Romualdo Virtual Machine.
type OpCode uint8

const (
	OpNop OpCode = iota
	OpConstant
	OpSay
	OpListen
	OpPop
	OpTrue
	OpFalse
	OpJumpIfFalse
	OpJump
)
