/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2023 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package main

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "romualdo",
	Short: "Romualdo is a programming language for Interactive Storytelling",
	Long: `A programming language designed for creating Interactive Storytelling
experiences. Whatever this means. And only for a certain definition
of Interactive Storytelling.`,
}

func init() {
	rootCmd.AddCommand(scanCmd, printASTCmd)
}