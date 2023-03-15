/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2023 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var printASTCmd = &cobra.Command{
	Use:   "print-ast",
	Short: "Parse the source code, print the AST",
	Long:  `Parse the source code, print the AST. AST stands for "Abstract Syntax Tree", and if you want to see it, that's your command.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Print AST command not implemented!") // TODO!
	},
}
