/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2024 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package vm

import (
	"hash/crc32"
	"io"

	"github.com/stackedboxes/romualdo/pkg/errs"
	"github.com/stackedboxes/romualdo/pkg/romutil"
)

const (
	// savedStateVersion is the current version of a Romualdo saved state.
	savedStateVersion uint32 = 0
)

// savedStateMagic is the "magic number" identifying a Romualdo VM saved state.
// It is comprised of the "RmldSav" string followed by a SUB character (which in
// times long gone used to represent a "soft end-of-file").
var savedStateMagic = []byte{0x52, 0x6D, 0x6C, 0x64, 0x53, 0x61, 0x76, 0x1A}

// Serialize serializes the VM state to the given io.Writer.
func (vm *VM) Serialize(w io.Writer) errs.Error {
	err := vm.serializeHeader(w)
	if err != nil {
		return err
	}

	crc32, err := vm.serializePayload(w)
	if err != nil {
		return err
	}

	err = vm.serializeFooter(w, crc32)
	return err
}

// serializedHeader writes the header of a VM saved state to the given
// io.Writer.
func (vm *VM) serializeHeader(w io.Writer) errs.Error {
	_, plainErr := w.Write(savedStateMagic)
	if plainErr != nil {
		return errs.NewRomualdoTool("serializing VM state magic: %v", plainErr)
	}

	err := romutil.SerializeU32(w, savedStateVersion)
	return err
}

// serializePayload writes the payload of a VM saved state to the given
// io.Writer. In other words, this the function doing the actual serialization.
// Returns the CRC32 of the data written to w, and an error.
func (vm *VM) serializePayload(w io.Writer) (uint32, errs.Error) {
	crc := crc32.NewIEEE()
	mw := io.MultiWriter(w, crc)

	// VM State
	err := romutil.SerializeU32(mw, uint32(vm.State))
	if err != nil {
		return 0, err
	}

	// Options
	err = romutil.SerializeString(mw, vm.Options)
	if err != nil {
		return 0, err
	}

	// Stack
	err = vm.stack.Serialize(mw)
	if err != nil {
		return 0, err
	}

	// Frames
	err = romutil.SerializeU32(mw, uint32(len(vm.frames)))
	if err != nil {
		return 0, err
	}
	for _, f := range vm.frames {
		err = f.Serialize(mw)
		if err != nil {
			return 0, err
		}
	}

	// Voilà!
	return crc.Sum32(), nil
}

// serializeFooter writes the footer of a VM saved state to the given io.Writer.
func (vm *VM) serializeFooter(w io.Writer, crc32 uint32) errs.Error {
	err := romutil.SerializeU32(w, crc32)
	return err
}

// Deserialize deserializes a VM state from the given io.Reader.
func (vm *VM) Deserialize(r io.Reader) errs.Error {
	err := vm.deserializeHeader(r)
	if err != nil {
		return err
	}

	crc32, err := vm.deserializePayload(r)
	if err != nil {
		return err
	}

	err = vm.deserializeFooter(r, crc32)
	if err != nil {
		return err
	}

	// Post-deserialization adjustments
	vm.frame = nil
	if len(vm.frames) > 0 {
		vm.frame = vm.frames[len(vm.frames)-1]
	}
	vm.outBuffer.Reset()

	return nil

}

// deserializeHeader reads and checks the header of a VM saved state from the
// given io.Reader. If everything is OK, it returns nil, otherwise it returns an
// error.
func (vm *VM) deserializeHeader(r io.Reader) errs.Error {
	// Magic
	readMagic := make([]byte, len(savedStateMagic))
	_, err := io.ReadFull(r, readMagic)
	if err != nil {
		// TODO: This isn't really a Romualdo tool error. But also not really a
		// Runtime error. And I am not sure I want to create a brand new error
		// type just for this. Think about it. And the same for other cases in
		// this file. *And*, all generic serialization stuff in romutil that
		// gets used here indirectly! (Maybe those should return normal Go
		// errors? So that callers need to add more context in an errs.Error?)
		return errs.NewRomualdoTool("deserializing VM state header magic: %v", err)
	}
	for i, b := range readMagic {
		if b != savedStateMagic[i] {
			// TODO: Could be friendlier here, by comparing readMagic with other
			// Romualdo magic numbers and reporting a more meaningful error.
			return errs.NewRomualdoTool("invalid VM state magic number")
		}
	}

	// Version
	readVersion, err := romutil.DeserializeU32(r)
	if err != nil {
		return errs.NewRomualdoTool("deserializing VM state header version: %v", err)
	}
	if readVersion != savedStateVersion {
		return errs.NewRomualdoTool("unsupported VM state version: %v", readVersion)
	}

	// Header is OK
	return nil
}

// deserializePayload reads the payload of a VM saved state from the given
// io.Reader. In other words, this the function doing the actual
// deserialization. Returns the CRC32 of the data read from r, and an error. It
// updates the VM state with the deserialized data as it goes.
func (vm *VM) deserializePayload(r io.Reader) (uint32, errs.Error) {
	crcSummer := crc32.NewIEEE()
	tr := io.TeeReader(r, crcSummer)

	// TODO: Check for compatibility between the saved state and the Storyworld
	// loaded into the VM.

	// VM State
	vmState, err := romutil.DeserializeU32(tr)
	if err != nil {
		return 0, err
	}
	vm.State = State(vmState)

	// Options
	options, err := romutil.DeserializeString(tr)
	if err != nil {
		return 0, err
	}
	vm.Options = options

	// Stack
	stack, err := DeserializeStack(tr)
	if err != nil {
		return 0, err
	}
	vm.stack = stack

	// Frames
	frameCount, err := romutil.DeserializeU32(tr)
	if err != nil {
		return 0, err
	}

	frames := make([]*callFrame, frameCount)
	for i := 0; i < int(frameCount); i++ {
		frame, err := DeserializeCallFrame(tr, vm.stack)
		if err != nil {
			return 0, err
		}
		frames = append(frames, frame)
	}
	vm.frames = frames

	// Voilà!
	return crcSummer.Sum32(), nil
}

// deserializeFooter reads and checks the footer of a CompiledStoryworld from
// the given io.Reader. You must pass the CRC32 of the payload previously read
// from r.
func (vm *VM) deserializeFooter(r io.Reader, crc32 uint32) errs.Error {
	readCRC32, err := romutil.DeserializeU32(r)
	if err != nil {
		return err
	}
	if readCRC32 != crc32 {
		return errs.NewRomualdoTool("compiled storyworld CRC32 mismatch")
	}
	return nil
}
