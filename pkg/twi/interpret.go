/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2023 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package twi

import "github.com/stackedboxes/romualdo/pkg/ast"

// Interpret interprets the storyworld whose AST is passed as argument.
//
// TODO: This will change a lot. For example, currently there is no provision
// for interactivity.
func Interpret(ast ast.Node, procedures map[string]*ast.ProcDecl) error {
	i := interpreter{
		ast:        ast,
		procedures: procedures,
	}

	return i.run()
}
