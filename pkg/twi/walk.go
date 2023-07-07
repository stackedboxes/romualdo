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

// interpretAST interprets the Storyworld whose AST is passed as argument.
//
// TODO: This will change a lot. For example, currently there is no provision
// for interactivity.
func interpretAST(ast ast.Node, procedures map[string]*ast.ProcedureDecl, mouth romutil.Mouth, ear romutil.Ear) errs.Error {
	i := interpreter{
		ast:        ast,
		procedures: procedures,
		mouth:      mouth,
		ear:        ear,
	}

	return i.run()
}

// WalkStoryworld interprets the Storyworld located at path using the tree-walk
// interpreter. Sends output to out.
func WalkStoryworld(path string, mouth romutil.Mouth, ear romutil.Ear) errs.Error {
	ast, err := frontend.ParseStoryworld(path)
	if err != nil {
		return err
	}

	gsv := newGlobalsSymbolVisitor()
	ast.Walk(gsv)
	procedures := gsv.Procedures()

	return interpretAST(ast, procedures, mouth, ear)
}
