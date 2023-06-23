/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2023 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package romutil

import (
	"encoding/binary"
	"io"
)

// Serializer is the interface implemented by objects that can serialize
// themselves.
type Serializer interface {
	// Serialize serializes the given object writing the serialized data to w.
	Serialize(w io.Writer) error
}

// Deserializer is the interface implemented by objects that can deserialize
// themselves.
type Deserializer interface {
	// Deserialize deserializes the given object reading the serialized data
	// from r.
	Deserialize(r io.Reader) error
}

// SerializeU32 writes a uint32 to the given io.Write, in little endian format.
func SerializeU32(w io.Writer, v uint32) error {
	var u32 [4]byte
	binary.LittleEndian.PutUint32(u32[:], v)
	_, err := w.Write(u32[:])
	return err
}
