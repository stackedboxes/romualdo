/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2023 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package vm

import (
	"bufio"
	"fmt"
	"os"
	"path"

	"github.com/stackedboxes/romualdo/pkg/backend"
	"github.com/stackedboxes/romualdo/pkg/bytecode"
	"github.com/stackedboxes/romualdo/pkg/errs"
	"github.com/stackedboxes/romualdo/pkg/frontend"
)

// CSWFromPath loads the CompiledStoryworld and DebugInfo from the given path,
// which can be either a compiled Storyworld (*.ras) file or a directory with
// the Storyworld source code (*.ral).
func CSWFromPath(path string) (*bytecode.CompiledStoryworld, *bytecode.DebugInfo, errs.Error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, nil, errs.NewRomualdoTool("stating %v: %v", path, err)
	}

	if fileInfo.IsDir() {
		return cswFromSource(path)
	}

	// TODO: This is a bit pointless. Could call LoadCompiledStoryworld
	// directly.
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

// RunCSW interprets the given CompiledStoryworld and (potentially nil)
// DebugInfo. If trace is true, it prints a trace/disassembly of the execution
// to stdout as it goes.
func RunCSW(csw *bytecode.CompiledStoryworld, di *bytecode.DebugInfo, trace bool) (err errs.Error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(*errs.Runtime); ok {
				err = e
				return
			}
			if e, ok := r.(error); ok {
				err = errs.NewICE("Unexpected error: %T (%v)", r, e)
				return
			}
			err = errs.NewICE("Unexpected error type: %T (%v)", r, r)
			return
		}
	}()

	theVM := New(csw, di)
	theVM.DebugTraceExecution = trace

	out := theVM.Start()
	for {
		fmt.Print(out)

		if theVM.State == StateEndOfStory {
			fmt.Println("-- The End --")
			return nil
		}

		if theVM.State != StateWaitingForInput {
			// TODO: Get rid of this assert-like check? Or at least make it
			// "throw" an errs.Runtime.
			panic("Should be waiting for input, right?")
		}

		fmt.Println(theVM.Options)
		fmt.Print("> ")

		s := bufio.NewScanner(os.Stdin)
		s.Scan()
		input := s.Text()

		out = theVM.Step(input)
	}
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
