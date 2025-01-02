/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2025 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package romutil

import (
	"encoding/binary"
	"io"

	"github.com/stackedboxes/romualdo/pkg/errs"
)

// SerializeU32 writes a uint32 to the given io.Writer, in little endian format.
func SerializeU32(w io.Writer, v uint32) errs.Error {
	var u32 [4]byte
	binary.LittleEndian.PutUint32(u32[:], v)
	_, err := w.Write(u32[:])
	if err != nil {
		return errs.NewRomualdoTool("serializing uint32: %v", err)
	}
	return nil
}

// DeserializeU32 reads a uint32 from the given io.Reader, in little endian
// format.
func DeserializeU32(r io.Reader) (uint32, errs.Error) {
	var u32 [4]byte
	_, err := io.ReadFull(r, u32[:])
	if err != nil {
		return 0, errs.NewRomualdoTool("deserializing uint32: %v", err)
	}
	return binary.LittleEndian.Uint32(u32[:]), nil
}

// SerializeI32 writes an int32 to the given io.Writer, in little endian format,
// two's complement.
func SerializeI32(w io.Writer, v int32) errs.Error {
	var u32 [4]byte
	binary.LittleEndian.PutUint32(u32[:], uint32(v))
	_, err := w.Write(u32[:])
	if err != nil {
		return errs.NewRomualdoTool("serializing int32: %v", err)
	}
	return nil
}

// DeserializeI32 reads an int32 from the given io.Reader, in little endian
// format, two's complement.
func DeserializeI32(r io.Reader) (int32, errs.Error) {
	var u32 [4]byte
	_, err := io.ReadFull(r, u32[:])
	if err != nil {
		return 0, errs.NewRomualdoTool("deserializing int32: %v", err)
	}
	return int32(binary.LittleEndian.Uint32(u32[:])), nil
}

// SerializeString writes a string to the given io.Writer. It first writes the
// length of the string (as in uint32, little endian), then the string data
// itself (UTF-8).
func SerializeString(w io.Writer, s string) errs.Error {
	err := SerializeU32(w, uint32(len(s)))
	if err != nil {
		return err
	}

	_, plainErr := io.WriteString(w, s)
	if plainErr != nil {
		return errs.NewRomualdoTool("serializing string: %v", plainErr)
	}
	return nil
}

// DeserializeString reads a string from the given io.Reader. It first reads the
// length of the string (as in uint32, little endian), then the string data
// itself (UTF-8).
func DeserializeString(r io.Reader) (string, errs.Error) {
	length, err := DeserializeU32(r)
	if err != nil {
		return "", err
	}

	buf := make([]byte, length)
	_, plainErr := io.ReadFull(r, buf)
	if plainErr != nil {
		return "", errs.NewRomualdoTool("deserializing string: %v", plainErr)
	}
	return string(buf), nil
}

// SerializeStringSliceNoLength writes a []string to a given io.Writer. For each
// string it writes first the length of the string (as in uint32, little
// endian), then the string data itself (UTF-8). The length of the slice is not
// written.
func SerializeStringSliceNoLength(w io.Writer, ss []string) errs.Error {
	for _, s := range ss {
		err := SerializeString(w, s)
		if err != nil {
			return err
		}
	}
	return nil
}

// DeserializeStringSliceNoLength reads a []string from the given io.Reader. The
// slice length must be provided.
func DeserializeStringSliceNoLength(r io.Reader, length int) ([]string, errs.Error) {
	ss := make([]string, length)
	for i := 0; i < length; i++ {
		s, err := DeserializeString(r)
		if err != nil {
			return nil, err
		}
		ss[i] = s
	}
	return ss, nil
}

// SerializeIntSliceAsU32 writes a []int to a given io.Writer. It first writes
// an uint32 with the slice length, then each of the ints in the slice. All
// numbers are written as uint32, little endian (so beware of overflows and
// negative numbers!)
func SerializeIntSliceAsU32(w io.Writer, ii []int) errs.Error {
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

// DeserializeIntSliceAsU32 reads a []int from the given io.Reader. It first
// reads an uint32 with the slice length, then each of the uint32s in the slice.
// Even though the return type is []int, the numbers are all read as uint32,
// little endian (so beware of overflows).
func DeserializeIntSliceAsU32(r io.Reader) ([]int, errs.Error) {
	length, err := DeserializeU32(r)
	if err != nil {
		return nil, err
	}

	ii := make([]int, length)
	for i := uint32(0); i < length; i++ {
		u32, err := DeserializeU32(r)
		if err != nil {
			return nil, err
		}
		ii[i] = int(u32)
	}
	return ii, nil
}
