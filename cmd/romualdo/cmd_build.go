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
	"github.com/stackedboxes/romualdo/pkg/romutil"
)

var buildCmd = &cobra.Command{
	Use:   "build <path>",
	Short: "Builds the Storyworld from source",
	Long:  `Builds the Storyworld from source.`,
	Args:  cobra.ExactArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		swPath := args[0]
		if isDir, err := romutil.IsDir(swPath); err != nil || !isDir {
			buErr := errs.NewBadUsage("The build command expects a directory, but %v isn't one", swPath)
			reportAndExit(buErr)
		}

		swAST, err := frontend.ParseStoryworld(swPath)
		reportAndExitOnError(err)

		csw, di, err := backend.GenerateCode(swAST)
		reportAndExitOnError(err)

		cswFile, plainErr := os.Create("csw.ras")
		err = csw.Serialize(cswFile)
		reportAndExitOnError(err)
		defer cswFile.Close()

		debugInfoFile, plainErr := os.Create("csw.rad")
		if plainErr != nil {
			err = errs.NewRomualdoTool("creating debug info file: %v", plainErr)
			reportAndExit(err)
		}
		defer debugInfoFile.Close()
		err = di.Serialize(debugInfoFile)
		reportAndExit(err)
	},
}
