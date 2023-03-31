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
)

// ForEachMatchingFileRecursive recursively traverses the filesystem from root,
// and calls action on every file found that matches pattern. Only the file name
// alone (not the full path) is used for pattern matching.
func ForEachMatchingFileRecursive(root string, pattern *regexp.Regexp, action func(path string) error) error {
	items, err := os.ReadDir(root)
	if err != nil {
		return err
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
				err = action(itemPath)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
