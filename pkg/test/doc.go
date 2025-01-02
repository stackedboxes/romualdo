/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2025 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

// The test package contains stuff used for testing Romualdo. It's primarily
// about things used by the `dev test` command that I wished to place elsewhere
// to keep any more involved logic away of the `main` package.
//
// It's also handy to call this from a "unit test" to run the test suite and
// get code coverage reports for it:
//
//	go test -coverpkg=github.com/stackedboxes/romualdo/... -covermode=count -coverprofile=cover.out ./...
//	go tool cover -html=cover.out
package test
