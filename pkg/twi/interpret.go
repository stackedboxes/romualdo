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
	"github.com/stackedboxes/romualdo/pkg/frontend"
)

// interpretAST interprets the Storyworld whose AST is passed as argument.
//
// TODO: This will change a lot. For example, currently there is no provision
// for interactivity.
func interpretAST(ast ast.Node, procedures map[string]*ast.ProcDecl, out io.Writer) error {
	i := interpreter{
		ast:        ast,
		procedures: procedures,
		out:        out,
	}

	return i.run()
}

// InterpretStoryworld interprets the Storyworld located at path.
func InterpretStoryworld(path string, out io.Writer) error {
	ast, err := frontend.ParseStoryworld(path)
	if err != nil {
		return err
	}

	gsv := newGlobalsSymbolVisitor()
	ast.Walk(gsv)
	procedures := gsv.Procedures()

	return interpretAST(ast, procedures, out)
}
