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
	"github.com/stackedboxes/romualdo/pkg/errs"
	"github.com/stackedboxes/romualdo/pkg/romutil"
	"github.com/stackedboxes/romualdo/pkg/twi"
	"github.com/stackedboxes/romualdo/pkg/vm"
)

type testConfig struct {
	ExpectedOutput []string
}

// swRunnerFunc is a function that can run a Storyworld at path, using mouth and
// ear for I/O.
type swRunnerFunc func(path string, mouth romutil.Mouth, ear romutil.Ear) errs.Error

var devTestCmd = &cobra.Command{
	Use:   "test",
	Short: "Run a Romualdo test suite",
	Long:  `Run a Romualdo test suite (i.e., meant to test Romualdo itself).`,
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Run tests concurrently. Like we do Storyworld parsing.

		var runner swRunnerFunc = nil

		if flagDevTestWalkDontRun {
			fmt.Println("Using the tree-walk interpreter.")
			runner = twi.WalkStoryworld
		} else {
			fmt.Println("Using the bytecode interpreter.")
			runner = func(path string, mouth romutil.Mouth, ear romutil.Ear) errs.Error {
				return vm.RunStoryworld(path, mouth, ear, false)
			}
		}

		err := romutil.ForEachMatchingFileRecursive(flagDevTestSuite, regexp.MustCompile("test.toml"),
			func(configPath string) errs.Error {
				testConf, err := readTestConfig(configPath)
				if err != nil {
					rtErr := errs.NewRomualdoTool("reading test config file: %v", err)
					return rtErr
				}

				testPath := path.Dir(configPath)
				srcPath := path.Join(testPath, "src")
				outBuilder := &strings.Builder{}
				mouth := romutil.NewWriterMouth(outBuilder)
				ear := romutil.NewReaderEar(os.Stdin) // TODO: Must come from test config!

				err = runner(srcPath, mouth, ear)
				if err != nil {
					return errs.NewTestSuite(testPath, "running the storyworld: %v", err)
				}

				actualOut := outBuilder.String()
				if actualOut != testConf.ExpectedOutput[0] {
					errTS := errs.NewTestSuite(testPath, "expected output '%v', got '%v'.", testConf.ExpectedOutput[0], actualOut)
					return errTS
				}

				fmt.Printf("Test case passed: %v.\n", testPath)
				return nil
			},
		)
		reportAndExit(err)
	},
}

// flagDevTestSuite is the value of the --suite flag of the `dev test` command.
var flagDevTestSuite string

// flagDevTestWalkDontRun is the value of the --walk-dont-run flag of the `dev
// test` command.
var flagDevTestWalkDontRun bool

func init() {
	devTestCmd.Flags().StringVarP(&flagDevTestSuite, "suite", "s",
		"./test", "Path to the test suite to run")

	devTestCmd.Flags().BoolVarP(&flagDevTestWalkDontRun, "walk-dont-run", "w",
		false, "Test using the walk tree interpreter instead of the bytecode one")
}

// readTestConfig reads a test configuration from a TOML file.
func readTestConfig(path string) (*testConfig, error) {
	tomlSource, err := os.ReadFile(path)
	if err != nil {
		tsErr := errs.NewTestSuite(path, "%v", err.Error())
		return nil, tsErr
	}
	tomlConfigData := &testConfig{}
	err = toml.Unmarshal(tomlSource, &tomlConfigData)
	if err != nil {
		tsErr := errs.NewTestSuite(path, "%v", err.Error())
		return nil, tsErr
	}

	return tomlConfigData, nil
}
