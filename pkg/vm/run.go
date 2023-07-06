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
func RunStoryworld(path string, out io.Writer, trace bool) errs.Error {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return errs.NewRomualdoTool("stating %v: %v", path, err)
	}

	if fileInfo.IsDir() {
		return runStoryworldFromSource(path, out, trace)
	}

	return runStoryworldFromBinary(path, out, trace)
}

func runStoryworldFromSource(path string, out io.Writer, trace bool) errs.Error {
	// Parse
	swAST, err := frontend.ParseStoryworld(path)
	if err != nil {
		return err
	}

	// Generate code
	csw, di, err := backend.GenerateCode(swAST)
	if err != nil {
		return err
	}

	// Run
	theVM := New(out)
	theVM.DebugTraceExecution = trace
	return theVM.Interpret(csw, di)
}

func runStoryworldFromBinary(rasFile string, out io.Writer, trace bool) errs.Error {
	var csw *bytecode.CompiledStoryworld
	var di *bytecode.DebugInfo

	csw, di, err := LoadCompiledStoryworldBinaries(rasFile, false)
	if err != nil {
		return err
	}

	theVM := New(out)
	theVM.DebugTraceExecution = trace
	return theVM.Interpret(csw, di)
}

func LoadCompiledStoryworldBinaries(cswPath string, diRequired bool) (*bytecode.CompiledStoryworld, *bytecode.DebugInfo, errs.Error) {
	// Compiled Storyworld itself
	cswFile, err := os.Open(cswPath)
	if err != nil {
		return nil, nil, errs.NewRomualdoTool("opening compiled storyworld file %v: %v", cswPath, err)
	}

	csw := &bytecode.CompiledStoryworld{}
	err = csw.Deserialize(cswFile)
	if err != nil {
		return nil, nil, errs.NewRomualdoTool("reading the storyworld file %v: %v", cswPath, err)
	}

	// Debug info
	diPath := cswPath[:len(cswPath)-len(path.Ext(cswPath))] + ".rad"
	diFile, err := os.Open(diPath)
	if err != nil {
		if diRequired {
			return nil, nil, errs.NewRomualdoTool("opening debug info file %v: %v", diPath, err)
		}
		return csw, nil, nil
	}

	di := &bytecode.DebugInfo{}
	err = di.Deserialize(diFile)
	if diRequired && err != nil {
		return nil, nil, errs.NewRomualdoTool("reading debug info from %v: %v", diPath, err)
	}

	return csw, di, nil
}
