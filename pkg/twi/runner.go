/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2023 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package twi

import (
	"github.com/stackedboxes/romualdo/pkg/ast"
	"github.com/stackedboxes/romualdo/pkg/errs"
	"github.com/stackedboxes/romualdo/pkg/frontend"
	"github.com/stackedboxes/romualdo/pkg/romutil"
)

// runner is a romutil.Runner that uses the tree-walk interpreter to run a
// Storyworld.
type runner struct {
	ast        *ast.Storyworld
	procedures map[string]*ast.ProcedureDecl
}

// NewRunner creates a new TWIRunner.
func NewRunner() romutil.Runner {
	return &runner{}
}

// Build satisfies the romutil.Runner interface.
func (r *runner) Build(path string) errs.Error {
	ast, err := frontend.ParseStoryworld(path)
	if err != nil {
		return err
	}
	r.ast = ast

	gsv := newGlobalsSymbolVisitor()
	ast.Walk(gsv)
	r.procedures = gsv.Procedures()

	return nil
}

// Run satisfies the romutil.Runner interface.
func (r *runner) Run(mouth romutil.Mouth, ear romutil.Ear) errs.Error {
	return interpretAST(r.ast, r.procedures, mouth, ear)
}
