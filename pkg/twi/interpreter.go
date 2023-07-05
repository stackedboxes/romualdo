/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2023 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package twi

import (
	"fmt"
	"io"

	"github.com/stackedboxes/romualdo/pkg/ast"
	"github.com/stackedboxes/romualdo/pkg/errs"
)

// interpreter is a tree-walk interpreter for a Romualdo AST.
type interpreter struct {
	ast        ast.Node
	procedures map[string]*ast.ProcedureDecl
	out        io.Writer
}

// run runs ("walks"?) the Storyworld whose AST is in i.ast.
func (i *interpreter) run() errs.Error {
	main, ok := i.procedures["/main"]
	if !ok {
		return errs.NewRuntime("Missing '/main' procedure")
	}
	return i.interpretProcedure(main)
}

//
// interpret*() methods
//

func (i *interpreter) interpretProcedure(proc *ast.ProcedureDecl) errs.Error {
	return i.interpretBlock(proc.Body)
}

func (i *interpreter) interpretBlock(block *ast.Block) errs.Error {
	for _, stmt := range block.Statements {
		if err := i.interpretStatement(stmt); err != nil {
			return err
		}
	}
	return nil
}

func (i *interpreter) interpretStatement(stmt ast.Node) errs.Error {
	switch n := stmt.(type) {
	case *ast.Lecture:
		fmt.Fprintf(i.out, n.Text)
	default:
		return errs.NewRuntime("unknown statement type: %T", stmt)
	}
	return nil
}
