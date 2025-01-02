/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2025 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package backend

import (
	"fmt"

	"github.com/stackedboxes/romualdo/pkg/ast"
	"github.com/stackedboxes/romualdo/pkg/bytecode"
	"github.com/stackedboxes/romualdo/pkg/errs"
)

// codeGenerator contains the code that is common among the actual code
// generation steps.
type codeGenerator struct {
	// csw is the CompiledStoryworld being generated.
	csw *bytecode.CompiledStoryworld

	// debugInfo is the DebugInfo corresponding to the CompiledStoryworld being
	// generated.
	debugInfo *bytecode.DebugInfo

	// compilationContext contains assorted information meant to be shared among
	// different compilation passes.
	compilationContext *compilationContext

	// nodeStack is used to keep track of the nodes being processed. The current
	// one is on the top.
	nodeStack []ast.Node

	// scopeDepth keeps track of the current scope depth we are in. Level 0 is
	// the global scope, and each nested block is one scope level deeper.
	scopeDepth int
}

//
// Other functions
//

// beginScope gets called when we enter into a new scope.
func (cg *codeGenerator) beginScope() {
	cg.scopeDepth++
}

// endScope gets called when we leave a scope.
func (cg *codeGenerator) endScope() {
	cg.scopeDepth--
}

// pushIntoNodeStack pushes a given node to the node stack.
func (cg *codeGenerator) pushIntoNodeStack(node ast.Node) {
	cg.nodeStack = append(cg.nodeStack, node)
}

// popFromNodeStack pops a node from the node stack.
func (cg *codeGenerator) popFromNodeStack() {
	cg.nodeStack = cg.nodeStack[:len(cg.nodeStack)-1]
}

// nodeStackTop returns the node on the top of the node stack.
func (cg *codeGenerator) nodeStackTop() ast.Node {
	return cg.nodeStack[len(cg.nodeStack)-1]
}

// currentLine returns the source code line corresponding to whatever we are
// currently compiling.
func (cg *codeGenerator) currentLine() int {
	return cg.nodeStack[len(cg.nodeStack)-1].Line()
}

// error panics, reporting an error on the current node with a given error
// message.
func (cg *codeGenerator) error(format string, a ...interface{}) {
	e := errs.CompileTime{
		Message:  fmt.Sprintf(format, a...),
		FileName: cg.nodeStackTop().SourceFile(),
		Line:     cg.currentLine(),
	}
	panic(e)
}

// ice reports an Internal Compiler Error.
func (cg *codeGenerator) ice(format string, a ...interface{}) {
	e := errs.NewICE(format, a...)
	panic(e)
}
