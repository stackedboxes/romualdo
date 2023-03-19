/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2023 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package main

import (
	"errors"
	"os"

	"github.com/spf13/cobra"
	"github.com/stackedboxes/romualdo/pkg/frontend"
	"github.com/stackedboxes/romualdo/pkg/twi"
)

var walkCmd = &cobra.Command{
	Use:   "walk <path>",
	Short: "Runs the source using the tree-walk interpreter",
	Long:  `Runs the source using the tree-walk interpreter.`,
	Args:  cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: I think this preamble tends to repeat itself in different
		// commands. Factor it out!
		path := args[0]
		source, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		ast := frontend.Parse(string(source))
		if ast == nil {
			return errors.New("Parsing error.")
		}

		// TODO: This is looking messy. We probably shouldn't be instantiating
		// and using the visitor ourselves here.
		gsv := twi.NewGlobalsSymbolVisitor()
		ast.Walk(gsv)
		procedures := gsv.Procedures()

		return twi.Interpret(ast, procedures)
	},
}
