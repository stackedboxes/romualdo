/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2023 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package twi

import (
	"github.com/stackedboxes/romualdo/pkg/ast"
)

// globalsSymbolVisitor is a visitor that produces a table of global symbols.
type globalsSymbolVisitor struct {
	level      int
	procedures map[string]*ast.ProcDecl
}

// newGlobalsSymbolVisitor cretes a new newGlobalsSymbolVisitor.
func newGlobalsSymbolVisitor() *globalsSymbolVisitor {
	return &globalsSymbolVisitor{
		procedures: map[string]*ast.ProcDecl{},
	}
}

// Procedures returns the symbol table of global procedures. Must be called
// after traversing the AST.
func (g *globalsSymbolVisitor) Procedures() map[string]*ast.ProcDecl {
	return g.procedures
}

func (g *globalsSymbolVisitor) Enter(node ast.Node) {
	defer func() { g.level++ }()
	switch n := node.(type) {
	case *ast.ProcDecl:
		// Level 0 is the File itself; globals are at level 1
		if g.level != 1 {
			return
		}
		// TODO: Check for duplicates here? Or somewhere else? Or both?
		g.procedures[n.Name] = n
	default:
		// Nothing
	}
}

func (g *globalsSymbolVisitor) Leave(ast.Node) {
	g.level--
}
