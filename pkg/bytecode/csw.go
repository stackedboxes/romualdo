/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2023 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package bytecode

const (
	// MaxConstants is the maximum number of constants we can have on a
	// CompiledStoryworld. This is equal to 2^31, so that it fits on an int even
	// on platforms that use 32-bit integers. And this number should be large
	// enough to ensure we don't run out of space for constants.
	MaxConstants = 2_147_483_648
)

// CompiledStoryworld is a compiled, binary version of a Romualdo Language
// Storyworld.
//
// TODO: Make it serializable and deserializable. All serialized data shall be
// little endian.
//
// TODO: Use a string interner to avoid having duplicate strings in memory.
// Make some measurements to ensure it's really beneficial.
type CompiledStoryworld struct {
	// Chunks is a slide with all Chunks of bytecode containing the compiled
	// data. There is one Chunk for each procedure in the Storyworld.
	//
	// TODO: And in the future, one Chunk for every version of every procedure.
	Chunks []*Chunk

	// FirstChunk indexes the element in Chunks from where the Storyworld
	// execution starts. In other words, it points to the "/main" chunk.
	FirstChunk int

	// The constant values used in all Chunks.
	Constants []Value
}

// SearchConstant searches the constant pool for a constant with the given
// value. If found, it returns the index of this constant into csw.Constants. If
// not found, it returns a negative value.
func (csw *CompiledStoryworld) SearchConstant(value Value) int {
	for i, v := range csw.Constants {
		if ValuesEqual(value, v) {
			return i
		}
	}

	return -1
}

// AddConstant adds a constant to the CompiledStoryworld and returns the index
// of the new constant into csw.Constants.
func (csw *CompiledStoryworld) AddConstant(value Value) int {
	csw.Constants = append(csw.Constants, value)
	return len(csw.Constants) - 1
}
