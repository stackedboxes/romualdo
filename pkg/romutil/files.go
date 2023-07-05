/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2023 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package romutil

import (
	"os"
	"path"
	"regexp"

	"github.com/stackedboxes/romualdo/pkg/errs"
)

// ForEachMatchingFileRecursive recursively traverses the filesystem from root,
// and calls action on every file found that matches pattern. Only the file name
// alone (not the full path) is used for pattern matching.
func ForEachMatchingFileRecursive(root string, pattern *regexp.Regexp, action func(path string) errs.Error) errs.Error {
	items, plainErr := os.ReadDir(root)
	if plainErr != nil {
		return errs.NewRomualdoTool("reading directory %v: %v", root, plainErr)
	}
	for _, item := range items {
		itemPath := path.Join(root, item.Name())
		if item.IsDir() {
			err := ForEachMatchingFileRecursive(itemPath, pattern, action)
			if err != nil {
				return err
			}
		} else {
			if pattern.Match([]byte(item.Name())) {
				err := action(itemPath)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// IsDir checks if path exists and is a directory.
func IsDir(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}

	return fileInfo.IsDir(), nil
}
