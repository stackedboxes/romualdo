/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2023 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package errs

const (
	// StatusCodeSuccess indicates a successful execution.
	StatusCodeSuccess = 0

	// StatusCodeCompileTimeError indicates a compile-time error.
	StatusCodeCompileTimeError = 1

	// StatusCodeTestSuiteError indicates a failure while running Romualdo's own
	// test suite.
	StatusCodeTestSuiteError = 2

	// StatusCodeBadUsage indicates some user error in the usage of the romualdo
	// tool (e.g., passing the wrong number of arguments, or passing a
	// nonexisting command-line flag).
	StatusCodeBadUsage = 50

	// StatusCodeICE indicates an Internal Compiler Error.
	StatusCodeICE = 125
)
