/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2023 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package ast

// BaseNode contains the functionality common to all AST nodes.
type BaseNode struct {
	// LineNumber stores the line number from where this node comes.
	LineNumber int
}

func (n *BaseNode) Line() int {
	return n.LineNumber
}

//
// The AST nodes
//

// Storyworld is an AST node representing everything from a Romualdo source code
// file. It is the root of the AST.
//
// TODO: Likely to not be the root of the AST once we start supporting multiple
// files.
type SourceFile struct {
	BaseNode

	// Declarations stores all the declarations found in the source file.
	Declarations []Node
}

func (n *SourceFile) Type() TypeTag {
	return TypeVoid
}

func (n *SourceFile) Walk(v Visitor) {
	v.Enter(n)
	for _, decl := range n.Declarations {
		decl.Walk(v)
	}
	v.Leave(n)
}

// ProcDecl is an AST node representing the declaration (and the definition,
// Romualdo doesn't have this distinction) of a procedure. A procedure can be
// either a function or a passage.
type ProcDecl struct {
	BaseNode

	Kind       ProcKind
	Name       string
	ReturnType TypeTag
	Parameters []Parameter
	Body       *Block
}

func (n *ProcDecl) Type() TypeTag {
	return TypeVoid
}

func (n *ProcDecl) Walk(v Visitor) {
	v.Enter(n)
	n.Body.Walk(v)
	v.Leave(n)
}

// Block is an AST node representing a block of code. Importantly, a block
// defines a scope.
type Block struct {
	BaseNode

	// The statements that make up this block.
	Statements []Node
}

func (n *Block) Type() TypeTag {
	return TypeVoid
}

func (n *Block) Walk(v Visitor) {
	v.Enter(n)
	for _, stmt := range n.Statements {
		stmt.Walk(v)
	}
	v.Leave(n)
}

// Text is an AST node representing a text statement.
//
// TODO: It's odd to call it a "text statement". Feels more like a "text
// literal" -- but it works as a statement in practice. Must thing about it.
// Especially, need to take into account how to create the equivalent thing in
// code mode, *and* how to deal with "interpolated" values within text in text
// mode.
type Text struct {
	BaseNode

	// Text contains the text's text. I swear this makes sense!
	Text string
}

func (n *Text) Type() TypeTag {
	// TODO: Here's another misty point of the design: what is the type of a
	// Text? As a statement, it can be void as we are doing here. But if we look
	// at it as a literal, it must have some non-void type! Maybe its own text
	// type? Maybe string?
	return TypeVoid
}

func (n *Text) Walk(v Visitor) {
	v.Enter(n)
	v.Leave(n)
}

//
// Helper types
//

// Parameter is a parameter of a procedure.
type Parameter struct {
	// Name is the parameter name.
	Name string

	// Type is the parameter type.
	//
	// TODO: Eventually we'll support user-defined types, then we'll use *Type
	// instead of TypeTag here (and a Type will have a TypeTag field).
	Type TypeTag
}

// ProcKind represents what kind of procedure a procedure is.
type ProcKind int

const (
	ProcKindFunction ProcKind = iota
	ProcKindPassage
)
