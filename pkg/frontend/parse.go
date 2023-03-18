/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2023 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package frontend

import (
	"github.com/stackedboxes/romualdo/pkg/ast"
)

// Parse parses Romualdo Language source code and returns its AST (Abstract
// Syntax Tree). In case of errors, returns nil and prints the error messages to
// the standard error.
func Parse(source string) ast.Node {
	p := newParser(source)
	root := p.parse()
	if root == nil {
		return nil
	}

	return root
}
