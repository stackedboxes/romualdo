/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2025 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package frontend

import (
	"github.com/stackedboxes/romualdo/pkg/ast"
	"github.com/stackedboxes/romualdo/pkg/errs"
)

// semanticChecker is a node visitor that implements assorted semantic checks.
//
// TODO: This operates at source file level. It should operate at package level,
// or storyworld level.
type semanticChecker struct {
	// fileName is the name of the file being semantically checked.
	fileName string

	// errors collects the errors for all semantic errors detected.
	errors *errs.CompileTimeCollection

	// nodeStack is used to keep track of the nodes being processed. The current
	// one is on the top.
	nodeStack []ast.Node

	// proceduresLine contains the line number where a given procedure was
	// found. The procedure names here do not contain the package name (the
	// semantic checker operates at one package at a time, so the package name
	// is not relevant).
	proceduresLine map[string]int
}

func NewSemanticChecker(fileName string) *semanticChecker {
	return &semanticChecker{
		fileName:       fileName,
		errors:         &errs.CompileTimeCollection{},
		proceduresLine: make(map[string]int),
	}
}

// The Visitor interface
func (sc *semanticChecker) Enter(node ast.Node) {
	sc.nodeStack = append(sc.nodeStack, node)

	switch n := node.(type) {
	case *ast.ProcedureDecl:
		// TODO: Do this check at Package or Storyworld level.
		if line, found := sc.proceduresLine[n.Name]; found {
			sc.errorAtCurrentNode("Duplicate procedure `%v`. First definition at line %v.",
				n.Name, line)
			break
		}
		sc.proceduresLine[n.Name] = n.LineNumber
	}
}

func (sc *semanticChecker) Leave(n ast.Node) {
	sc.nodeStack = sc.nodeStack[:len(sc.nodeStack)-1]

	// TODO: checking for `main` here for now; will need to look at the whole
	// Root Package when we have proper support for Packages. At which point
	// we'll want to use `/main` in the message.
	if _, ok := n.(*ast.SourceFile); ok {
		if _, found := sc.proceduresLine["main"]; !found {
			sc.errorWithoutLine("Procedure `main` not found.")
		}
	}
}

func (sc *semanticChecker) Event(node ast.Node, event ast.EventType) {
	// Nothing
}

//
// Error reporting
//

// errorWithoutLine reports an error without a specific line number.
func (tc *semanticChecker) errorWithoutLine(format string, a ...interface{}) {
	tc.errors.Add(errs.NewCompileTimeWithoutLine(tc.fileName, format, a...))
}

// errorAtCurrentNode reports an error at the node we are currently checking.
func (tc *semanticChecker) errorAtCurrentNode(format string, a ...interface{}) {
	tc.errors.Add(errs.NewCompileTime(tc.fileName, tc.currentLine(), format, a...))
}

// currentLine returns the source code line corresponding to whatever we are
// currently analyzing.
func (tc *semanticChecker) currentLine() int {
	return tc.nodeStack[len(tc.nodeStack)-1].Line()
}
