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
		os.Exit(0)
	case errors.As(err, &compTimeError):
		fmt.Printf("%v\n", compTimeError)
		os.Exit(1) // TODO: Document what each status code means.
	case errors.As(err, &testSuiteError):
		fmt.Printf("%v\n", testSuiteError)
		os.Exit(2) // TODO: Document what each status code means.
	case errors.As(err, &compTimeColl):
		fmt.Printf("%v", compTimeColl)
		os.Exit(1) // TODO: Document what each status code means.
	case errors.As(err, &iceErr):
		fmt.Printf("Internal Compiler Error: %v\n", iceErr)
		os.Exit(171) // TODO: Document what each status code means.
	default:
		fmt.Printf("Internal Compiler Error: unexpected error of type %T: %v\n", err, err)
		os.Exit(171) // TODO: Document what each status code means.
	}
}
