/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2023 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package frontend

import (
	"github.com/stackedboxes/romualdo/pkg/ast"
	"github.com/stackedboxes/romualdo/pkg/errs"
)

// typeChecker is a node visitor that implements type checking.
type typeChecker struct {
	// fileName is the name of the file being type checked.
	fileName string

	// errors collects the errors for all semantic errors detected.
	errors *errs.CompileTimeCollection

	// nodeStack is used to keep track of the nodes being processed. The current
	// one is on the top.
	nodeStack []ast.Node
}

func NewTypeChecker(fileName string) *typeChecker {
	return &typeChecker{
		fileName: fileName,
		errors:   &errs.CompileTimeCollection{},
	}
}

// The Visitor interface
func (tc *typeChecker) Enter(node ast.Node) {
	tc.nodeStack = append(tc.nodeStack, node)

	switch n := node.(type) {
	case *ast.Listen:
		tc.checkListen(n)
	case *ast.IfStmt:
		tc.checkIfStmt(n)
	}
}

func (tc *typeChecker) Leave(ast.Node) {
	tc.nodeStack = tc.nodeStack[:len(tc.nodeStack)-1]
}

func (tc *typeChecker) Event(node ast.Node, event ast.EventType) {
	// Nothing
}

//
// Type checking
//

// checkListen type checks a listen expression.
func (tc *typeChecker) checkListen(node *ast.Listen) {
	optionsType := node.Options.Type()
	if optionsType != ast.TypeString {
		tc.errorWithoutLine("listen expects a string argument, got a %v.", optionsType)
	}
}

// checkIfStmt type checks an if statement.
func (tc *typeChecker) checkIfStmt(node *ast.IfStmt) {
	conditionType := node.Condition.Type()
	if conditionType != ast.TypeBool {
		tc.errorAtCurrentNode("'if' condition must be a Boolean expression, got a %v.", conditionType)
	}
}

// errorWithoutLine reports an error without a specific line number.
func (tc *typeChecker) errorWithoutLine(format string, a ...interface{}) {
	tc.errors.Add(errs.NewCompileTimeWithoutLine(tc.fileName, format, a...))
}

// errorAtCurrentNode reports an error at the node we are currently checking.
func (tc *typeChecker) errorAtCurrentNode(format string, a ...interface{}) {
	tc.errors.Add(errs.NewCompileTime(tc.fileName, tc.currentLine(), format, a...))
}

// currentLine returns the source code line corresponding to whatever we are
// currently analyzing.
func (tc *typeChecker) currentLine() int {
	return tc.nodeStack[len(tc.nodeStack)-1].Line()
}
