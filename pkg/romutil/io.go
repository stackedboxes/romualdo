/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2023 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package romutil

import (
	"bufio"
	"io"
	"os"
)

//
// Body parts
//

// A Mouth is something that can produce output for Romualdo. It is the
// abstraction representing how a Romualdo interpreter outputs data. A Mouth
// never returns an error, which is technically wrong but should be true enough
// for the uses cases that matter.
type Mouth interface {
	// Say outputs the given string.
	Say(string)
}

// An Ear is something that can receive input from Romualdo. It is the the
// abstraction representing how a Romualdo interpreter gets data from the
// outside world. An Ear never returns an error, which is technically wrong but
// should be true enough for the uses cases that matter.
type Ear interface {
	// Listen returns the next string from the input source.
	Listen() string
}

//
// writerMouth
//

// NewWriterMouth creates a new Mouth that outputs to the given io.Writer.
func NewWriterMouth(w io.Writer) Mouth {
	return &writerMouth{w}
}

// writerMouth is a Mouth that outputs to an io.Writer.
type writerMouth struct {
	w io.Writer
}

// Say outputs the given string to the underlying io.Writer.
func (wo *writerMouth) Say(s string) {
	// Ignore errors. Hopefully this will not be too bad for the envisioned use
	// cases (std output and in-memory buffers).
	_, _ = wo.w.Write([]byte(s))
}

//
// readerEar
//

// NewReaderEar creates a new Ear that reads from the given io.Reader.
func NewReaderEar(r io.Reader) Ear {
	s := bufio.NewScanner(r)
	return &readerEar{s}
}

// readerEar is an Ear that gets data from an io.Reader.
type readerEar struct {
	s *bufio.Scanner
}

// Listen returns the next string from the underlying io.Reader.
func (ri *readerEar) Listen() string {
	ri.s.Scan()
	return ri.s.Text()
}

//
// fatefulEar
//

// Returns an Ear that produces a predefined sequence of strings. It will
// produce, in sequence, each of the strings in inputs, and after that it will
// always produce an empty string.
func NewFatefulEar(inputs []string) Ear {
	return &fatefulEar{inputs}
}

// fatefulEar is an Ear that produces a fixed sequence of strings.
type fatefulEar struct {
	inputs []string
}

// Listen returns the next string from the fixed sequence of inputs.
func (fi *fatefulEar) Listen() string {
	// No more inputs; return an empty string.
	if len(fi.inputs) == 0 {
		return ""
	}

	// Return the next input.
	s := fi.inputs[0]
	fi.inputs = fi.inputs[1:]
	return s
}

//
// Helpers
//

// StdMouthAndEar returns a Mouth and an Ear that use the standard input and
// output.
func StdMouthAndEar() (Mouth, Ear) {
	return NewWriterMouth(os.Stdout), NewReaderEar(os.Stdin)
}
