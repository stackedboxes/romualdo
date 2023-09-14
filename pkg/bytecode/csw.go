/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2023 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package bytecode

import (
	"errors"
	"fmt"
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
// TODO: Use a string interner to avoid having duplicate strings in memory.
// Make some measurements to ensure it's really beneficial.
type CompiledStoryworld struct {
	// Swid is the Storyworld ID. It is a free string that should uniquely
	// identify the Storyworld.
	Swid string

	// Swov is the Storyworld version. A negative value means that this is a
	// development version. A non-negative value means that this is a release
	// version. For example, if the currently released version of a Storyworld
	// is 3, the development version will be -4. When the development version
	// becomes stable, it will be released as version 4.
	//
	// A version equals to zero is invalid.
	Swov int32

	// The constant values used in all Chunks.
	Constants []Value

	// Chunks is a slice with all Chunks of bytecode containing the compiled
	// data. There is one Chunk for each procedure in the Storyworld.
	//
	// TODO: And in the future, one Chunk for every version of every procedure.
	Chunks []*Chunk

	// InitialChunk indexes the element in Chunks from where the Storyworld
	// execution starts. In other words, it points to the latest version of the
	// "/main" chunk.
	InitialChunk int
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
func (csw *CompiledStoryworld) Serialize(w io.Writer) errs.Error {
	err := csw.serializeHeader(w)
	if err != nil {
		return err
	}

	crc32, err := csw.serializePayload(w)
	if err != nil {
		return err
	}

	err = csw.serializeFooter(w, crc32)
	return err
}

// serializedHeader writes the header of a CompiledStoryworld to the given
// io.Writer.
func (csw *CompiledStoryworld) serializeHeader(w io.Writer) errs.Error {
	_, plainErr := w.Write(CSWMagic)
	if plainErr != nil {
		return errs.NewRomualdoTool("serializing compiled storyworld magic: %v", plainErr)
	}

	err := romutil.SerializeU32(w, CSWVersion)
	return err
}

// serializePayload writes the payload of a CompiledStoryworld to the given
// io.Writer. In other words, this the function doing the actual serialization.
// Returns the CRC32 of the data written to w, and an error.
func (csw *CompiledStoryworld) serializePayload(w io.Writer) (uint32, errs.Error) {
	crc := crc32.NewIEEE()
	mw := io.MultiWriter(w, crc)

	// Swid and swov
	err := romutil.SerializeString(mw, csw.Swid)
	if err != nil {
		return 0, err
	}

	err = romutil.SerializeI32(mw, csw.Swov)
	if err != nil {
		return 0, err
	}

	// Constants
	err = romutil.SerializeU32(mw, uint32(len(csw.Constants)))
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
		_, plainErr := mw.Write(chunk.Code)
		if plainErr != nil {
			return 0, errs.NewRomualdoTool("serializing chunk code: %v", plainErr)
		}
	}

	// InitialChunk
	err = romutil.SerializeU32(mw, uint32(csw.InitialChunk))
	if err != nil {
		return 0, err
	}

	// Voilà!
	return crc.Sum32(), nil
}

// serializeFooter writes the footer of a CompiledStoryworld to the given
// io.Writer.
func (csw *CompiledStoryworld) serializeFooter(w io.Writer, crc32 uint32) errs.Error {
	err := romutil.SerializeU32(w, crc32)
	return err
}

// Deserialize deserializes a CompiledStoryworld from the given io.Reader.
func (csw *CompiledStoryworld) Deserialize(r io.Reader) error {
	err := csw.deserializeHeader(r)
	if err != nil {
		return err
	}

	crc32, err := csw.deserializePayload(r)
	if err != nil {
		return err
	}

	err = csw.deserializeFooter(r, crc32)
	return err
}

// deserializeHeader reads and checks the header of a CompiledStoryworld from
// the given io.Reader. If everything is OK, it returns nil, otherwise it
// returns an error.
func (csw *CompiledStoryworld) deserializeHeader(r io.Reader) error {
	// Magic
	readMagic := make([]byte, len(CSWMagic))
	_, err := io.ReadFull(r, readMagic)
	if err != nil {
		return err
	}
	for i, b := range readMagic {
		if b != CSWMagic[i] {
			// TODO: Could be friendlier here, by comparing readMagic with other
			// Romualdo magic numbers and reporting a more meaningful error.
			return errors.New("invalid compiled storyworld magic number")
		}
	}

	// Version
	readVersion, err := romutil.DeserializeU32(r)
	if err != nil {
		return err
	}
	if readVersion != CSWVersion {
		return fmt.Errorf("unsupported compiled storyworld version: %v", readVersion)
	}

	// Header is OK
	return nil
}

// deserializePayload reads the payload of a CompiledStoryworld from the given
// io.Reader. In other words, this the function doing the actual
// deserialization. Returns the CRC32 of the data read from r, and an error. It
// updates the CompiledStoryworld with the deserialized data as it goes.
func (csw *CompiledStoryworld) deserializePayload(r io.Reader) (uint32, errs.Error) {

	crcSummer := crc32.NewIEEE()
	tr := io.TeeReader(r, crcSummer)

	// Swid and swov
	swid, err := romutil.DeserializeString(tr)
	if err != nil {
		return 0, err
	}
	csw.Swid = swid

	swov, err := romutil.DeserializeI32(tr)
	if err != nil {
		return 0, err
	}
	csw.Swov = swov

	// Constants
	lenConstants, err := romutil.DeserializeU32(tr)
	if err != nil {
		return 0, err
	}

	csw.Constants = make([]Value, lenConstants)

	for i := range csw.Constants {
		csw.Constants[i], err = DeserializeValue(tr)
		if err != nil {
			return 0, err
		}
	}

	// Chunks
	lenChunks, err := romutil.DeserializeU32(tr)
	if err != nil {
		return 0, err
	}
	csw.Chunks = make([]*Chunk, lenChunks)
	for i := range csw.Chunks {
		lenChunkCode, err := romutil.DeserializeU32(tr)
		if err != nil {
			return 0, err
		}
		csw.Chunks[i] = &Chunk{
			Code: make([]byte, lenChunkCode),
		}
		_, plainErr := io.ReadFull(tr, csw.Chunks[i].Code)
		if plainErr != nil {
			return 0, errs.NewRomualdoTool("deserializing chunk code: %v", plainErr)
		}
	}

	// InitialChunk
	i32, err := romutil.DeserializeU32(tr)
	if err != nil {
		return 0, err
	}
	csw.InitialChunk = int(i32)

	// Voilà!
	return crcSummer.Sum32(), nil
}

// deserializeFooter reads and checks the footer of a CompiledStoryworld from
// the given io.Reader. You must pass the CRC32 of the payload previously read
// from r.
func (csw *CompiledStoryworld) deserializeFooter(r io.Reader, crc32 uint32) error {
	readCRC32, err := romutil.DeserializeU32(r)
	if err != nil {
		return err
	}
	if readCRC32 != crc32 {
		return errors.New("compiled storyworld CRC32 mismatch")
	}
	return nil
}
