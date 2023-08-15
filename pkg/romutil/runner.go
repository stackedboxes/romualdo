/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2023 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package romutil

import (
	"github.com/stackedboxes/romualdo/pkg/errs"
)

//
// Abstract Runner interface
//

// A Runner can build and run a Storyworld. Meant to abstract away the
// differences between the tree-walk interpreter and the bytecode VM with
// regards to building and running.
type Runner interface {
	// Build builds the Storyworld located at path. Can be called multiple
	// times.
	Build(path string) errs.Error

	// Run runs the Storyworld. Can be called only once after each successful
	// call to Build().
	Run(mouth Mouth, ear Ear) errs.Error
}
