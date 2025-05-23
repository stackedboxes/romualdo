/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2025 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package main

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:          "romualdo",
	SilenceUsage: true,
	Short:        "Romualdo is a programming language for Interactive Storytelling",
	Long: `A programming language designed for creating Interactive Storytelling
experiences. Whatever this means. And only for a certain definition
of Interactive Storytelling.`,
}

func init() {
	devCmd.AddCommand(devScanCmd, devPrintASTCmd, devTestCmd, devDisassembleCmd, devHashCmd)
	rootCmd.AddCommand(buildCmd, runCmd, devCmd)

	runCmd.Flags().BoolVarP(&runDebugTraceExecution, "trace", "t", false, "debug trace execution")
}
