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
	"github.com/stackedboxes/romualdo/pkg/romutil"
)

// interpretAST interprets the Storyworld whose AST is passed as argument.
func interpretAST(ast ast.Node, procedures map[string]*ast.ProcedureDecl, mouth romutil.Mouth, ear romutil.Ear) errs.Error {
	i := interpreter{
		ast:        ast,
		procedures: procedures,
		mouth:      mouth,
		ear:        ear,
	}

	return i.run()
}
