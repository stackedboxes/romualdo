/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2023 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package main

import (
	"github.com/spf13/cobra"
	"github.com/stackedboxes/romualdo/pkg/test"
)

var devTestCmd = &cobra.Command{
	Use:   "test",
	Short: "Run a Romualdo test suite",
	Long:  `Run a Romualdo test suite (i.e., meant to test Romualdo itself).`,
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		err := test.ExecuteSuite(flagDevTestWalkDontRun, flagDevTestSuite)
		reportAndExit(err)
	},
}

// flagDevTestSuite is the value of the --suite flag of the `dev test` command.
var flagDevTestSuite string

// flagDevTestWalkDontRun is the value of the --walk-dont-run flag of the `dev
// test` command.
var flagDevTestWalkDontRun bool

func init() {
	devTestCmd.Flags().StringVarP(&flagDevTestSuite, "suite", "s",
		"./test", "Path to the test suite to run")

	devTestCmd.Flags().BoolVarP(&flagDevTestWalkDontRun, "walk-dont-run", "w",
		false, "Test using the walk tree interpreter instead of the bytecode one")
}
