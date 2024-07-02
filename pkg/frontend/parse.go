/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2024 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package frontend

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"

	"github.com/stackedboxes/romualdo/pkg/ast"
	"github.com/stackedboxes/romualdo/pkg/errs"
	"github.com/stackedboxes/romualdo/pkg/romutil"
)

// ParseStoryworld parses the Storyworld at a given directory swRoot. It
// recursively looks for Romualdo source files (*.ral), parses each of them
// concurrently, and places all declarations into an ast.Storyworld.
func ParseStoryworld(swRoot string) (*ast.Storyworld, errs.Error) {
	sourceFiles, err := findRomualdoSourceFiles(swRoot)
	if err != nil {
		ctErr := errs.NewCompileTimeWithoutLine(swRoot, err.Error())
		return nil, ctErr
	}

	if len(sourceFiles) == 0 {
		ctErr := errs.NewCompileTimeWithoutLine(swRoot, "No Romualdo source files (*.ral) found.")
		return nil, ctErr
	}

	chFiles := make(chan *ast.SourceFile, 1024)
	chError := make(chan error, 1024)

	for _, sourceFile := range sourceFiles {
		go parseFileAsync(sourceFile, swRoot, chFiles, chError)
	}

	sw := &ast.Storyworld{}
	allErrors := &errs.CompileTimeCollection{}

	for i := 0; i < len(sourceFiles); i++ {
		select {
		case sfNode := <-chFiles:
			sw.Declarations = append(sw.Declarations, sfNode.Declarations...)
		case err := <-chError:
			compErrs := &errs.CompileTimeCollection{}
			if errors.As(err, &compErrs) {
				allErrors.AddMany(compErrs)
			} else {
				return nil, errs.NewICE("While parsing the Storyworld got an error of type %T: %v", err, err)
			}
		}
	}

	if !allErrors.IsEmpty() {
		return nil, allErrors
	}
	return sw, nil
}

// findRomualdoSourceFiles traverses the filesystem starting at swRoot looking for
// Romualdo source files (*.ral). Returns a slice with all files found.
func findRomualdoSourceFiles(swRoot string) ([]string, error) {
	files := []string{}
	err := romutil.ForEachMatchingFileRecursive(swRoot, regexp.MustCompile(`.*\.ral`),
		func(path string) errs.Error {
			files = append(files, path)
			return nil
		},
	)
	return files, err
}

// parseFile parses the Romualdo source file located at fileName and returns its
// corresponding AST. swRoot is the path to the root of the Storyworld, and is
// used to compute the file name relative to the Storyworld root.
func ParseFile(fileName, swRoot string) (*ast.SourceFile, errs.Error) {
	source, err := os.ReadFile(fileName)
	if err != nil {
		rtErr := errs.NewRomualdoTool("reading %v: %v", fileName, err)
		return nil, rtErr
	}

	fileNameFromSWRoot, err := filepath.Rel(swRoot, fileName)
	if err != nil {
		rtErr := errs.NewRomualdoTool("computing relative path of %v from %v: %v", fileName, swRoot, err)
		return nil, rtErr
	}
	fileNameFromSWRoot = filepath.Clean(fileNameFromSWRoot)

	p := newParser(fileNameFromSWRoot, string(source))
	sfNode, err := p.parse()
	if err != nil {
		ctErr := errs.NewCompileTimeWithoutLine(fileNameFromSWRoot, err.Error())
		return nil, ctErr
	}

	// Assorted semantic checks (but no type checks)
	sc := NewSemanticChecker(fileName)
	sfNode.Walk(sc)
	if !sc.errors.IsEmpty() {
		return nil, sc.errors
	}

	// Type checking
	tc := NewTypeChecker(fileName)
	sfNode.Walk(tc)
	if !tc.errors.IsEmpty() {
		return nil, tc.errors
	}

	return sfNode, nil
}

// parseFileAsync is a way to call ParseFile with everything wired up for being
// called asynchronously (i.e., it is not async by itself, but is designed to be
// called from a goroutine). It is guaranteed to send once to either one (but
// not both) of the channels it receives: either an error or the AST
// corresponding to the parsed file.
func parseFileAsync(path, swRoot string, chFiles chan<- *ast.SourceFile, chError chan<- error) {
	sfNode, err := ParseFile(path, swRoot)
	if err != nil {
		chError <- err
		return
	}
	chFiles <- sfNode
}
