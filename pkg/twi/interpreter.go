/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2023 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package twi

import (
	"fmt"
	"os"

	"github.com/stackedboxes/romualdo/pkg/ast"
	"github.com/stackedboxes/romualdo/pkg/bytecode"
	"github.com/stackedboxes/romualdo/pkg/errs"
	"github.com/stackedboxes/romualdo/pkg/romutil"
)

// interpreter is a tree-walk interpreter for a Romualdo AST.
type interpreter struct {
	ast        ast.Node
	procedures map[string]*ast.ProcedureDecl
	mouth      romutil.Mouth
	ear        romutil.Ear
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

// interpretStatement interprets the stmt statement.
func (i *interpreter) interpretStatement(stmt ast.Node) errs.Error {
	switch n := stmt.(type) {
	case *ast.Lecture:
		i.mouth.Say(n.Text)

	case *ast.Say:
		for _, lecture := range n.Lectures {
			if err := i.interpretStatement(lecture); err != nil {
				return err
			}
		}
		// This is a no-op. The actual `say`ing is done by the Lecture nodes
		// exist within the Say node.

	case *ast.ExpressionStmt:
		// Interpret the expression and discard the result.
		_, err := i.interpretExpression(n.Expr)
		return err

	case *ast.IfStmt:
		condition, err := i.interpretExpression(n.Condition)
		if err != nil {
			return err
		}
		if !condition.IsBool() {
			return errs.NewRuntime("if condition must be a Boolean, got %T", condition.Value)
		}

		if condition.AsBool() {
			return i.interpretBlock(n.Then)
		}

		if n.Else != nil {
			// Else can be either a block or (in the case of an "elseif") an
			// "if" statement.
			if block, ok := n.Else.(*ast.Block); ok {
				return i.interpretBlock(block)
			}
			return i.interpretStatement(n.Else)
		}

	case *ast.Curlies:
		v, err := i.interpretExpression(n.Expr)
		if err != nil {
			return err
		}
		i.mouth.Say(v.AsString())

	default:
		return errs.NewRuntime("unknown statement type: %T", stmt)
	}

	return nil
}

// interpretExpression interprets the expr expression and returns its value.
func (i *interpreter) interpretExpression(expr ast.Node) (bytecode.Value, errs.Error) {
	switch n := expr.(type) {
	case *ast.StringLiteral:
		return bytecode.Value{Value: n.Value}, nil

	case *ast.BoolLiteral:
		return bytecode.Value{Value: n.Value}, nil

	case *ast.Listen:
		// TODO: Currently this just assumes the argument to listen is a string
		// literal. This will break bad once we have more complex expressions.
		// Should do for now, though.
		options := n.Options.(*ast.StringLiteral).Value

		i.mouth.Say(options) // TODO: Temporary, to see what's happening.

		fmt.Fprintf(os.Stdout, "%v\n", options) // TODO: Temporary, to see what's happening.

		fmt.Fprint(os.Stdout, "> ") // TODO: Temporary, to see what's happening.
		choice := i.ear.Listen()
		fmt.Fprintf(os.Stdout, "USER INPUT: "+choice) // TODO: Temporary, to see what's happening.

		return bytecode.Value{Value: choice}, nil

	case *ast.Binary:
		lhsValue, err := i.interpretExpression(n.LHS)
		if err != nil {
			return bytecode.Value{}, err
		}
		rhsValue, err := i.interpretExpression(n.RHS)
		if err != nil {
			return bytecode.Value{}, err
		}

		switch n.Operator {
		case "==":
			return bytecode.Value{Value: bytecode.ValuesEqual(lhsValue, rhsValue)}, nil
		case "!=":
			return bytecode.Value{Value: !bytecode.ValuesEqual(lhsValue, rhsValue)}, nil
		default:
			return bytecode.Value{}, errs.NewRuntime("unknown binary operator: '%v'", n.Operator)
		}

	default:
		return bytecode.Value{}, errs.NewRuntime("unknown expression type: %T", expr)
	}
}
