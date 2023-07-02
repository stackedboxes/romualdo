/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2023 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package main

import (
	"os"
	"path"

	"github.com/stackedboxes/romualdo/pkg/bytecode"
	"github.com/stackedboxes/romualdo/pkg/errs"
)

// loadBinaries loads the compiled Storyworld and the debug info from files,
// exiting the program properly in case of errors. The rasFile argument is
// desired compiled Storyworld (*.ras) to load; the debug info file name is
// inferred from there: it's the same name, but with the extension changed to
// .rad. If diRequired is true, an error is reported if the debug info file is
// not found or fails to load for whatever other reason.
func loadBinariesExitingOnError(rasFile string, diRequired bool) (*bytecode.CompiledStoryworld, *bytecode.DebugInfo) {
	var csw *bytecode.CompiledStoryworld
	var di *bytecode.DebugInfo

	diPath := rasFile[:len(rasFile)-len(path.Ext(rasFile))] + ".rad"

	csw, err := loadCompiledStoryworld(rasFile)
	if err != nil {
		errs.ReportAndExit(err)
	}

	di, err = loadDebugInfo(diPath)

	if err != nil && diRequired {
		errs.ReportAndExit(err)
	}

	return csw, di
}

// loadCompiledStoryworld loads a compiled Storyworld from a given file.
func loadCompiledStoryworld(cswPath string) (*bytecode.CompiledStoryworld, error) {
	cswFile, err := os.Open(cswPath)
	if err != nil {
		err := errs.NewCommandPrep("could not open compiled storyworld file %v: %v", cswPath, err)
		errs.ReportAndExit(err)
	}

	csw := &bytecode.CompiledStoryworld{}
	err = csw.Deserialize(cswFile)
	if err != nil {
		err := errs.NewCommandPrep("error reading the storyworld file %v: %v", cswPath, err)
		errs.ReportAndExit(err)
	}

	return csw, nil
}

// loadDebugInfo loads debug info from a given file.
func loadDebugInfo(diPath string) (*bytecode.DebugInfo, error) {
	diFile, err := os.Open(diPath)
	if err != nil {
		err := errs.NewCommandPrep("could not open debug info file %v: %v", diPath, err)
		return nil, err
	}

	di := &bytecode.DebugInfo{}
	err = di.Deserialize(diFile)
	if err != nil {
		err := errs.NewCommandPrep("error reading the debug info from %v: %v", diPath, err)
		return nil, err
	}

	return di, nil
}
