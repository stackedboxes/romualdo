/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2023 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package main

import (
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/cobra"
	"github.com/stackedboxes/romualdo/pkg/romutil"
	"github.com/stackedboxes/romualdo/pkg/twi"
)

type testConfig struct {
	ExpectedOutput []string
}

var devTestCmd = &cobra.Command{
	Use:   "test",
	Short: "Run a Romualdo test suite",
	Long:  `Run a Romualdo test suite (i.e., meant to test Romualdo itself).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Run tests concurrently. Like we do Storyworld parsing.
		err := romutil.ForEachMatchingFileRecursive(flagDevTestSuite, regexp.MustCompile("test.toml"),
			func(configPath string) error {
				testConf, err := readTestConfig(configPath)
				if err != nil {
					return err
				}

				testPath := path.Dir(configPath)
				return runTestCase(testPath, testConf)
			},
		)
		return err
	},
}

// flagDevTestSuite is the value of the `suite` flag of the `dev test` command.
var flagDevTestSuite string

func init() {
	devTestCmd.Flags().StringVarP(&flagDevTestSuite, "suite", "s",
		"./test", "Path to the test suite to run")
}

// readTestConfig reads a test configuration from a TOML file.
func readTestConfig(path string) (*testConfig, error) {
	tomlSource, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	tomlConfigData := &testConfig{}
	err = toml.Unmarshal(tomlSource, &tomlConfigData)
	if err != nil {
		return nil, err
	}

	return tomlConfigData, nil
}

// runTestCase runs the test case rooted at testPath, and whose configuration is
// testConf.
func runTestCase(testPath string, testConf *testConfig) error {
	// TODO: Add support fot interactivity.
	output := &strings.Builder{}
	err := twi.InterpretStoryworld(testPath, output)
	if err != nil {
		return err
	}

	actualOut := output.String()
	if actualOut != testConf.ExpectedOutput[0] {
		// TODO: Need better error handling and reporting!
		return fmt.Errorf("Error on test %q: Expected output %q, got %q.", testPath, actualOut, testConf.ExpectedOutput[0])
	}
	return nil
}
