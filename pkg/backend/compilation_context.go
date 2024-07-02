/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2024 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package backend

// A compilationContext stores information needed throughout different
// compilation passes.
type compilationContext struct {

	// procNameToIndex maps a fully-qualified Procedure name to its index into
	// the slice of Chunks.
	procNameToIndex map[string]int
}

// newCompilationContext creates a new compilationContext.
func newCompilationContext() *compilationContext {
	return &compilationContext{
		procNameToIndex: map[string]int{},
	}
}
