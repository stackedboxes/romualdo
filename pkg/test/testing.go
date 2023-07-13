/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2023 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package test

import (
	"fmt"
	"os"
	"path"
	"regexp"

	"github.com/pelletier/go-toml/v2"
	"github.com/stackedboxes/romualdo/pkg/errs"
	"github.com/stackedboxes/romualdo/pkg/romutil"
	"github.com/stackedboxes/romualdo/pkg/twi"
	"github.com/stackedboxes/romualdo/pkg/vm"
)

// config is the structure mirroring the test case TOML file.
type config struct {
	Type          string
	SourceDir     string
	Input         []string
	Output        []string
	ExitCode      int
	ErrorMessages []string

	Steps []step `toml:"step"`
}

// step is the structure mirroring a single step in a test case TOML file.
type step struct {
	Type          string
	SourceDir     string
	Input         []string
	Output        []string
	ExitCode      int
	ErrorMessages []string
}

// swRunnerFunc is a function that can run a Storyworld at path, using mouth and
// ear for I/O.
type swRunnerFunc func(path string, mouth romutil.Mouth, ear romutil.Ear) errs.Error

// ExecuteSuite runs the test suite at suitePath. If walkDontRun is true, the
// tree-walk interpreter is used instead of the bytecode one.
func ExecuteSuite(walkDontRun bool, suitePath string) errs.Error {
	// TODO: Run tests concurrently. Like we do Storyworld parsing.

	var runner swRunnerFunc = nil

	if walkDontRun {
		fmt.Println("Using the tree-walk interpreter.")
		runner = twi.WalkStoryworld
	} else {
		fmt.Println("Using the bytecode interpreter.")
		runner = func(path string, mouth romutil.Mouth, ear romutil.Ear) errs.Error {
			return vm.RunStoryworld(path, mouth, ear, false)
		}
	}

	err := romutil.ForEachMatchingFileRecursive(suitePath, regexp.MustCompile("test.toml"),
		func(configPath string) errs.Error {
			return runCase(configPath, runner)
		},
	)

	return err
}

// runCase runs the test case defined in configPath using the given runner.
func runCase(configPath string, runner swRunnerFunc) errs.Error {
	testPath := path.Dir(configPath)
	testCase := testPath

	testConf, err := readConfig(configPath)
	if err != nil {
		return err
	}
	canonicalizeConfig(testConf)
	err = validateConfig(testCase, testConf)
	if err != nil {
		return err
	}

	for _, step := range testConf.Steps {
		srcPath := path.Join(testPath, step.SourceDir)
		mouth := &romutil.MemoryMouth{}
		ear := romutil.NewFatefulEar(step.Input)

		err = runner(srcPath, mouth, ear)

		// Check status code
		if err != nil {
			if err.ExitCode() != step.ExitCode {
				return errs.NewTestSuite(testCase, "expected exit code %v, got %v.", step.ExitCode, err.ExitCode())
			}
		}

		// Check error messages
		stepErrs := err
		for _, expectedErrMsg := range step.ErrorMessages {
			re, err := regexp.Compile(expectedErrMsg)
			if err != nil {
				return errs.NewTestSuite(testCase, "compiling regexp '%v': %v.", expectedErrMsg, err.Error())
			}

			if !re.Match([]byte(stepErrs.Error())) {
				return errs.NewTestSuite(testCase, "expected error message '%v', got '%v'.", expectedErrMsg, stepErrs.Error())
			}
		}

		if stepErrs != nil {
			// If we had errors and reached this point, it means the error was
			// expected. The outputs don't matter, go on to the next step.
			continue
		}

		// Check output
		if len(step.Output) != len(mouth.Outputs) {
			return errs.NewTestSuite(testCase, "got %v outputs, expected %v.", len(mouth.Outputs), len(step.Output))
		}
		for i, actualOutput := range mouth.Outputs {
			if actualOutput != step.Output[i] {
				return errs.NewTestSuite(testCase, "at index %v: expected output '%v', got '%v'.", i, step.Output[0], actualOutput)
			}
		}
	}

	fmt.Printf("Test case passed: %v.\n", testPath)
	return nil
}

// readConfig reads a test configuration from a TOML file.
func readConfig(path string) (*config, errs.Error) {
	tomlSource, err := os.ReadFile(path)
	if err != nil {
		tsErr := errs.NewTestSuite(path, "%v", err.Error())
		return nil, tsErr
	}
	tomlConfigData := &config{}
	err = toml.Unmarshal(tomlSource, &tomlConfigData)
	if err != nil {
		tsErr := errs.NewTestSuite(path, "%v", err.Error())
		return nil, tsErr
	}

	return tomlConfigData, nil
}

// canonicalizeConfig makes sure testConf is in the canonical form.
// Specifically, it:
//
//   - Makes sure there is at least one element in Steps. (If there is no
//     explicit step defined, we create one with the data from the top-level
//     fields.)
//   - Makes sure all fields in all Steps have values: either the values
//     explicitly set or, or the values from the top-level fields, or the
//     default values.
func canonicalizeConfig(testConf *config) {
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
		testConf.Steps = append(testConf.Steps, step{
			Type:          testConf.Type,
			SourceDir:     testConf.SourceDir,
			Input:         testConf.Input,
			Output:        testConf.Output,
			ExitCode:      testConf.ExitCode,
			ErrorMessages: testConf.ErrorMessages,
		})
	}

	// Give values to all fields of all steps.
	for i, step := range testConf.Steps {
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

		testConf.Steps[i] = step
	}
}

// validateConfig validates a test configuration that is already in canonical
// format. Returns nil if the configuration is valid, or an error otherwise.
func validateConfig(testCase string, testConf *config) errs.Error {
	for _, step := range testConf.Steps {
		// Validate step type
		if step.Type != "build-and-run" {
			return errs.NewTestSuite(testCase, "invalid test type '%v'; only 'build-and-run' supported for now", step.Type)
		}
	}
	return nil
}
