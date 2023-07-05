/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2023 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package twi

import (
	"io"

	"github.com/stackedboxes/romualdo/pkg/ast"
	"github.com/stackedboxes/romualdo/pkg/errs"
	"github.com/stackedboxes/romualdo/pkg/frontend"
)

// interpretAST interprets the Storyworld whose AST is passed as argument.
//
// TODO: This will change a lot. For example, currently there is no provision
// for interactivity.
func interpretAST(ast ast.Node, procedures map[string]*ast.ProcedureDecl, out io.Writer) errs.Error {
	i := interpreter{
		ast:        ast,
		procedures: procedures,
		out:        out,
	}

	return i.run()
}

// WalkStoryworld interprets the Storyworld located at path using the tree-walk
// interpreter. Sends output to out.
func WalkStoryworld(path string, out io.Writer) errs.Error {
	ast, err := frontend.ParseStoryworld(path)
	if err != nil {
		return err
	}

	gsv := newGlobalsSymbolVisitor()
	ast.Walk(gsv)
	procedures := gsv.Procedures()

	return interpretAST(ast, procedures, out)
}
