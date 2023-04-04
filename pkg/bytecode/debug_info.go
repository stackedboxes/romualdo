/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2023 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package bytecode

// DebugInfo contains debug information matching a CompiledStoryworld. All
// information that is not strictly necessary to run a Storyworld but is useful
// for debugging, producing better error reporting, etc, belongs here.
//
// TODO: Make it serializable and deserializable. All serialized data shall be
// little endian.
type DebugInfo struct {
	// ChunksNames contains the names of the procedures on a CompiledStoryworld.
	// There is one entry for each entry in the corresponding
	// CompiledStoryworld.Chunks.
	ChunksNames []string

	// ChunksSourceFiles contains the source files every Chunk was compiled
	// from. The indices here match those in CompiledStoryworld.Chunks. The file
	// names here contain the path from the root of the Storyworld.
	ChunksSourceFiles []string

	// ChunksLines contains the source code line that generated each instruction
	// of each Chunk. This must be interpreted like this:
	// ChunksLines[chunkIndex][codeIndex] contains the line that generated the
	// bytecode at CompiledStoryworld.Chunks[chunkIndex].Code[codeIndex].
	//
	// Notice that we have one entry for each entry in Code. Very
	// space-inefficient, but very simple.
	//
	// TODO: Use run-length encoding (RLE) or something like that to spare some
	// memory and storage.
	ChunksLines [][]int
}
