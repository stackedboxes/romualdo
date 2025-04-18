/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2025 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/stackedboxes/romualdo/pkg/ast"
	"github.com/stackedboxes/romualdo/pkg/frontend"
	"github.com/stackedboxes/romualdo/pkg/romutil"
)

var devPrintASTCmd = &cobra.Command{
	Use:   "print-ast <path>",
	Short: "Parse the source code, print the AST",
	Long: `Parse the source code, print the AST. AST stands for "Abstract Syntax Tree",
and if you want to see it, that's your command.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]
		ast, err := frontend.ParseFile(path, filepath.Dir(path))
		reportAndExitOnError(err)

		ap := &ASTPrinter{}
		ast.Walk(ap)
		fmt.Println(ap) // TODO: It's a bit odd to use ASTPrinter.String().
	},
}

type ASTPrinter struct {
	indentLevel int
	builder     strings.Builder
}

func (ap *ASTPrinter) String() string {
	return ap.builder.String()
}

func (ap *ASTPrinter) Enter(node ast.Node) {
	ap.builder.WriteString(indent(ap.indentLevel))

	switch n := node.(type) {
	case *ast.Binary:
		ap.builder.WriteString(fmt.Sprintf("Binary [%v]\n", n.Operator))
	case *ast.Block:
		ap.builder.WriteString("Block\n")
	case *ast.BoolLiteral:
		ap.builder.WriteString(fmt.Sprintf("BoolLiteral [%v]\n", n.Value))
	case *ast.Curlies:
		ap.builder.WriteString("Curlies\n")
	case *ast.ExpressionStmt:
		ap.builder.WriteString("ExpressionStmt\n")
	case *ast.IfStmt:
		ap.builder.WriteString("If\n")
	case *ast.Lecture:
		ap.builder.WriteString(fmt.Sprintf("Lecture [%v]\n", romutil.FormatTextForDisplay(n.Text)))
	case *ast.Listen:
		ap.builder.WriteString("Listen\n")
	case *ast.ProcedureDecl:
		ap.builder.WriteString(fmt.Sprintf("ProcDecl [%v %v(%v):%v]\n", n.Kind, n.Name, n.Parameters, n.ReturnType))
	case *ast.Say:
		ap.builder.WriteString("Say\n")
	case *ast.SourceFile:
		ap.builder.WriteString("SourceFile\n")
	case *ast.StringLiteral:
		ap.builder.WriteString(fmt.Sprintf("StringLiteral [%v]\n", romutil.FormatTextForDisplay(n.Value)))
	default:
		panic(fmt.Sprintf("Unexpected node type: %T", n))
	}

	ap.indentLevel++
}

func (ap *ASTPrinter) Leave(ast.Node) {
	ap.indentLevel--
}

func (ap *ASTPrinter) Event(node ast.Node, event ast.EventType) {
	// Nothing
}

// indent returns a string good for indenting code level levels deep.
func indent(level int) string {
	return strings.Repeat("\t", level)
}
