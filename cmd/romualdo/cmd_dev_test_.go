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

type testStep struct {
	Type          string
	SourceDir     string
	Input         []string
	Output        []string
	ExitCode      int
	ErrorMessages []string
}

type testConfig struct {
	Type          string
	SourceDir     string
	Input         []string
	Output        []string
	ExitCode      int
	ErrorMessages []string

	Steps []testStep `toml:"step"`
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
				return runTestCase(configPath, runner)
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

// runTestCase runs the test case defined in configPath using the given runner.
func runTestCase(configPath string, runner swRunnerFunc) errs.Error {
	testPath := path.Dir(configPath)
	testCase := testPath

	testConf, err := readTestConfig(configPath)
	if err != nil {
		return err
	}
	canonicalizeTestConfig(testConf)
	err = validateConfig(testCase, testConf)
	if err != nil {
		return err
	}

	for _, step := range testConf.Steps {
		srcPath := path.Join(testPath, step.SourceDir)
		outBuilder := &strings.Builder{}
		mouth := romutil.NewWriterMouth(outBuilder)
		ear := romutil.NewReaderEar(os.Stdin) // TODO: Must come from test config!

		err = runner(srcPath, mouth, ear)
		if err != nil {
			return errs.NewTestSuite(testCase, "running the storyworld: %v", err)
		}

		actualOut := outBuilder.String()
		if actualOut != testConf.Output[0] {
			errTS := errs.NewTestSuite(testCase, "expected output '%v', got '%v'.", testConf.Output[0], actualOut)
			return errTS
		}
	}

	fmt.Printf("Test case passed: %v.\n", testPath)
	return nil
}

// readTestConfig reads a test configuration from a TOML file.
func readTestConfig(path string) (*testConfig, errs.Error) {
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

// canonicalizeTestConfig makes sure testConf is in the canonical form.
// Specifically, it:
//
//   - Makes sure there is at least one element in Steps. (If there is no
//     explicit step defined, we create one with the data from the top-level
//     fields.)
//   - Makes sure all fields in all Steps have values: either the values
//     explicitly set or, or the values from the top-level fields, or the
//     default values.
func canonicalizeTestConfig(testConf *testConfig) {
	// Give default values to all empty fields in the top-level config.
	if testConf.Type == "" {
		testConf.Type = "build-and-run"
	}
	if testConf.SourceDir == "" {
		testConf.SourceDir = "src"
	}
	if testConf.Input == nil {
		testConf.Input = []string{}
	}
	if testConf.Output == nil {
		testConf.Output = []string{}
	}
	if testConf.ErrorMessages == nil {
		testConf.ErrorMessages = []string{}
	}

	// Make sure we have one step.
	if len(testConf.Steps) == 0 {
		testConf.Steps = append(testConf.Steps, testStep{
			Type:          testConf.Type,
			SourceDir:     testConf.SourceDir,
			Input:         testConf.Input,
			Output:        testConf.Output,
			ExitCode:      testConf.ExitCode,
			ErrorMessages: testConf.ErrorMessages,
		})
	}

	// Give values to all fields of all steps.
	for _, step := range testConf.Steps {
		if step.Type == "" {
			step.Type = testConf.Type
		}
		if step.SourceDir == "" {
			step.SourceDir = testConf.SourceDir
		}
		if step.Input == nil {
			step.Input = testConf.Input
		}
		if step.Output == nil {
			step.Output = testConf.Output
		}
		if step.ErrorMessages == nil {
			step.ErrorMessages = testConf.ErrorMessages
		}
		if step.ExitCode == 0 && testConf.ExitCode != 0 {
			step.ExitCode = testConf.ExitCode
		}
	}
}

// validateConfig validates a test configuration that is already in canonical
// format. Returns nil if the configuration is valid, or an error otherwise.
func validateConfig(testCase string, testConf *testConfig) errs.Error {
	for _, step := range testConf.Steps {
		// Validate step type
		if step.Type != "build-and-run" {
			return errs.NewTestSuite(testCase, "invalid test type '%v'; only 'build-and-run' supported for now", step.Type)
		}
	}
	return nil
}
