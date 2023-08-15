/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2023 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package vm

import (
	"os"
	"path"

	"github.com/stackedboxes/romualdo/pkg/backend"
	"github.com/stackedboxes/romualdo/pkg/bytecode"
	"github.com/stackedboxes/romualdo/pkg/errs"
	"github.com/stackedboxes/romualdo/pkg/frontend"
	"github.com/stackedboxes/romualdo/pkg/romutil"
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
func RunStoryworld(path string, mouth romutil.Mouth, ear romutil.Ear, trace bool) errs.Error {
	csw, di, err := cswFromPath(path)
	if err != nil {
		return err
	}

	return runCSW(csw, di, mouth, ear, trace)
}

// cswFromPath loads the CompiledStoryworld and DebugInfo from the given path,
// which can be either a compiled Storyworld (.ras) file or a directory with the
// Storyworld source code.
func cswFromPath(path string) (*bytecode.CompiledStoryworld, *bytecode.DebugInfo, errs.Error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, nil, errs.NewRomualdoTool("stating %v: %v", path, err)
	}

	if fileInfo.IsDir() {
		return cswFromSource(path)
	}

	return cswFromFile(path)
}

// cswFromSource compiles the Storyworld source located at path and returns the
// CompiledStoryworld and DebugInfo.
func cswFromSource(path string) (*bytecode.CompiledStoryworld, *bytecode.DebugInfo, errs.Error) {
	// Parse
	swAST, err := frontend.ParseStoryworld(path)
	if err != nil {
		return nil, nil, err
	}

	// Generate code
	return backend.GenerateCode(swAST)
}

// cswFromFile loads the CompiledStoryworld and DebugInfo from the given
// compiled Storyworld (.ras) file.
func cswFromFile(path string) (*bytecode.CompiledStoryworld, *bytecode.DebugInfo, errs.Error) {
	csw, di, err := LoadCompiledStoryworldBinaries(path, false)
	if err != nil {
		return nil, nil, err
	}
	return csw, di, nil
}

// runCSW interprets the given CompiledStoryworld and (potentially nil)
// DebugInfo.
func runCSW(csw *bytecode.CompiledStoryworld, di *bytecode.DebugInfo, out romutil.Mouth, in romutil.Ear, trace bool) errs.Error {
	theVM := New(out, in)
	theVM.DebugTraceExecution = trace
	return theVM.Interpret(csw, di)
}

// LoadCompiledStoryworldBinaries loads the CompiledStoryworld from cwPath. It
// also looks for the corresponding DebugInfo file and loads it if found. If the
// DebugInfo file is not found, it returns an error only if diRequired is true.
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
