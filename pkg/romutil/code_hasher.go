/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2025 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package romutil

import (
	"crypto/sha256"
	"fmt"
	"hash"

	"github.com/stackedboxes/romualdo/pkg/ast"
)

// CodeHash can store the hash of some code bit.
type CodeHash [sha256.Size]byte

// CodeHasher is a node visitor that computes the hash of procedures and
// globals.
//
// Hashing is used to detect meaningful changes to code -- changes that require
// a new version of the procedure to be created.
//
// TODO: This operates at source file level. It should operate at package level,
// or storyworld level.
type CodeHasher struct {
	// hash is the Hash object used to hash the code contents.
	hash hash.Hash

	// ProcedureHashes stores the code hashes for the procedures. Maps the
	// fully-qualified procedure names to their hashes.
	ProcedureHashes map[string]CodeHash

	// GlobalHashes stores the code hashes for the globals. Maps the
	// fully-qualified global names to their hashes.
	GlobalHashes map[string]CodeHash
}

func NewCodeHasher() *CodeHasher {
	return &CodeHasher{
		hash:            sha256.New(),
		ProcedureHashes: make(map[string]CodeHash),
		GlobalHashes:    make(map[string]CodeHash),
	}
}

// The Visitor interface
func (ch *CodeHasher) Enter(node ast.Node) {
	switch n := node.(type) {

	case *ast.Binary:
		ch.writeToken("(")

	case *ast.BoolLiteral:
		if n.Value {
			ch.writeToken("true")
		} else {
			ch.writeToken("false")
		}

	case *ast.Curlies:
		ch.writeToken("}")

	case *ast.IfStmt:
		ch.writeToken("if")

	case *ast.Lecture:
		ch.writeToken(n.Text)

	case *ast.Listen:
		ch.writeToken("listen")

	case *ast.ProcedureDecl:
		// Entering a brand new procedure, so reset the hash object.
		ch.hash.Reset()

		// Start by writing the "function" or "passage" token.
		switch n.Kind {
		case ast.ProcKindFunction:
			ch.writeToken("function")
		case ast.ProcKindPassage:
			ch.writeToken("passage")
		default:
			panic("Unexpected procedure type")
		}

		// The procedure name.
		ch.writeToken(n.Name)

		// Then the parameters.
		ch.writeToken("(")
		for i, param := range n.Parameters {
			if i > 0 {
				ch.writeToken(",")
			}

			ch.writeToken(param.Name)
			ch.writeToken(":")
			ch.writeToken(typeStringFromTag(param.Type))
		}

		ch.writeToken(")")

		// And finally, the return type.
		ch.writeToken(":")
		ch.writeToken(typeStringFromTag(n.ReturnType))

	case *ast.Say:
		ch.writeToken("say")

	case *ast.StringLiteral:
		ch.writeToken("\"" + n.Value + "\"")

	case *ast.Block, *ast.SourceFile, *ast.Storyworld:
		// Nothing to do!

	default:
		// This will cause tests to fail if we forget to handle some node type.
		panic(fmt.Sprintf("Unhandled Node type: %T", n))
	}
}

func (ch *CodeHasher) Leave(node ast.Node) {
	switch n := node.(type) {

	case *ast.Binary:
		ch.writeToken(")")

	case *ast.Curlies:
		ch.writeToken("}")

	case *ast.IfStmt:
		ch.writeToken("end")

	case *ast.ProcedureDecl:
		ch.writeToken("end")

		fqn := n.Package + n.Name
		if _, exists := ch.ProcedureHashes[fqn]; exists {
			panic(fmt.Sprintf("Duplicate procedure: `%v`", fqn))
		}
		ch.ProcedureHashes[fqn] = CodeHash(ch.hash.Sum(nil))

	case *ast.Block, *ast.BoolLiteral, *ast.Lecture, *ast.Listen, *ast.Say,
		*ast.SourceFile, *ast.Storyworld, *ast.StringLiteral:
		// Nothing to do!

	default:
		// This will cause tests to fail if we forget to handle some node type.
		panic(fmt.Sprintf("Unhandled Node type: %T", n))
	}
}

func (ch *CodeHasher) Event(node ast.Node, event ast.EventType) {
	switch event {
	case ast.EventAfterIfCondition:
		ch.writeToken("then")

	case ast.EventBeforeElse:
		ch.writeToken("else")

	case ast.EventAfterBinaryLHS:
		bop, ok := node.(*ast.Binary)
		if !ok {
			panic(fmt.Sprintf("Expected a Binary AST node, got a %T", node))
		}
		ch.writeToken(bop.Operator)
	}
}

// Writes a token so that it gets hashed.
//
// Notice that we add a zero byte after the string representation of the token
// to avoid ambiguous results. This is more of an extra precaution, as the only
// example that comes to mind doesn't really happen in practice: preventing the
// sequence of tokens "else" and "if" to have the same hash as the single token
// "elseif". (The codeHasher doesn't generate "elseif" tokens, only separate
// "else" and "if" ones that's why this case can't happen in practice.)
func (ch *CodeHasher) writeToken(token string) {
	_, err := ch.hash.Write([]byte(token))
	if err != nil {
		panic(err)
	}

	_, err = ch.hash.Write([]byte{0})
	if err != nil {
		panic(err)
	}
}

// typeStringFromTag is a quick and dirty conversion function to obtain the
// string representation of a type.
//
// TODO: This supports only built-in type (via the type tag), but eventually
// we'll need to support user-defined types. On this day, we should move this
// code to the ast package (or something like it), as we'll need to use it in
// more places (like for naming types in error reporting).
func typeStringFromTag(tag ast.TypeTag) string {
	switch tag {
	case ast.TypeVoid:
		return "void"
	case ast.TypeInt:
		return "int"
	case ast.TypeFloat:
		return "float"
	case ast.TypeBNum:
		return "bnum"
	case ast.TypeBool:
		return "bool"
	case ast.TypeString:
		return "string"
	default:
		panic(fmt.Sprintf("Unexpected type tag: %T (%v)", tag, tag))
	}
}
