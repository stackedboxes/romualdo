/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2023 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package bytecode

import (
	"hash/crc32"
	"io"

	"github.com/stackedboxes/romualdo/pkg/errs"
	"github.com/stackedboxes/romualdo/pkg/romutil"
)

const (
	// MaxConstants is the maximum number of constants we can have on a
	// CompiledStoryworld. This is equal to 2^31, so that it fits on an int even
	// on platforms that use 32-bit integers. And this number should be large
	// enough to ensure we don't run out of space for constants.
	MaxConstants uint32 = 2_147_483_648

	// CSWVersion is the current version of a Romualdo Compiled Storyworld.
	CSWVersion uint32 = 0
)

// CSWMagic is the "magic number" identifying a Romualdo Compiled Storyworld. It
// is comprised of the "RmldCSW" string followed by a SUB character (which in
// times long gone used to represent a "soft end-of-file").
var CSWMagic = []byte{0x52, 0x6D, 0x6C, 0x64, 0x43, 0x53, 0x57, 0x1A}

// CompiledStoryworld is a compiled, binary version of a Romualdo Language
// Storyworld.
//
// TODO: Make it serializable and deserializable. All serialized data shall be
// little endian.
//
// TODO: Use a string interner to avoid having duplicate strings in memory.
// Make some measurements to ensure it's really beneficial.
type CompiledStoryworld struct {
	// The constant values used in all Chunks.
	Constants []Value

	// Chunks is a slide with all Chunks of bytecode containing the compiled
	// data. There is one Chunk for each procedure in the Storyworld.
	//
	// TODO: And in the future, one Chunk for every version of every procedure.
	Chunks []*Chunk

	// FirstChunk indexes the element in Chunks from where the Storyworld
	// execution starts. In other words, it points to the lastest version of the
	// "/main" chunk.
	FirstChunk int
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

//
// romutil.Serializer and romutil.Deserializer interfaces
//

// Serialize serializes the CompiledStoryworld to the given io.Writer.
func (csw *CompiledStoryworld) Serialize(w io.Writer) error {
	err := csw.serializeHeader(w)
	if err != nil {
		return errs.NewCommandFinish("serializing compiled storyworld header: %v", err)
	}

	crc32, err := csw.serializePayload(w)
	if err != nil {
		return errs.NewCommandFinish("serializing compiled storyworld payload: %v", err)
	}

	err = csw.serializeFooter(w, crc32)
	if err != nil {
		return errs.NewCommandFinish("serializing compiled storyworld footer: %v", err)
	}

	return nil
}

// serializedHeader writes the header of a CompiledStoryworld to the given
// io.Writer.
func (csw *CompiledStoryworld) serializeHeader(w io.Writer) error {
	_, err := w.Write(CSWMagic)
	if err != nil {
		return err
	}

	err = romutil.SerializeU32(w, CSWVersion)
	return err
}

// serializePayload writes the payload of a CompiledStoryworld to the given
// io.Writer. In other words, this the function doing the actual serialization.
// Returns the CRC32 of the data written to w, and an error.
func (csw *CompiledStoryworld) serializePayload(w io.Writer) (uint32, error) {
	crc := crc32.NewIEEE()
	mw := io.MultiWriter(w, crc)

	// Constants
	err := romutil.SerializeU32(mw, uint32(len(csw.Constants)))
	if err != nil {
		return 0, err
	}

	for _, v := range csw.Constants {
		err = v.Serialize(mw)
		if err != nil {
			return 0, err
		}
	}

	// Chunks
	err = romutil.SerializeU32(mw, uint32(len(csw.Chunks)))
	if err != nil {
		return 0, err
	}

	for _, chunk := range csw.Chunks {
		err = romutil.SerializeU32(mw, uint32(len(chunk.Code)))
		if err != nil {
			return 0, err
		}
		_, err = mw.Write(chunk.Code)
		if err != nil {
			return 0, err
		}
	}

	// FirstChunk
	err = romutil.SerializeU32(mw, uint32(csw.FirstChunk))
	if err != nil {
		return 0, err
	}

	// Voilà!
	return crc.Sum32(), nil
}

// serializeFooter writes the footer of a CompiledStoryworld to the given
// io.Writer.
func (csw *CompiledStoryworld) serializeFooter(w io.Writer, crc32 uint32) error {
	err := romutil.SerializeU32(w, crc32)
	return err
}

// func (csw *CompiledStoryworld) Deserialize(r io.Reader) error {
// // TODO!
// }
