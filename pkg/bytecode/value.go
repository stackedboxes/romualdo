/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2024 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package bytecode

import (
	"fmt"
	"io"
	"reflect"

	"github.com/stackedboxes/romualdo/pkg/errs"
	"github.com/stackedboxes/romualdo/pkg/romutil"
)

// A ValueKind represents one of the types a value in the Romualdo Virtual
// Machine can have. This is the type from the perspective of the VM (in the
// sense that user-defined types are obviously not directly represented here).
// We use "kind" in the name because "type" is a keyword in Go.
type ValueKind int

const (
	// ValueBool identifies a Boolean value.
	ValueBool ValueKind = iota

	// ValueString identifies a string value.
	ValueString

	// ValueLecture identifies a Lecture value.
	ValueLecture

	// ValueProcedure identifies a procedure value (either a Passage or a
	// Function).
	ValueProcedure
)

// Procedure is the runtime representation of a Procedure (i.e., a Passage or a
// Function). We don't include any sort of information about return and
// parameter types because type-checking is all done statically at compile-time.
type Procedure struct {
	// ChunkIndex points to the Chunk that contains this function's bytecode.
	// It's an index into the CompiledStoryworld slice of Chunks.
	ChunkIndex int
}

// Lecture is the runtime representation of a Lecture. Lectures are just
// strings, but we wrap them in a struct so that we can differentiate between
// strings and Lectures.
type Lecture struct {
	// Text is the text of the Lecture.
	Text string
}

// TODO: Create wrapper (in the same vein as Lecture) for bnums. (Rationale:
// blend is more expensive than normal float operations, so any cost related
// with unwrapping is better paid by bnums than by normal floats.)

// Value is a Romualdo language value.
type Value struct {
	Value interface{}
}

// NewValueBool creates a new Value of type bool, representing a Boolean with
// the given value.
func NewValueBool(value bool) Value {
	return Value{
		Value: value,
	}
}

// NewValueString creates a new Value of type string, representing a string with
// the given text.
func NewValueString(text string) Value {
	return Value{
		Value: text,
	}
}

// NewValueLecture creates a new Value of type Lecture, representing a
// Lecture with the given text.
func NewValueLecture(text string) Value {
	return Value{
		Value: Lecture{
			Text: text,
		},
	}
}

// NewValueProcedure creates a new Value of type Procedure, representing a
// Procedure that will run the code at the given Chunk index.
func NewValueProcedure(index int) Value {
	return Value{
		Value: Procedure{
			ChunkIndex: index,
		},
	}
}

// AsBool returns this Value's value, assuming it is a Boolean value.
func (v Value) AsBool() bool {
	return v.Value.(bool)
}

// AsString returns this Value's value, assuming it is a string value.
func (v Value) AsString() string {
	return v.Value.(string)
}

// AsLecture returns this Value's value, assuming it is a Lecture value.
func (v Value) AsLecture() Lecture {
	return v.Value.(Lecture)
}

// AsProcedure returns this Value's value, assuming it is a Procedure value.
func (v Value) AsProcedure() Procedure {
	return v.Value.(Procedure)
}

// IsBool checks if the value contains a Boolean value.
func (v Value) IsBool() bool {
	_, ok := v.Value.(bool)
	return ok
}

// IsString checks if the value contains a string value.
func (v Value) IsString() bool {
	_, ok := v.Value.(string)
	return ok
}

// IsLecture checks if the value contains a Lecture value.
func (v Value) IsLecture() bool {
	_, ok := v.Value.(Lecture)
	return ok
}

// IsProcedure checks if the value contains a Procedure value.
func (v Value) IsProcedure() bool {
	_, ok := v.Value.(Procedure)
	return ok
}

// String converts the value to a string. This is also used by the VM to convert
// values to strings, so the output must be user-friendly.
func (v Value) String() string {
	switch vv := v.Value.(type) {
	case bool:
		if vv {
			return "true"
		}
		return "false"

	case string:
		return vv

	case Lecture:
		// There are no variables of type Lecture, so users will never manually
		// convert a Lecture to a string. This will appear in debug traces, but
		// otherwise we don't need to worry about a user-friendly representation
		// here.
		return fmt.Sprintf("<Lecture: %v>", romutil.FormatTextForDisplay(vv.Text))

	case Procedure:
		// TODO: Would be nice to include the function name if we had the debug
		// information around. Hard to access this info from here, though. Could
		// we easily move these string conversions to the VM or whoever has
		// access to the debug info?
		return fmt.Sprintf("<procedure %d>", vv.ChunkIndex)

	default:
		return fmt.Sprintf("<Unexpected type %T>", vv)
	}
}

