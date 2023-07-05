/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2023 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package vm

import (
	"io"
	"os"
	"path"

	"github.com/stackedboxes/romualdo/pkg/backend"
	"github.com/stackedboxes/romualdo/pkg/bytecode"
	"github.com/stackedboxes/romualdo/pkg/errs"
	"github.com/stackedboxes/romualdo/pkg/frontend"
)

// RunStoryworld interprets the Storyworld located at path using the VM-based
// interpreter. Sends output to out.
//
// The path parameter accepts two different things:
//
//  1. A compiled Storyworld (*.ras file). In this case, the file is loaded
//     along with the corresponding debug info (*.rad file) and interpreted.
//  2. A directory containing a Storyworld source. In this case, the source is
//     compiled and interpreted.
//
// trace tells if you want to debug-trace the execution of the VM.
func RunStoryworld(path string, out io.Writer, trace bool) error {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return errs.NewCommandPrep("stating %v: %v", path, err)
	}

	if fileInfo.IsDir() {
		return runStoryworldFromSource(path, out, trace)
	}

	return runStoryworldFromBinary(path, out, trace)
}

func runStoryworldFromSource(path string, out io.Writer, trace bool) error {
	// Parse
	swAST, err := frontend.ParseStoryworld(path)
	if err != nil {
		return errs.NewCommandPrep("parsing the storyworld: %v", err)
	}

	// Generate code
	csw, di, err := backend.GenerateCode(swAST)
	if err != nil {
		return errs.NewCommandPrep("generating code: %v", err)
	}

	// Run
	theVM := New(out)
	theVM.DebugTraceExecution = trace
	return theVM.Interpret(csw, di)
}

func runStoryworldFromBinary(rasFile string, out io.Writer, trace bool) error {
	var csw *bytecode.CompiledStoryworld
	var di *bytecode.DebugInfo

	csw, di, err := LoadCompiledStoryworldBinaries(rasFile, false)
	if err != nil {
		return errs.NewCommandPrep("loading compiled storyworld: %v", err)
	}

	theVM := New(out)
	theVM.DebugTraceExecution = trace
	return theVM.Interpret(csw, di)
}

func LoadCompiledStoryworldBinaries(cswPath string, diRequired bool) (*bytecode.CompiledStoryworld, *bytecode.DebugInfo, error) {
	// Compiled Storyworld itself
	cswFile, err := os.Open(cswPath)
	if err != nil {
		return nil, nil, errs.NewCommandPrep("opening compiled storyworld file %v: %v", cswPath, err)
	}

	csw := &bytecode.CompiledStoryworld{}
	err = csw.Deserialize(cswFile)
	if err != nil {
		return nil, nil, errs.NewCommandPrep("reading the storyworld file %v: %v", cswPath, err)
	}

	// Debug info
	diPath := cswPath[:len(cswPath)-len(path.Ext(cswPath))] + ".rad"
	diFile, err := os.Open(diPath)
	if err != nil {
		if diRequired {
			return nil, nil, errs.NewCommandPrep("opening debug info file %v: %v", diPath, err)
		}
		return csw, nil, nil
	}

	di := &bytecode.DebugInfo{}
	err = di.Deserialize(diFile)
	if diRequired && err != nil {
		return nil, nil, errs.NewCommandPrep("reading debug info from %v: %v", diPath, err)
	}

	return csw, di, nil
}
