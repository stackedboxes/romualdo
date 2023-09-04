/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2023 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package vm

import (
	"io"

	"github.com/stackedboxes/romualdo/pkg/errs"
)

const (
	// savedStateVersion is the current version of a Romualdo saved state.
	savedStateVersion uint32 = 0
)

// savedStateMagic is the "magic number" identifying a Romualdo VM saved state.
// It is comprised of the "RmldSav" string followed by a SUB character (which in
// times long gone used to represent a "soft end-of-file").
var savedStateMagic = []byte{0x52, 0x6D, 0x6C, 0x64, 0x53, 0x62, 0x76, 0x1A}

// xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
func (vm *VM) SaveState(dst io.Writer) errs.Error {
	return &errs.Runtime{Message: "Not implemented"}
}

// xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
func (vm *VM) LoadState(src io.Reader) errs.Error {
	return &errs.Runtime{Message: "Not implemented"}
}
