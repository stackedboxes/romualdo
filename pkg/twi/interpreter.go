/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2023 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package twi

import (
	"errors"
	"fmt"

	"github.com/stackedboxes/romualdo/pkg/ast"
)

// interpreter is a tree-walk interpreter for a Romualdo AST.
type interpreter struct {
	ast        ast.Node
	procedures map[string]*ast.ProcDecl
}

// run runs ("walks"?) the Storyworld whose AST is in i.ast.
func (i *interpreter) run() error {
	main, ok := i.procedures["main"]
	if !ok {
		return errors.New(`Missing "main" procedure`)
	}
	return i.interpretProcedure(main)
}

//
// interpret*() methods
//

func (i *interpreter) interpretProcedure(proc *ast.ProcDecl) error {
	return i.interpretBlock(proc.Body)
}

func (i *interpreter) interpretBlock(block *ast.Block) error {
	for _, stmt := range block.Statements {
		if err := i.interpretStatement(stmt); err != nil {
			return err
		}
	}
	return nil
}

func (i *interpreter) interpretStatement(stmt ast.Node) error {
	switch n := stmt.(type) {
	case *ast.Lecture:
		fmt.Println(n.Text)
	default:
		return fmt.Errorf("unknown statement type: %T", stmt)
	}
	return nil
}