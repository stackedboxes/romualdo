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

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan the source code and print the tokens",
	Long:  `Scan the source code and print the tokens. This is only useful for testing when developing Romualdo itself.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Scan command not implemented!") // TODO!
	},
}
