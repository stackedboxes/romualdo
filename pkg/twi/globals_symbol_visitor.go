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

// GlobalsSymbolVisitor is a visitor that produces a table of global symbols.
type GlobalsSymbolVisitor struct {
	level      int
	procedures map[string]*ast.ProcDecl
}

// NewGlobalsSymbolVisitor cretes a new NewGlobalsSymbolVisitor.
func NewGlobalsSymbolVisitor() *GlobalsSymbolVisitor {
	return &GlobalsSymbolVisitor{
		procedures: map[string]*ast.ProcDecl{},
	}
}

// Procedures returns the symbol table of global procedures. Must be called
// after traversing the AST.
func (g *GlobalsSymbolVisitor) Procedures() map[string]*ast.ProcDecl {
	return g.procedures
}

func (g *GlobalsSymbolVisitor) Enter(node ast.Node) {
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

func (g *GlobalsSymbolVisitor) Leave(ast.Node) {
	g.level--
}