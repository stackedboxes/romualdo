/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2025 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package backend

import (
	"github.com/stackedboxes/romualdo/pkg/ast"
	"github.com/stackedboxes/romualdo/pkg/bytecode"
	"github.com/stackedboxes/romualdo/pkg/errs"
	"github.com/stackedboxes/romualdo/pkg/romutil"
)

// GenerateCode generates the bytecode for a given AST. The file name is used
// for error messages and debug information.
func GenerateCode(root ast.Node) (
	csw *bytecode.CompiledStoryworld,
	debugInfo *bytecode.DebugInfo,
	err errs.Error) {

	defer func() {
		if r := recover(); r != nil {
			csw = nil
			debugInfo = nil
			switch e := r.(type) {
			case *errs.CompileTime:
				err = e
				return
			case *errs.ICE:
				err = e
				return
			default:
				err = errs.NewICE("unexpected error type: %T (%v)", r, r)
				return
			}
		}
	}()

	// For the lack of better place at the moment, we'll hash the (source) code
	// here, before we generate (binary) code. When running the test suite, this
	// shall catch any Node type we forgot to handle -- but we don't do anything
	// useful with the hashes for now.
	codeHasher := romutil.NewCodeHasher()
	root.Walk(codeHasher)

	// Now we have the actual code generation.
	passOne := &codeGeneratorPassOne{
		codeGenerator: &codeGenerator{
			csw:                &bytecode.CompiledStoryworld{},
			debugInfo:          &bytecode.DebugInfo{},
			compilationContext: newCompilationContext(),
			nodeStack:          make([]ast.Node, 0, 64),
		},
	}
	root.Walk(passOne)

	if len(passOne.codeGenerator.nodeStack) > 0 {
		return nil, nil, errs.NewICE("node stack not empty between passes")
	}

	passTwo := &codeGeneratorPassTwo{
		codeGenerator: &codeGenerator{
			csw:                passOne.codeGenerator.csw,
			debugInfo:          passOne.codeGenerator.debugInfo,
			compilationContext: passOne.codeGenerator.compilationContext,
			nodeStack:          passOne.codeGenerator.nodeStack,
		},
		currentChunkIndex: -1, // start with an invalid value, for easier debugging
	}
	root.Walk(passTwo)
	return passTwo.codeGenerator.csw, passTwo.codeGenerator.debugInfo, nil
}
