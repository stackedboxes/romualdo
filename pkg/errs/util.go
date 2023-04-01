/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2023 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package errs

import (
	"errors"
	"fmt"
	"os"
)

// ReportAndExit reports the error err to the end user and exits with the
// appropriate status code. It's fine if err is nil, we handle this case here.
func ReportAndExit(err error) {
	compTimeError := &CompileTime{}
	compTimeColl := &CompileTimeCollection{}
	testSuiteError := &TestSuite{}
	iceErr := &ICE{}
	switch {
	case err == nil:
		os.Exit(StatusCodeSuccess)

	case errors.As(err, &compTimeError):
		fmt.Printf("%v\n", compTimeError)
		os.Exit(StatusCodeCompileTimeError)

	case errors.As(err, &compTimeColl):
		fmt.Printf("%v", compTimeColl)
		os.Exit(StatusCodeCompileTimeError)

	case errors.As(err, &testSuiteError):
		fmt.Printf("%v\n", testSuiteError)
		os.Exit(StatusCodeTestSuiteError)

	case errors.As(err, &iceErr):
		fmt.Printf("Internal Compiler Error: %v\n", iceErr)
		os.Exit(StatusCodeICE)

	default:
		fmt.Printf("Internal Compiler Error: unexpected error of type %T: %v\n", err, err)
		os.Exit(StatusCodeICE)
	}
}
