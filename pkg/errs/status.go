/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2024 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package errs

const (
	// statusCodeSuccess indicates a successful execution.
	StatusCodeSuccess = 0

	// statusCodeCompileTimeError indicates a compile-time error.
	statusCodeCompileTimeError = 1

	// statusCodeTestSuiteError indicates a failure while running Romualdo's own
	// test suite.
	statusCodeTestSuiteError = 2

	// statusCodeBadUsage indicates some user error in the usage of the romualdo
	// tool (e.g., passing the wrong number of arguments, or passing a
	// nonexisting command-line flag).
	StatusCodeBadUsage = 3

	// statusCodeRomualdoToolError indicates an error while running the romualdo
	// tool that doesn't fit in any of the other categories.
	statusCodeRomualdoToolError = 4

	// statusCodeRuntimeError indicates something bad happened at runtime. This
	// isn't expected to happen, and should indicate a bug in the compiler or in
	// the language. (Well, ideally. As of July 2023 I cannot promise this is
	// valid!)
	statusCodeRuntimeError = 100

	// statusCodeICE indicates an Internal Compiler Error.
	statusCodeICE = 125
)