// DebugString converts the value to a string usable in debug contexts.
// debugInfo can be nil (but this will result in less information in the
// resulting strings).
func (v Value) DebugString(debugInfo *DebugInfo) string {
	switch vv := v.Value.(type) {
	case bool:
		if vv {
			return "true"
		}
		return "false"

	case string:
		return romutil.FormatTextForDisplay(vv)

	case Lecture:
		// There are no variables of type Lecture, so users will never manually
		// convert a Lecture to a string. So, we don't need to worry about a
		// user-friendly representation here.
		return fmt.Sprintf("<Lecture: %v>", romutil.FormatTextForDisplay(vv.Text))

	case Procedure:
		procName := ""
		if debugInfo != nil {
			procName = " (" + debugInfo.ChunksSourceFiles[vv.ChunkIndex] + ")"
		}
		return fmt.Sprintf("<procedure %v%v>", vv.ChunkIndex, procName)

	default:
		return fmt.Sprintf("<Unexpected type %T>", vv)
	}
}

// ValuesEqual checks if a and b are considered equal.
func ValuesEqual(a, b Value) bool {
	if reflect.TypeOf(a.Value) != reflect.TypeOf(b.Value) {
		return false
	}

	switch va := a.Value.(type) {
	case bool:
		return va == b.Value.(bool)

	case string:
		return va == b.Value.(string)

	case Lecture:
		return va.Text == b.Value.(Lecture).Text

	case Procedure:
		return va.ChunkIndex == b.Value.(Procedure).ChunkIndex

	default:
		panic(fmt.Sprintf("Unexpected Value type: %T", va))
	}
}

//
// Serialization and deserialization
//
// Note we don't implement the romutil.Deserializer interface for Values,
// because Values are, well, value types, and this interface is for reference
// types. The spirit is the same, though.
//

// These are the in-disk constants that identify the type of a Romualdo value.
//
// TODO: Need at least a comment explaining why don't need Lecture here (in
// summary, because they are never serialized because there are no Lecture
// variables ever).
const (
	cswBoolFalse byte = 0
	cswBoolTrue  byte = 1
	cswInt       byte = 2
	cswFloat     byte = 3
	cswBNum      byte = 4
	cswString    byte = 5
	cswLecture   byte = 6
)

// Serialize serializes the Value to the given io.Writer.
func (v Value) Serialize(w io.Writer) errs.Error {
	switch vv := v.Value.(type) {
	case bool:
		inDiskValue := cswBoolFalse
		if vv {
			inDiskValue = cswBoolTrue
		}

		bs := []byte{inDiskValue}
		_, plainErr := w.Write(bs)
		if plainErr != nil {
			return errs.NewRomualdoTool("serializing bool: %v", plainErr)
		}
		return nil

	// TODO: When serializing floats, remember to do the proper endianness
	// handling, as shown in https://pkg.go.dev/encoding/binary#example-Write

	case string:
		bs := []byte{cswString}
		_, plainErr := w.Write(bs)
		if plainErr != nil {
			return errs.NewRomualdoTool("serializing string: %v", plainErr)
		}

		err := romutil.SerializeString(w, vv)
		return err

	case Lecture:
		bs := []byte{cswLecture}
		_, plainErr := w.Write(bs)
		if plainErr != nil {
			return errs.NewRomualdoTool("serializing lecture: %v", plainErr)
		}

		err := romutil.SerializeString(w, vv.Text)
		return err

	case Procedure:
		return errs.NewICE("cannot serialize procedure values")

	default:
		// Can't happen
		return errs.NewICE("unexpected value type: %T", vv)
	}
}

// DeserializeValue deserializes a Value from the given io.Reader.
func DeserializeValue(r io.Reader) (Value, errs.Error) {
	v := Value{}
	b := make([]byte, 1)
	_, plainErr := r.Read(b)
	if plainErr != nil {
		return v, errs.NewRomualdoTool("deserializing value: %v", plainErr)
	}

	switch b[0] {
	case cswBoolFalse:
		v.Value = false

	case cswBoolTrue:
		v.Value = true

	case cswString:
		text, err := romutil.DeserializeString(r)
		if err != nil {
			return v, err
		}
		v.Value = text

	case cswLecture:
		text, err := romutil.DeserializeString(r)
		if err != nil {
			return v, err
		}
		v.Value = Lecture{text}

	default:
		// Can happen with corrupted or invalid data
		return v, errs.NewRomualdoTool("unexpected value identifier: %v", b[0])
	}

	return v, nil
}
