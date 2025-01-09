/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2025 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package romutil

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/stackedboxes/romualdo/pkg/ast"
	"github.com/stackedboxes/romualdo/pkg/errs"
)

// SerializeBool writes a Boolean value to the given io.Writer. The Boolean is
// encoded as a single byte, with the value 0 (false) or 1 (true).
func SerializeBool(w io.Writer, b bool) errs.Error {
	bb := []byte{0}
	if b {
		bb[0] = 1
	}
	_, err := w.Write(bb)
	if err != nil {
		return errs.NewRomualdoTool("serializing bool: %v", err)
	}
	return nil
}

// DeserializeBool reads a Boolean from the given io.Reader. It is read as a
// single byte, with value 0 (false) or 1 (true).
func DeserializeBool(r io.Reader) (bool, errs.Error) {
	var bb [1]byte
	_, err := io.ReadFull(r, bb[:])
	if err != nil {
		return false, errs.NewRomualdoTool("deserializing bool: %v", err)
	}
	switch bb[0] {
	case 0:
		return false, nil
	case 1:
		return true, nil
	default:
		return false, errs.NewRomualdoTool("deserializing bool: unexpected value %v", bb[0])
	}
}

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

// SerializeStringSlice writes a []string to a given io.Writer. It first writes
// the slice length (uint32, little endian). Then for each string it writes
// first the length of the string (as in uint32, little endian), then the string
// data itself (UTF-8).
func SerializeStringSlice(w io.Writer, ss []string) errs.Error {
	err := SerializeU32(w, uint32(len(ss)))
	if err != nil {
		return err
	}

	for _, s := range ss {
		err := SerializeString(w, s)
		if err != nil {
			return err
		}
	}
	return nil
}

// DeserializeStringSlice reads a []string from the given io.Reader. It reads
// the number of elements before reading the strings themselves.
func DeserializeStringSlice(r io.Reader) ([]string, errs.Error) {
	length, err := DeserializeU32(r)
	if err != nil {
		return nil, err
	}

	ss := make([]string, length)
	for i := 0; i < int(length); i++ {
		s, err := DeserializeString(r)
		if err != nil {
			return nil, err
		}
		ss[i] = s
	}
	return ss, nil
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

// SerializeType serializes the given type.
func SerializeType(w io.Writer, tag ast.TypeTag) errs.Error {
	b := []byte{0}
	switch tag {
	case ast.TypeVoid:
		b[0] = 0
	case ast.TypeBool:
		b[0] = 1
	case ast.TypeInt:
		b[0] = 2
	case ast.TypeFloat:
		b[0] = 3
	case ast.TypeBNum:
		b[0] = 4
	case ast.TypeString:
		b[0] = 5
	default:
		// An invalid in-computer state means a bug, so we can jus panic here.
		panic(fmt.Sprintf("serializing type: unknown type tag: %v", int(tag)))
	}

	_, err := w.Write(b[:])
	if err != nil {
		return errs.NewRomualdoTool("serializing type: %v", err)
	}
	return nil
}

// DeserializeType deserializes a type from the given io.Reader.
func DeserializeType(r io.Reader) (ast.TypeTag, errs.Error) {
	var b [1]byte
	_, err := io.ReadFull(r, b[:])
	if err != nil {
		return ast.TypeInvalid, errs.NewRomualdoTool("deserializing code hash: %v", err)
	}
	switch b[0] {
	case 0:
		return ast.TypeVoid, nil
	case 1:
		return ast.TypeBool, nil
	case 2:
		return ast.TypeInt, nil
	case 3:
		return ast.TypeFloat, nil
	case 4:
		return ast.TypeBNum, nil
	case 5:
		return ast.TypeString, nil
	default:
		// When deserializing, an invalid value doesn't necessarily means a bug,
		// so we better report it properly.
		return ast.TypeInvalid, errs.NewRomualdoTool("deserializing code hash: invalid type tag: %v", b[0])
	}
}

// SerializeCodeHash serializes the given code hash. This is pretty
// straightforward, it just outputs the 32 bytes sequentially.
func SerializeCodeHash(w io.Writer, hash CodeHash) errs.Error {
	_, err := w.Write(hash[:])
	if err != nil {
		return errs.NewRomualdoTool("serializing code hash: %v", err)
	}
	return nil
}

// DeserializeCodeHash deserializes a code hash from the given io.Reader. The
// expected format is pretty straightforward: a sequence of 32 bytes.
func DeserializeCodeHash(r io.Reader) (CodeHash, errs.Error) {
	var hash CodeHash
	_, err := io.ReadFull(r, hash[:])
	if err != nil {
		return hash, errs.NewRomualdoTool("deserializing code hash: %v", err)
	}

	return hash, nil
}
