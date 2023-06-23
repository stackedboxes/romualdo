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
	"github.com/stackedboxes/romualdo/pkg/backend"
	"github.com/stackedboxes/romualdo/pkg/errs"
	"github.com/stackedboxes/romualdo/pkg/frontend"
)

var buildCmd = &cobra.Command{
	Use:   "build <path>",
	Short: "Builds the Storyworld from source",
	Long:  `Builds the Storyworld from source.`,
	Args:  cobra.ExactArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]
		fileInfo, err := os.Stat(path)
		if err != nil {
			cpErr := errs.NewCommandPrep(err.Error())
			errs.ReportAndExit(cpErr)
		}

		if !fileInfo.IsDir() {
			buErr := errs.NewBadUsage("the build command expects a directory, but %v isn't one", path)
			errs.ReportAndExit(buErr)
		}

		swAST, err := frontend.ParseStoryworld(path)
		if err != nil {
			errs.ReportAndExit(err)
		}
		csw, di, err := backend.GenerateCode(swAST, path)
		_ = di

		// TODO: Save the Compiled Storyworld and debug info to disk.
		cswFile, err := os.Create("csw.ras")
		if err != nil {
			errs.ReportAndExit(err)
		}
		defer cswFile.Close()

		err = csw.Serialize(cswFile)

		errs.ReportAndExit(err)
	},
}
