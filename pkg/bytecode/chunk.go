/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2025 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package bytecode

import (
	"encoding/binary"
	"math"

	"github.com/stackedboxes/romualdo/pkg/romutil"
)

// A Chunk is a chunk of bytecode. We have one Chunk for each version of each
// Procedure in a Storyworld.
//
// TODO: In the future, maybe, chunks for implicitly-defined procedures that
// initialize globals and stuff. (Doesn't feel like the simplest way to
// initialize globals, but it's a possibility.)
type Chunk struct {
	// Code contains the bytecode itself. Includes both OpCodes and immediate
	// arguments needed by the opcodes.
	Code []uint8

	// Released tells whether this Chunk belongs to a released version or not. A
	// value of true means this Chunk is the result of a `romualdo release`;
	// false means it came form a `romualdo build`.
	//
	// Released Chunks must remain in the Compiled Storyworld forever to ensure
	// compatibility with saved story progress from old releases of the
	// Storyworld.
	Released bool

	// Hash is the code hash of the source code that generated this Chunk. It is
	// used by the compiler at build time to check if a given Procedure has
	// changed since the last release.
	Hash romutil.CodeHash
}

// Encodes a signed 32-bit integer into the four first bytes of bytecode.
func EncodeInt32(bytecode []byte, v int) {
	binary.LittleEndian.PutUint32(bytecode, uint32(v))
}

// Decodes the first four bytes in bytecode into a signed 32-bit integer.
func DecodeInt32(bytecode []byte) int {
	v := binary.LittleEndian.Uint32(bytecode)
	return int(v)
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
