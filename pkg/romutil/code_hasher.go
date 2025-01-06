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

	// Hashes stores the code hashes. Maps the fully-qualified symbol names to
	// their hashes.
	Hashes map[string]CodeHash
}

func NewCodeHasher() *CodeHasher {
	return &CodeHasher{
		hash:   sha256.New(),
		Hashes: make(map[string]CodeHash),
	}
}

// The Visitor interface
func (hasher *CodeHasher) Enter(node ast.Node) {
	switch n := node.(type) {

	case *ast.Binary:
		hasher.writeToken("(")

	case *ast.BoolLiteral:
		if n.Value {
			hasher.writeToken("true")
		} else {
			hasher.writeToken("false")
		}

	case *ast.Curlies:
		hasher.writeToken("}")

	case *ast.IfStmt:
		hasher.writeToken("if")

	case *ast.Lecture:
		hasher.writeToken(n.Text)

	case *ast.Listen:
		hasher.writeToken("listen")

	case *ast.ProcedureDecl:
		// Entering a brand new procedure, so reset the hash object.
		hasher.hash.Reset()

		// Start by writing the "function" or "passage" token.
		switch n.Kind {
		case ast.ProcKindFunction:
			hasher.writeToken("function")
		case ast.ProcKindPassage:
			hasher.writeToken("passage")
		default:
			panic("Unexpected procedure type")
		}

		// The procedure name.
		hasher.writeToken(n.Name)

		// Then the parameters.
		hasher.writeToken("(")
		for i, param := range n.Parameters {
			if i > 0 {
				hasher.writeToken(",")
			}

			hasher.writeToken(param.Name)
			hasher.writeToken(":")
			hasher.writeToken(typeStringFromTag(param.Type))
		}

		hasher.writeToken(")")

		// And finally, the return type.
		hasher.writeToken(":")
		hasher.writeToken(typeStringFromTag(n.ReturnType))

	case *ast.Say:
		hasher.writeToken("say")

	case *ast.StringLiteral:
		hasher.writeToken("\"" + n.Value + "\"")

	case *ast.Block, *ast.SourceFile, *ast.Storyworld:
		// Nothing to do!

	default:
		// This will cause tests to fail if we forget to handle some node type.
		panic(fmt.Sprintf("Unhandled Node type: %T", n))
	}
}

func (hasher *CodeHasher) Leave(node ast.Node) {
	switch n := node.(type) {

	case *ast.Binary:
		hasher.writeToken(")")

	case *ast.Curlies:
		hasher.writeToken("}")

	case *ast.IfStmt:
		hasher.writeToken("end")

	case *ast.ProcedureDecl:
		hasher.writeToken("end")

		fqn := n.Package + n.Name
		if _, exists := hasher.Hashes[fqn]; exists {
			panic(fmt.Sprintf("Duplicate symbol: `%v`", fqn))
		}
		hasher.Hashes[fqn] = CodeHash(hasher.hash.Sum(nil))

	case *ast.Block, *ast.BoolLiteral, *ast.Lecture, *ast.Listen, *ast.Say,
		*ast.SourceFile, *ast.Storyworld, *ast.StringLiteral:
		// Nothing to do!

	default:
		// This will cause tests to fail if we forget to handle some node type.
		panic(fmt.Sprintf("Unhandled Node type: %T", n))
	}
}

func (hasher *CodeHasher) Event(node ast.Node, event ast.EventType) {
	switch event {
	case ast.EventAfterIfCondition:
		hasher.writeToken("then")

	case ast.EventBeforeElse:
		hasher.writeToken("else")

	case ast.EventAfterBinaryLHS:
		bop, ok := node.(*ast.Binary)
		if !ok {
			panic(fmt.Sprintf("Expected a Binary AST node, got a %T", node))
		}
		hasher.writeToken(bop.Operator)
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
func (hasher *CodeHasher) writeToken(token string) {
	_, err := hasher.hash.Write([]byte(token))
	if err != nil {
		panic(err)
	}

	_, err = hasher.hash.Write([]byte{0})
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
