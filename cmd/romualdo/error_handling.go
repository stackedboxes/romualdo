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

	"github.com/stackedboxes/romualdo/pkg/errs"
)

// reportAndExit reports the error err to the end user and exits with the
// appropriate status code. It's fine if err is nil: this just means we had a
// successful execution and therefore we'll exit successfully.
func reportAndExit(err errs.Error) {
	if err == nil {
		os.Exit(errs.StatusCodeSuccess)
	}
	fmt.Fprintf(os.Stderr, "%v\n", err)
	os.Exit(err.ExitCode())
}

// reportAndExitOnError is similar to reportAndExit, but is a no-op if err is
// nil.
func reportAndExitOnError(err errs.Error) {
	if err == nil {
		return
	}
	reportAndExit(err)
}
