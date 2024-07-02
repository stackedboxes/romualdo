/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2024 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package backend

import (
	"github.com/stackedboxes/romualdo/pkg/ast"
	"github.com/stackedboxes/romualdo/pkg/bytecode"
)

// TODO: Probably rename to something meaningful. create_procedures_pass?

// codeGeneratorPassOne creates the Chunks where the bytecode will be eventually
// written to.
//
// This implements the ast.Visitor interface.
type codeGeneratorPassOne struct {
	codeGenerator *codeGenerator
}

//
// The ast.Visitor interface
//

func (cg *codeGeneratorPassOne) Enter(node ast.Node) {
	if _, ok := node.(*ast.Block); ok {
		cg.codeGenerator.beginScope()
	}
	if cg.codeGenerator.scopeDepth > 0 {
		return
	}

	switch n := node.(type) {

	case *ast.ProcedureDecl:
		csw := cg.codeGenerator.csw
		di := cg.codeGenerator.debugInfo
		cc := cg.codeGenerator.compilationContext

		n.ChunkIndex = len(cg.codeGenerator.csw.Chunks)
		newChunk := &bytecode.Chunk{}
		csw.Chunks = append(csw.Chunks, newChunk)
		di.ChunksNames = append(di.ChunksNames, n.Name)
		di.ChunksSourceFiles = append(di.ChunksSourceFiles, n.SourceFile())
		di.ChunksLines = append(di.ChunksLines, []int{})

		fqn := n.Package + n.Name
		if _, exists := cc.procNameToIndex[fqn]; exists {
			cg.codeGenerator.ice("duplicate definition of procedure name '%v' during pass one",
				n.Name)
		}
		cc.procNameToIndex[fqn] = n.ChunkIndex
	}
}

func (cg *codeGeneratorPassOne) Leave(node ast.Node) {
	if _, ok := node.(*ast.Block); ok {
		cg.codeGenerator.endScope()
	}

	if cg.codeGenerator.scopeDepth > 0 {
		return
	}
}

func (cg *codeGeneratorPassOne) Event(node ast.Node, event ast.EventType) {
	// Nothing
}
