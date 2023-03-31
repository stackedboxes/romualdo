/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2023 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package frontend

import (
	"fmt"
	"os"
	"regexp"

	"github.com/stackedboxes/romualdo/pkg/ast"
	"github.com/stackedboxes/romualdo/pkg/romutil"
)

// ParseStoryworld parses the Storyworld at a given root directory. It
// recursively looks for Romualdo source files (*.rd), parse each of them
// concurrently, and places all declarations into an ast.Storyworld. In case of
// errors, returns nil and prints the error messages to the standard error.
//
// TODO: Should probably return an error instead (maybe one that contains a list
// of errors). Parsing itself shouldn't print to stderr directly.
func ParseStoryworld(root string) *ast.Storyworld {
	sourceFiles, err := findRomualdoSourceFiles(root)
	if err != nil {
		return nil
	}

	chFiles := make(chan *ast.SourceFile, 1024)
	chError := make(chan error, 1024)

	for _, sourceFile := range sourceFiles {
		go parseFileAsync(sourceFile, chFiles, chError)
	}

	sw := &ast.Storyworld{}

	gotAnError := false
	for i := 0; i < len(sourceFiles); i++ {
		select {
		case sfNode := <-chFiles:
			sw.Declarations = append(sw.Declarations, sfNode.Declarations...)
		case err := <-chError:
			gotAnError = true
			// TODO: Odd error handling/reporting
			fmt.Fprintf(os.Stderr, "Parsing the Storyworld: %v\n", err)
		}
	}

	if gotAnError {
		return nil
	}
	return sw
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
		return nil, err
	}

	p := newParser(string(source))
	sfNode := p.parse()
	if sfNode == nil {
		// TODO: Bad error handling! (parse() should not print to stderr directly!)
		return nil, fmt.Errorf("Error parsing %q", path)
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
	}
	chFiles <- sfNode
}
