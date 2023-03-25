/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2023 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/spf13/cobra"
	"github.com/stackedboxes/romualdo/pkg/ast"
	"github.com/stackedboxes/romualdo/pkg/frontend"
	"github.com/stackedboxes/romualdo/pkg/romutil"
)

var printASTCmd = &cobra.Command{
	Use:   "print-ast",
	Short: "Parse the source code, print the AST",
	Long: `Parse the source code, print the AST. AST stands for "Abstract Syntax Tree",
and if you want to see it, that's your command.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := args[0]
		source, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		ast := frontend.Parse(string(source))
		if ast == nil {
			return errors.New("Parsing error.")
		}

		ap := &ASTPrinter{}
		ast.Walk(ap)
		fmt.Println(ap) // TODO: It's a bit odd to use ASTPrinter.String().

		return nil
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
	case *ast.SourceFile:
		ap.builder.WriteString("SourceFile\n")
	case *ast.ProcDecl:
		ap.builder.WriteString(fmt.Sprintf("ProcDecl [%v %v(%v):%v]\n", n.Kind, n.Name, n.Parameters, n.ReturnType))
	case *ast.Block:
		ap.builder.WriteString("Block\n")
	case *ast.Lecture:
		ap.builder.WriteString(fmt.Sprintf("Lecture [%v]\n", romutil.FormatTextForDisplay(n.Text)))
	default:
		panic(fmt.Sprintf("Unexpected node type: %T", n))
	}

	ap.indentLevel++
}

func (ap *ASTPrinter) Leave(ast.Node) {
	ap.indentLevel--
}

func (ap *ASTPrinter) Event(node ast.Node, event int) {
}

// indent returns a string good for indenting code level levels deep.
func indent(level int) string {
	return strings.Repeat("\t", level)
}
