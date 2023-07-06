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

// DebugInfo contains debug information matching a CompiledStoryworld. All
// information that is not strictly necessary to run a Storyworld but is useful
// for debugging, producing better error reporting, etc, belongs here.
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

//
// romutil.Serializer and romutil.Deserializer interfaces
//

const (
	// DebugInfoVersion is the current version of a Romualdo DebugInfo.
	DebugInfoVersion uint32 = 0
)

// DebugInfoMagic is the "magic number" identifying a Romualdo DebugInfo. It is
// comprised of the "RmldDbg" string followed by a SUB character (which in times
// long gone used to represent a "soft end-of-file").
var DebugInfoMagic = []byte{0x52, 0x6D, 0x6C, 0x64, 0x44, 0x62, 0x67, 0x1A}

// romutil.Serializer and romutil.Deserializer interfaces
//
// Serialize serializes the DebugInfo to the given io.Writer.
func (di *DebugInfo) Serialize(w io.Writer) errs.Error {
	err := di.serializeHeader(w)
	if err != nil {
		return err
	}

	crc32, err := di.serializePayload(w)
	if err != nil {
		return err
	}

	err = di.serializeFooter(w, crc32)
	return err
}

// serializedHeader writes the header of a DebugInfo to the given io.Writer.
func (di *DebugInfo) serializeHeader(w io.Writer) errs.Error {
	_, plainErr := w.Write(DebugInfoMagic)
	if plainErr != nil {
		return errs.NewRomualdoTool("serializing debug info magic: %v", plainErr)
	}

	err := romutil.SerializeU32(w, DebugInfoVersion)
	return err
}

// serializePayload writes the payload of a CompiledStoryworld to the given
// io.Writer. In other words, this the function doing the actual serialization.
// Returns the CRC32 of the data written to w, and an error.
func (di *DebugInfo) serializePayload(w io.Writer) (uint32, errs.Error) {
	crc := crc32.NewIEEE()
	mw := io.MultiWriter(w, crc)

	// Number of chunks
	err := romutil.SerializeU32(mw, uint32(len(di.ChunksNames)))
	if err != nil {
		return 0, err
	}

	// Chunks Names
	err = romutil.SerializeStringSliceNoLength(mw, di.ChunksNames)
	if err != nil {
		return 0, err
	}

	// Chunks Source Files
	err = romutil.SerializeStringSliceNoLength(mw, di.ChunksSourceFiles)
	if err != nil {
		return 0, err
	}

	// Chunks Lines
	for _, lines := range di.ChunksLines {
		err = romutil.SerializeIntSliceAsU32(mw, lines)
		if err != nil {
			return 0, err
		}
	}

	// Voilà!
	return crc.Sum32(), nil
}

// serializeFooter writes the footer of a CompiledStoryworld to the given
// io.Writer.
func (di *DebugInfo) serializeFooter(w io.Writer, crc32 uint32) errs.Error {
	err := romutil.SerializeU32(w, crc32)
	return err
}

func (di *DebugInfo) Deserialize(r io.Reader) errs.Error {
	err := di.deserializeHeader(r)
	if err != nil {
		return errs.NewRomualdoTool("deserializing debug info header: %v", err)
	}

	crc32, err := di.deserializePayload(r)
	if err != nil {
		return errs.NewRomualdoTool("deserializing debug info payload: %v", err)
	}

	err = di.deserializeFooter(crc32, r)
	if err != nil {
		return errs.NewRomualdoTool("deserializing debug info footer: %v", err)
	}

	return nil
}

// deserializeHeader reads the header of a DebugInfo from the given io.Reader.
// It returns an error if the header is invalid.
func (di *DebugInfo) deserializeHeader(r io.Reader) error {
	readMagic := make([]byte, len(DebugInfoMagic))
	_, err := io.ReadFull(r, readMagic)
	if err != nil {
		return err
	}

	if string(readMagic) != string(DebugInfoMagic) {
		return errors.New("invalid debug info magic number")
	}

	version, err := romutil.DeserializeU32(r)
	if err != nil {
		return err
	}

	if version != DebugInfoVersion {
		return fmt.Errorf("unsupported debug info version: %v", version)
	}

	return nil
}

// deserializePayload reads the payload of a DebugInfo from the given io.Reader.
// Returns the CRC32 of the data read from r, and an error. It updates the
// DebugInfo with the deserialized data as it goes.
func (di *DebugInfo) deserializePayload(r io.Reader) (uint32, error) {
	crcSummer := crc32.NewIEEE()
	tr := io.TeeReader(r, crcSummer)

	// Number of chunks
	chunksCount, err := romutil.DeserializeU32(tr)
	if err != nil {
		return 0, err
	}

	// Chunks Names
	di.ChunksNames, err = romutil.DeserializeStringSliceNoLength(tr, int(chunksCount))
	if err != nil {
		return 0, err
	}

	// Chunks Source Files
	di.ChunksSourceFiles, err = romutil.DeserializeStringSliceNoLength(tr, int(chunksCount))
	if err != nil {
		return 0, err
	}

	// Chunks Lines
	di.ChunksLines = make([][]int, chunksCount)
	for i := range di.ChunksLines {
		di.ChunksLines[i], err = romutil.DeserializeIntSliceAsU32(tr)
		if err != nil {
			return 0, err
		}
	}

	// Voilà!
	return crcSummer.Sum32(), nil
}

// deserializeFooter reads the footer of a DebugInfo from the given io.Reader.
// You must pass the CRC32 of the payload previously read from r.
func (di *DebugInfo) deserializeFooter(crc32 uint32, r io.Reader) error {
	readCRC32, err := romutil.DeserializeU32(r)
	if err != nil {
		return err
	}
	if readCRC32 != crc32 {
		return errors.New("debug info CRC32 mismatch")
	}
	return nil
}
