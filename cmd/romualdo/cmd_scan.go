/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2024 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/stackedboxes/romualdo/pkg/errs"
	"github.com/stackedboxes/romualdo/pkg/frontend"
	"github.com/stackedboxes/romualdo/pkg/romutil"
)

var scanCmd = &cobra.Command{
	Use:   "scan <path>",
	Short: "Scan the source code and print the tokens",
	Long: `Scan the source code and print the tokens.
This is only useful for testing when developing Romualdo itself.`,
	Args: cobra.ExactArgs(1),

	// For the purposes of the testing, the scanner will run in code mode except
	// between pairs of \passage and \end backslashed keywords (where it will
	// run in lecture mode).
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]
		source, err := os.ReadFile(path)
		if err != nil {
			ctErr := errs.NewCompileTimeWithoutLine(path, err.Error())
			reportAndExit(ctErr)
		}

		scanner := frontend.NewScanner(string(source))
		fmt.Printf("== File: %v\n", path)
		for {
			tok := scanner.Token()
			fmt.Printf("-- Token %.6v %v\n", tok.Line, tok.Kind)
			fmt.Printf("%v\n", romutil.FormatTextForDisplay(tok.Lexeme))

			switch tok.Kind {
			case frontend.TokenKindEOF, frontend.TokenKindError:
				return
			case frontend.TokenKindPassage:
				if tok.IsBackslashed() {
					scanner.SetMode(frontend.ScannerModeLecture)
				}
			case frontend.TokenKindEnd:
				if tok.IsBackslashed() {
					scanner.SetMode(frontend.ScannerModeCode)
				}
			}
		}
	},
}
