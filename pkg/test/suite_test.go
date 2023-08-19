/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2023 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package test

import (
	"testing"
)

// TestRunSuite runs the Romualdo test suite. This is not a proper unit test,
// but instead a simple way to run our end-to-end tests and, more importantly,
// to get code coverage reports for them.
func TestRunSuite(t *testing.T) {
	err := ExecuteSuite("../../test/suite")
	if err != nil {
		t.Fatalf("Error running test suite: %v", err)
	}
}
