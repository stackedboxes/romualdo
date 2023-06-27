/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2023 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package bytecode

import (
	"encoding/binary"
	"math"
)

// A Chunk is a chunk of bytecode. We'll have one Chunk for each procedure in a
// Storyworld.
//
// TODO: In the future, one chunk for each version of each procedure.
//
// TODO: In the future, probably, chunks for implicitly-defined procedures that
// initialize globals and stuff.
type Chunk struct {
	// The bytecode itself. Includes both OpCodes and immediate arguments needed
	// by the opcodes.
	Code []uint8
}

// Encodes an unsigned 31-bit integer into the four first bytes of bytecode.
// Panics if v does not fit into 31 bits.
func EncodeUInt31(bytecode []byte, v int) {
	if v < 0 || v > math.MaxInt32 {
		panic("Value does not fit into 31 bits")
	}
	binary.LittleEndian.PutUint32(bytecode, uint32(v))
}

// Decodes the first four bytes in bytecode into an unsigned 31-bit integer.
// Panics if the value read does not fit into 31 bits.
func DecodeUInt31(bytecode []byte) int {
	v := binary.LittleEndian.Uint32(bytecode)
	if v > math.MaxInt32 {
		panic("Value does not fit into 31 bits")
	}
	return int(v)
}
