/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2023 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package vm

import (
	"github.com/stackedboxes/romualdo/pkg/bytecode"
	"github.com/stackedboxes/romualdo/pkg/errs"
	"github.com/stackedboxes/romualdo/pkg/romutil"
)

// runner is a romutil.Runner that uses the bytecode VM to run a Storyworld.
type runner struct {
	trace bool
	csw   *bytecode.CompiledStoryworld
	di    *bytecode.DebugInfo
}

// NewRunner creates a new Runner based on the bytecode VM.
func NewRunner(trace bool) romutil.Runner {
	return &runner{
		trace: trace,
	}
}

// Build satisfies the romutil.Runner interface.
func (r *runner) Build(path string) errs.Error {
	csw, di, err := cswFromPath(path)
	if err != nil {
		return err
	}
	r.csw = csw
	r.di = di
	return nil
}

// Run satisfies the romutil.Runner interface.
func (r *runner) Run(mouth romutil.Mouth, ear romutil.Ear) errs.Error {
	return runCSW(r.csw, r.di, mouth, ear, r.trace)
}
