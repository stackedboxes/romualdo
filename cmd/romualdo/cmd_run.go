/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2023 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package main

import (
	"github.com/spf13/cobra"
	"github.com/stackedboxes/romualdo/pkg/romutil"
	"github.com/stackedboxes/romualdo/pkg/vm"
)

// runDebugTraceExecution is for the flag --trace.
var runDebugTraceExecution bool

var runCmd = &cobra.Command{
	Use:   "run <ras-file or storyworld-path>",
	Short: "Runs a Storyworld using the VM-based interpreter",
	Long: `Runs a Storyworld using the VM-based interpreter. Can run either a compiled
Storyworld (*.ras) or a Storyworld source directory.`,
	Args: cobra.ExactArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		mouth, ear := romutil.StdMouthAndEar()
		err := vm.RunStoryworld(args[0], mouth, ear, runDebugTraceExecution)
		reportAndExit(err)
	},
}
