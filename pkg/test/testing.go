/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2025 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package test

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"regexp"

	"github.com/pelletier/go-toml/v2"
	"github.com/stackedboxes/romualdo/pkg/errs"
	"github.com/stackedboxes/romualdo/pkg/frontend"
	"github.com/stackedboxes/romualdo/pkg/romutil"
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
	Hashes        map[string]string

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
	Hashes        map[string]string
}

// ExecuteSuite runs the test suite at suitePath.
func ExecuteSuite(suitePath string) errs.Error {
	// TODO: Run tests concurrently. Like we do Storyworld parsing.
	err := romutil.ForEachMatchingFileRecursive(suitePath, regexp.MustCompile("test.toml"),
		func(configPath string) errs.Error {
			return runCase(configPath)
		},
	)

	return err
}

// runCase runs the test case defined in configPath using the given runner.
func runCase(configPath string) errs.Error {
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

	var theVM *vm.VM
	var savedState []byte

	for i, step := range testConf.Steps {
		srcPath := path.Join(testPath, step.SourceDir)

		var story []string // the VM output
		var err errs.Error = nil

		switch step.Type {
		case "build":
			theVM, err = stepBuild(srcPath)

		case "run":
			err = stepRun(theVM, testCase, step.Input, &story)

		case "build-and-run":
			theVM, err = stepBuild(srcPath)
			if err != nil {
				return err
			}
			err = stepRun(theVM, testCase, step.Input, &story)

		case "save-state":
			bw := &bytes.Buffer{}
			err = theVM.Serialize(bw)
			if err != nil {
				return err
			}
			savedState = bw.Bytes()

		case "load-state":
			br := bytes.NewReader(savedState)
			err = theVM.Deserialize(br)
			if err != nil {
				return err
			}

		case "hash":
			err = stepHash(srcPath, testCase, step.Hashes)
			if err != nil {
				return err
			}

		default:
			return errs.NewTestSuite(testCase, "Unknown step type '%v'.", step.Type)
		}

		isLastStep := i == len(testConf.Steps)-1
		if isLastStep {
			// Check status code
			exitCode := 0
			if err != nil {
				exitCode = err.ExitCode()
			}
			if exitCode != step.ExitCode {
				return errs.NewTestSuite(testCase, "expected exit code %v, got %v.", step.ExitCode, err.ExitCode())
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

			// If we had errors and reached this point, it means the error was
			// expected for this test case. So we must skip the following
			// checks, as the outputs don't matter.
			if stepErrs != nil {
				break
			}
		} else {
			// We are not in the last step, so we expect no errors.
			if err != nil {
				return err
			}
		}

		// Check output
		if len(step.Output) != len(story) {
			return errs.NewTestSuite(testCase, "got %v outputs, expected %v.", len(story), len(step.Output))
		}
		for i, actualOutput := range story {
			if actualOutput != step.Output[i] {
				return errs.NewTestSuite(testCase, "at index %v: expected output '%v', got '%v'.", i, step.Output[0], actualOutput)
			}
		}
	}

	fmt.Printf("Test case passed: %v.\n", testPath)
	return nil
}

func stepBuild(srcPath string) (*vm.VM, errs.Error) {
	csw, di, err := vm.CSWFromPath(srcPath)
	if err != nil {
		return nil, err
	}
	theVM := vm.New(csw, di)
	return theVM, nil
}

func stepRun(theVM *vm.VM, testCase string, inputs []string, story *[]string) errs.Error {
	if theVM.State == vm.StateNew {
		output := theVM.Start()
		if output != "" {
			*story = append(*story, output)
		}
	}

	for _, choice := range inputs {
		if theVM.State == vm.StateEndOfStory {
			return errs.NewTestSuite(testCase, "Reached end of story but there are still unused inputs.")
		}

		if theVM.State != vm.StateWaitingForInput {
			return errs.NewICE("Inconsistent VM state: not waiting for input after Start() or Step()")
		}

		output := theVM.Step(choice)
		if output != "" {
			*story = append(*story, output)
		}
	}
	return nil
}

func stepHash(srcPath, testCase string, expectedHashes map[string]string) errs.Error {
	// Parse.
	swAST, err := frontend.ParseStoryworld(srcPath)
	if err != nil {
		return err
	}

	// Hash.
	ch := romutil.NewCodeHasher()
	swAST.Walk(ch)

	// Check.
	for sym, expHash := range expectedHashes {
		actualHash, found := ch.ProcedureHashes[sym]
		actualHashHex := fmt.Sprintf("%x", actualHash)
		if !found {
			return errs.NewTestSuite(testCase, "Symbol %v not found", sym)
		}
		if expHash != actualHashHex {
			return errs.NewTestSuite(testCase, "Wrong hash for %v: got %v, expected %v", sym, actualHashHex, expHash)
		}
	}

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
//     explicitly set, or the values from the top-level fields, or the
//     default values.
func canonicalizeConfig(testConf *config) {
	// Give default values to all empty fields in the top-level config.
	if testConf.Type == "" {
		if testConf.ExitCode != 0 && len(testConf.Steps) == 0 {
			testConf.Type = "build"
		} else {
			testConf.Type = "build-and-run"
		}
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
	if testConf.Hashes == nil {
		testConf.Hashes = map[string]string{}
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
			Hashes:        testConf.Hashes,
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
		if step.Hashes == nil && testConf.Hashes != nil {
			step.Hashes = testConf.Hashes
		}

		testConf.Steps[i] = step
	}
}

// validateConfig validates a test configuration that is already in canonical
// format. Returns nil if the configuration is valid, or an error otherwise.
func validateConfig(testCase string, testConf *config) errs.Error {
	var supportedTypes = map[string]bool{
		"build":         true,
		"build-and-run": true,
		"save-state":    true,
		"load-state":    true,
		"hash":          true,
	}
	for _, step := range testConf.Steps {
		// Validate step type
		if !supportedTypes[step.Type] {
			return errs.NewTestSuite(testCase, "invalid test step type `%v`", step.Type)
		}
	}
	return nil
}
