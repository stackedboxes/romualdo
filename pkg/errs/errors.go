/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2023 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package errs

import (
	"fmt"
	"strings"
)

//
// The Error interface
//

// Error is a Romualdo error.
type Error interface {
	error
	ExitCode() int
}

//
// CompileTime
//

// CompileTime is an error used to represent any compile-time error.
type CompileTime struct {
	// Message contains a user-friendly error message.
	Message string

	// FileName is the name of the file where the error was detected.
	FileName string

	// Line contains the line number where the error was detected.
	Line int

	// Lexeme contains the lexeme where the error was detected.
	Lexeme string
}

// NewCompileTime is a handy way to create a CompileTime error at some specific
// line of code.
func NewCompileTime(fileName string, line int, format string, a ...any) *CompileTime {
	return &CompileTime{
		Message:  fmt.Sprintf(format, a...),
		FileName: fileName,
		Line:     line,
	}
}

// NewCompileTimeWithoutLine is a handy way to create a CompileTime error that
// is not related with a specific line of code.
func NewCompileTimeWithoutLine(fileName, format string, a ...any) *CompileTime {
	return &CompileTime{
		Message:  fmt.Sprintf(format, a...),
		FileName: fileName,
		Line:     -1,
	}
}

// Error converts the CompileTime to a string. Fulfills the error interface.
func (e *CompileTime) Error() string {
	line := ""
	if e.Line > 0 {
		line = fmt.Sprintf(":%v", e.Line)
	}
	at := ""
	if e.Lexeme != "" {
		if e.Lexeme == "end of file" {
			at = fmt.Sprintf(" at %v", e.Lexeme)
		} else {
			at = fmt.Sprintf(" at `%v`", e.Lexeme)
		}
	}
	return fmt.Sprintf("%v%v%v: %v", e.FileName, line, at, e.Message)
}

// ExitCode fulfills the Error interface.
func (e *CompileTime) ExitCode() int {
	return statusCodeCompileTimeError
}

//
// CompileTimeCollection
//

// CompileTimeCollection is a collection of CompileTime errors.
type CompileTimeCollection struct {
	// Errors is the collection of CompileTime errors.
	Errors []*CompileTime
}

// Add adds a new error to the collection of errors. A no-op if err is nil.
func (e *CompileTimeCollection) Add(err *CompileTime) {
	if err == nil {
		return
	}
	e.Errors = append(e.Errors, err)
}

// AddMany adds all the errors in errs to e.
func (e *CompileTimeCollection) AddMany(errs *CompileTimeCollection) {
	e.Errors = append(e.Errors, errs.Errors...)
}

// IsEmpty checks if this CompileTimeCollection is empty (i.e., if it is collection
// of errors without any errors inside it).
func (e *CompileTimeCollection) IsEmpty() bool {
	return len(e.Errors) == 0
}

// Error converts the CompileTimeCollection to a string -- a multiline string at
// that, with one error per line. Fulfills the error interface.
func (e *CompileTimeCollection) Error() string {
	s := strings.Builder{}
	s.WriteString("Compile-time errors:\n")
	for _, err := range e.Errors {
		s.WriteString(err.Error())
		s.WriteByte('\n')
	}
	return s.String()
}

// ExitCode fulfills the Error interface.
func (e *CompileTimeCollection) ExitCode() int {
	return statusCodeCompileTimeError
}

//
// RomualdoTool
//

// RomualdoTool is an error that happened when running the romualdo tool that
// doesn't fit any of the other error types. Could be, e.g., an error opening
// some file.
type RomualdoTool struct {
	// Message contains a message explaining what went wrong.
	Message string
}

// NewRomualdoTool is a handy way to create a RomualdoTool error.
func NewRomualdoTool(format string, a ...any) *RomualdoTool {
	return &RomualdoTool{
		Message: fmt.Sprintf(format, a...),
	}
}

// Error converts the RomualdoTool to a string. Fulfills the error interface.
func (e *RomualdoTool) Error() string {
	return e.Message
}

// ExitCode fulfills the Error interface.
func (e *RomualdoTool) ExitCode() int {
	return statusCodeRomualdoToolError
}

//
// TestSuite
//

// TestSuite is an error that happened when running the Romualdo test suite
// (i.e.¸when testing Romualdo itself).
type TestSuite struct {
	// TestCase contains the path to the test case that failed.
	TestCase string

	// Message contains a message explaining how the test failed.
	Message string
}

// NewTestSuite is a handy way to create a TestSuite error.
func NewTestSuite(testCase, format string, a ...any) *TestSuite {
	return &TestSuite{
		TestCase: testCase,
		Message:  fmt.Sprintf(format, a...),
	}
}

// Error converts the TestSuite to a string. Fulfills the error interface.
func (e *TestSuite) Error() string {
	return fmt.Sprintf("%v: %v", e.TestCase, e.Message)
}

// ExitCode fulfills the Error interface.
func (e *TestSuite) ExitCode() int {
	return statusCodeTestSuiteError
}

//
// BadUsage
//

// BadUsage is an error that happened because the romualdo tool was called in
// the wrong way (like incorrect command-line arguments).
type BadUsage struct {
	// Message contains a message explaining what happened.
	Message string
}

// NewBadUsage is a handy way to create a BadUsage error.
func NewBadUsage(format string, a ...any) *BadUsage {
	return &BadUsage{
		Message: fmt.Sprintf(format, a...),
	}
}

// Error converts the BadUsage to a string. Fulfills the error interface.
func (e *BadUsage) Error() string {
	return "Usage error: " + e.Message
}

// ExitCode fulfills the Error interface.
func (e *BadUsage) ExitCode() int {
	return StatusCodeBadUsage
}

//
// Runtime
//

// Runtime is an error that happened while running a Storyworld. Not sure I am
// ready to vouch for this, but I think that this if this ever happens, it means
// there is a bug in the tooling.
type Runtime struct {
	// Message contains a message explaining what happened.
	Message string
}

// NewRuntime is a handy way to create a Runtime error.
func NewRuntime(format string, a ...any) *Runtime {
	return &Runtime{
		Message: fmt.Sprintf(format, a...),
	}
}

// Error converts the Runtime to a string. Fulfills the error interface.
func (e *Runtime) Error() string {
	return "Runtime error: " + e.Message
}

// ExitCode fulfills the Error interface.
func (e *Runtime) ExitCode() int {
	return statusCodeRuntimeError
}

//
// ICE
//

// ICE is an Internal Compiler Error. Used to report some unexpected issue with
// the compiler -- like when we find it is on a state it wasn't expected to be.
// It's always a bug.
type ICE struct {
	// Message contains some message to contextualize the situation in which the
	// error happened. Hopefully will be good enough to help fixing the bug.
	Message string
}

// NewICE is a handy way to create an ICE.
func NewICE(format string, a ...any) *ICE {
	return &ICE{
		Message: fmt.Sprintf(format, a...),
	}
}

// Error converts the ICE to a string. Fulfills the error interface.
func (e *ICE) Error() string {
	return "Internal Compiler Error: " + e.Message
}

// ExitCode fulfills the Error interface.
func (e *ICE) ExitCode() int {
	return statusCodeICE
}
