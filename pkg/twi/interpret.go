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
func InterpretAST(ast ast.Node, procedures map[string]*ast.ProcDecl, out io.Writer) error {
	i := interpreter{
		ast:        ast,
		procedures: procedures,
		out:        out,
	}

	return i.run()
}

// InterpretSource interprets the Storyworld whose source is passed as argument.
func InterpretSource(path string, out io.Writer) error {
	// TODO: I think this preamble tends to repeat itself in different
	// commands. Factor it out!
	source, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	ast := frontend.Parse(string(source))
	if ast == nil {
		return errors.New("Parsing error.")
	}

	// TODO: This is looking messy. We probably shouldn't be instantiating
	// and using the visitor ourselves here.
	gsv := newGlobalsSymbolVisitor()
	ast.Walk(gsv)
	procedures := gsv.Procedures()

	return InterpretAST(ast, procedures, out)
}
