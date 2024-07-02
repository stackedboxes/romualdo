/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2024 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package main

import (
	"os"

	"github.com/stackedboxes/romualdo/pkg/errs"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(errs.StatusCodeBadUsage)
	}
}
