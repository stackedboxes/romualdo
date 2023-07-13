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
			sc.error("Duplicate procedure `%v` at line %v. The first one was at line %v.",
				n.Name, n.LineNumber, line)
			break
		}
		sc.proceduresLine[n.Name] = n.LineNumber
	}
}

func (sc *semanticChecker) Leave(n ast.Node) {
	sc.nodeStack = sc.nodeStack[:len(sc.nodeStack)-1]

	// TODO: checking for `main` here for now; will need to look at the whole
	// Root Package when we have proper support for Packages.
	if _, ok := n.(*ast.SourceFile); ok {
		if _, found := sc.proceduresLine["main"]; !found {
			sc.error("Procedure `main` not found.")
		}
	}
}

func (sc *semanticChecker) Event(node ast.Node, event ast.EventType) {
	// Nothing
}

//
// Semantic checking
//

// error reports an error.
func (sc *semanticChecker) error(format string, a ...interface{}) {
	sc.errors.Add(errs.NewCompileTimeWithoutLine(sc.fileName, format, a...))
}

// currentLine returns the source code line corresponding to whatever we are
// currently analyzing.
func (sc *semanticChecker) currentLine() int {
	return sc.nodeStack[len(sc.nodeStack)-1].Line()
}
