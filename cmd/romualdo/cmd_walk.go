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
	"github.com/stackedboxes/romualdo/pkg/errs"
	"github.com/stackedboxes/romualdo/pkg/twi"
)

var walkCmd = &cobra.Command{
	Use:   "walk <path>",
	Short: "Runs the Storyworld using the tree-walk interpreter",
	Long:  `Runs the Storyworld using the tree-walk interpreter.`,
	Args:  cobra.ExactArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]
		fileInfo, err := os.Stat(path)
		if err != nil {
			cpErr := errs.NewCommandPrep(err.Error())
			errs.ReportAndExit(cpErr)
		}

		if !fileInfo.IsDir() {
			buErr := errs.NewBadUsage("the walk command expects a directory, but %v isn't one", path)
			errs.ReportAndExit(buErr)
		}

		err = twi.WalkStoryworld(path, os.Stdout)
		errs.ReportAndExit(err)
	},
}
