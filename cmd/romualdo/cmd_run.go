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

	"github.com/spf13/cobra"
	"github.com/stackedboxes/romualdo/pkg/bytecode"
	"github.com/stackedboxes/romualdo/pkg/errs"
	"github.com/stackedboxes/romualdo/pkg/vm"
)

// runDebugTraceExecution is for the flag --trace.
var runDebugTraceExecution bool

var runCmd = &cobra.Command{
	Use:   "run <ras file>",
	Short: "Runs a compiled Storyworld",
	Long:  `Runs a compiled Storyworld.`,
	Args:  cobra.ExactArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		var csw *bytecode.CompiledStoryworld
		var di *bytecode.DebugInfo

		cswPath := args[0]
		diPath := cswPath[:len(cswPath)-len(path.Ext(cswPath))] + ".rad"

		csw, err := loadCompiledStoryworld(cswPath)
		if err != nil {
			errs.ReportAndExit(err)
		}

		// TODO: It's fine not to have debugInfo. Always ignore the error for
		// now, but in the future, we may want a flag to control this.
		di, _ = loadDebugInfo(diPath)

		theVM := vm.New()
		theVM.DebugTraceExecution = runDebugTraceExecution
		err = theVM.Interpret(csw, di)
		errs.ReportAndExit(err)
	},
}

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
