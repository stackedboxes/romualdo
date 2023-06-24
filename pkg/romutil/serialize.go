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

// SerializeU32 writes a uint32 to the given io.Writer, in little endian format.
func SerializeU32(w io.Writer, v uint32) error {
	var u32 [4]byte
	binary.LittleEndian.PutUint32(u32[:], v)
	_, err := w.Write(u32[:])
	return err
}

// SerializeStringSliceNoLength writes a []string to a given io.Writer. For each
// string it writes first the length of the string (as in uint32, little
// endian), then the string data itself (UTF-8). The length of the slice is not
// written.
func SerializeStringSliceNoLength(w io.Writer, ss []string) error {
	for _, s := range ss {
		err := SerializeU32(w, uint32(len(s)))
		if err != nil {
			return err
		}

		_, err = io.WriteString(w, s)
		if err != nil {
			return err
		}
	}
	return nil
}

// SerializeIntSlice writes a []int to a given io.Writer. It first writes an
// uint32 with the slice length, then each of the ints in the slice. All numbers
// are written as uint32, little endian (so beware of overflows and negative
// numbers!)
func SerializeIntSliceAsU32(w io.Writer, ii []int) error {
	err := SerializeU32(w, uint32(len(ii)))
	if err != nil {
		return err
	}

	for _, i := range ii {
		err = SerializeU32(w, uint32(i))
		if err != nil {
			return err
		}
	}
	return nil
}
