/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2023 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package main

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/stackedboxes/romualdo/pkg/twi"
)

var walkCmd = &cobra.Command{
	Use:   "walk <path>",
	Short: "Runs the source using the tree-walk interpreter",
	Long:  `Runs the source using the tree-walk interpreter.`,
	Args:  cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		path := args[0]
		return twi.InterpretSource(path, os.Stdout)
	},
}
