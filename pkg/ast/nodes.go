/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2023 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package ast

import "fmt"

// BaseNode contains the functionality common to all AST nodes.
type BaseNode struct {
	// SrcFile stores the file name (from the Storyworld root) where this node
	// was defined.
	SrcFile string

	// LineNumber stores the line number where this node was defined.
	LineNumber int
}

// SourceFile returns the file name (from the Storyworld root) where this node
// was defined.
func (n *BaseNode) SourceFile() string {
	return n.SrcFile
}

// Line returns the line number where this node was defined.
func (n *BaseNode) Line() int {
	return n.LineNumber
}

//
// The AST nodes
//

// Storyworld is an AST mode representing the whole Storyworld, regardless of
// its structure in terms of files. See SourceFile for a more file-centric view.
//
// TODO: This doesn't support parallel compilation beyond parsing. Need to
// change to a more traditional compile / link process.
type Storyworld struct {
	BaseNode

	// Declarations stores all the declarations found in all source files that
	// compose the Storyworld.
	Declarations []Node
}

func (n *Storyworld) Type() TypeTag {
	return TypeVoid
}

func (n *Storyworld) Walk(v Visitor) {
	v.Enter(n)
	for _, decl := range n.Declarations {
		decl.Walk(v)
	}
	v.Leave(n)
}

// SourceFile is an AST node representing everything from a Romualdo source code
// file. It is the root of the formal AST (though in normal usage the romualdo
// tool will parse several files concurrently and move all parsed declarations
// to a single Storyworld node).
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

// ProcedureDecl is an AST node representing the declaration (and the
// definition, Romualdo doesn't have this distinction) of a Procedure. A
// Procedure can be either a Function or a Passage.
type ProcedureDecl struct {
	BaseNode

	// Kind tells if this is this a Function or a Passage. Important mainly for
	// error reporting, because internally it doesn't matter much.
	Kind ProcKind

	// Package is the absolute path of the package this Procedure belongs to.
	Package string

	// Name is the Procedure name.
	Name string

	// ReturnType contains the return type of this Procedure.
	ReturnType TypeTag

	// Parameters contains the parameters expected by this Procedure.
	Parameters []Parameter

	// Block contains the Procedure body (i.e., the statements that make it up).
	Body *Block

	//
	// Fields used for code generation
	//

	// ChunkIndex is the index into the array of Chunks where the bytecode for
	// this procedure is stored.
	ChunkIndex int
}

func (n *ProcedureDecl) Type() TypeTag {
	return TypeVoid
}

func (n *ProcedureDecl) Walk(v Visitor) {
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

// IfStmt is an AST node representing an if statement.
type IfStmt struct {
	BaseNode

	// Condition is the if condition.
	Condition Node

	// Then is the block of code executed if the condition is true.
	Then *Block

	// Else is the code executed if the condition is false. Can be either a
	// proper block or an `if` statement (in the case of an `elseif`). Might
	// also be nil (when no `else` block is present).
	Else Node
}

func (n *IfStmt) Type() TypeTag {
	return TypeVoid
}

func (n *IfStmt) Walk(v Visitor) {
	v.Enter(n)
	n.Condition.Walk(v)
	n.Then.Walk(v)
	if n.Else != nil {
		n.Else.Walk(v)
	}
	v.Leave(n)
}

// ExpressionStmt is an AST node representing an expression when used as a
// statement. In other words, this is an expression presumably used for its
// side-effects, given that the expression value is discarded.
type ExpressionStmt struct {
	BaseNode

	// Expr is the expression used as a statement.
	Expr Node
}

func (n *ExpressionStmt) Type() TypeTag {
	return TypeVoid
}

func (n *ExpressionStmt) Walk(v Visitor) {
	v.Enter(n)
	n.Expr.Walk(v)
	v.Leave(n)
}

// Lecture is an AST node representing a Lecture.
//
// A Lecture is an unorthodox language construct, that looks like a literal, but
// works like a statement. Essentially, a Lecture is automatically `say`d by the
// Storyworld when evaluated.
type Lecture struct {
	BaseNode

	// Text contains the Lecture's text.
	Text string
}

func (n *Lecture) Type() TypeTag {
	return TypeVoid
}

func (n *Lecture) Walk(v Visitor) {
	v.Enter(n)
	v.Leave(n)
}

// Listen is an AST node representing a Listen expression.
type Listen struct {
	BaseNode

	// Options contains the options for this listen expression.
	Options Node
}

func (n *Listen) Type() TypeTag {
	// TODO: Eventually, will be a map, right?
	return TypeString
}

func (n *Listen) Walk(v Visitor) {
	v.Enter(n)
	n.Options.Walk(v)
	v.Leave(n)
}

// BoolLiteral is an AST node representing a Boolean value literal.
type BoolLiteral struct {
	BaseNode

	// Value is the bool literal's value.
	Value bool
}

func (n *BoolLiteral) Type() TypeTag {
	return TypeBool
}

func (n *BoolLiteral) Walk(v Visitor) {
	v.Enter(n)
	v.Leave(n)
}

// StringLiteral is an AST node representing a string literal value.
type StringLiteral struct {
	BaseNode

	// Value is the string literal's value.
	Value string
}

func (n *StringLiteral) Type() TypeTag {
	return TypeString
}

func (n *StringLiteral) Walk(v Visitor) {
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

func (kind ProcKind) String() string {
	switch kind {
	case ProcKindFunction:
		return "Function"
	case ProcKindPassage:
		return "Passage"
	default:
		return fmt.Sprintf("<Unknown ProcKind: %v>", int(kind))
	}
}
