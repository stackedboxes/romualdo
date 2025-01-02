/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2025 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package main

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/stackedboxes/romualdo/pkg/errs"
	"github.com/stackedboxes/romualdo/pkg/frontend"
	"github.com/stackedboxes/romualdo/pkg/romutil"
)

var devHashCmd = &cobra.Command{
	Use:   "hash <path>",
	Short: "Computes the code hash of procedures and/or globals",
	Long: `Computes the code hash of procedures and/or globals. If you pass the
name of a procedure or global variable via the optional --symbol flag, the
command will print the hash of the requested symbol only. Otherwise, it will print
the hash of all symbols.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]
		ast, err := frontend.ParseFile(path, filepath.Dir(path))
		reportAndExitOnError(err)

		ch := romutil.NewCodeHasher()
		ast.Walk(ch)

		// Did the user ask for a specific symbol?
		if flagDevHashSymbol != "" {
			if hash, ok := ch.ProcedureHashes[flagDevHashSymbol]; ok {
				fmt.Printf("%x  %v\n", hash, flagDevHashSymbol)
				return
			}
			if hash, ok := ch.GlobalHashes[flagDevHashSymbol]; ok {
				fmt.Printf("%x  %v\n", hash, flagDevHashSymbol)
				return
			}

			reportAndExitOnError(errs.NewRomualdoTool("Symbol not found: %v", flagDevHashSymbol))
		}

		// Nope, print hashes for all symbols.
		for sym, hash := range ch.ProcedureHashes {
			fmt.Printf("%x  %v\n", hash, sym)
		}
		for sym, hash := range ch.GlobalHashes {
			fmt.Printf("%x  %v\n", hash, sym)
		}
	},
}

// flagDevHashSymbol is the value of the --symbol flag of the `dev hash`
// command.
var flagDevHashSymbol string

func init() {
	devHashCmd.Flags().StringVarP(&flagDevHashSymbol, "symbol", "s",
		"", "Fully-qualified name of the desired procedure or global variable.")
}
