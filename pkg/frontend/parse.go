/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2023 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package frontend

import (
	"errors"
	"os"
	"regexp"

	"github.com/stackedboxes/romualdo/pkg/ast"
	"github.com/stackedboxes/romualdo/pkg/errs"
	"github.com/stackedboxes/romualdo/pkg/romutil"
)

// ParseStoryworld parses the Storyworld at a given root directory. It
// recursively looks for Romualdo source files (*.rd), parses each of them
// concurrently, and places all declarations into an ast.Storyworld.
func ParseStoryworld(root string) (*ast.Storyworld, error) {
	sourceFiles, err := findRomualdoSourceFiles(root)
	if err != nil {
		ctErr := errs.NewGenericCompileTime(root, err.Error())
		return nil, ctErr
	}

	if len(sourceFiles) == 0 {
		ctErr := errs.NewGenericCompileTime(root, "No Romualdo source files (*.rd) found.")
		return nil, ctErr
	}

	chFiles := make(chan *ast.SourceFile, 1024)
	chError := make(chan error, 1024)

	for _, sourceFile := range sourceFiles {
		go parseFileAsync(sourceFile, chFiles, chError)
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

// findRomualdoSourceFiles traverses the filesystem starting at root looking for
// Romualdo source files (*.rd). Returns a slice with all files found.
func findRomualdoSourceFiles(root string) ([]string, error) {
	files := []string{}
	err := romutil.ForEachMatchingFileRecursive(root, regexp.MustCompile(`.*\.rd`),
		func(path string) error {
			files = append(files, path)
			return nil
		},
	)
	return files, err
}

// parseFile parses the Romualdo source file located at path and returns its
// corresponding AST.
func ParseFile(path string) (*ast.SourceFile, error) {
	source, err := os.ReadFile(path)
	if err != nil {
		ctErr := errs.NewGenericCompileTime(path, err.Error())
		return nil, ctErr
	}

	p := newParser(path, string(source))
	sfNode, err := p.parse()
	if err != nil {
		return nil, err
	}

	// Assorted semantic checks (but no type checks)
	sc := NewSemanticChecker(path)
	sfNode.Walk(sc)
	if !sc.errors.IsEmpty() {
		return nil, sc.errors
	}

	return sfNode, nil
}

// parseFileAsync is a way to call ParseFile with everything wired up for being
// called asynchronously (i.e., it is not async by itself, but is designed to be
// called from a goroutine). It is guaranteed to send once to either one (but
// not both) of the channels it receives: either an error or the AST
// corresponding to the parsed file.
func parseFileAsync(path string, chFiles chan<- *ast.SourceFile, chError chan<- error) {
	sfNode, err := ParseFile(path)
	if err != nil {
		chError <- err
		return
	}
	chFiles <- sfNode
}
