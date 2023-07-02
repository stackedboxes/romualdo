/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2023 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package main

import (
	"github.com/spf13/cobra"
	"github.com/stackedboxes/romualdo/pkg/errs"
	"github.com/stackedboxes/romualdo/pkg/vm"
)

// runDebugTraceExecution is for the flag --trace.
var runDebugTraceExecution bool

var runCmd = &cobra.Command{
	Use:   "run <ras-file>",
	Short: "Runs a compiled Storyworld",
	Long:  `Runs a compiled Storyworld.`,
	Args:  cobra.ExactArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		csw, di := loadBinariesExitingOnError(args[0], false)

		theVM := vm.New()
		theVM.DebugTraceExecution = runDebugTraceExecution
		err := theVM.Interpret(csw, di)
		errs.ReportAndExit(err)
	},
}
