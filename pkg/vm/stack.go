/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2025 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package vm

import (
	"io"

	"github.com/stackedboxes/romualdo/pkg/bytecode"
	"github.com/stackedboxes/romualdo/pkg/errs"
	"github.com/stackedboxes/romualdo/pkg/romutil"
)

// Stack implements the VM runtime stack, which is a stack of bytecode.Values.
type Stack struct {
	data []bytecode.Value
}

// size returns the number of elements in the stack.
func (s *Stack) size() int {
	return len(s.data)
}

// top returns the value at the top of the stack, without popping it. Panics if
// the stack is empty.
func (s *Stack) top() bytecode.Value {
	return s.data[len(s.data)-1]
}

// push pushes a new value into the stack.
func (s *Stack) push(v bytecode.Value) {
	s.data = append(s.data, v)
}

// pop pops a value from the top of the stack and returns it. Panics on
// underflow.
func (s *Stack) pop() bytecode.Value {
	top := s.top()
	s.data = s.data[:len(s.data)-1]
	return top
}

// popN pops n values from the top of the stack and discards them. Panics on
// underflow.
func (s *Stack) popN(n int) {
	s.data = s.data[:len(s.data)-n]
}

// peek returns a value on the stack that is a given distance from the top.
// Passing 0 means "give me the value on the top of the stack". The stack is not
// changed at all. Panics if trying to get a value beyond the bottom of the
// stack.
func (s *Stack) peek(distance int) bytecode.Value {
	return s.data[len(s.data)-1-distance]
}

// at returns a value at a given index of the stack. In other words, accesses
// the stack as an array. The stack is not changed at all. Panics if trying to
// get a value that is out-of-bounds.
func (s *Stack) at(index int) bytecode.Value {
	return s.data[index]
}

// setAt sets the value at a given index of the stack. In other words, accesses
// the stack as an array. Panics if trying to set a value that is out-of-bounds.
func (s *Stack) setAt(index int, value bytecode.Value) {
	s.data[index] = value
}

// createView creates a read-write view into the Stack, so that the view looks
// like a new stack on top of the backing stack. The view stack will share
// offset elements with the backing stack. For example, passing 0 as the offset
// means that the view will be like a new, empty stack on top of the backing
// stack, without sharing any elements. Passing 1 means that the view will start
// with one single element (namely, the one that was on top of the backing
// stack).
func (s *Stack) createView(offset int) *StackView {
	return &StackView{
		stack: s,
		base:  s.size() - offset,
	}
}

// Serialize serializes the Stack to the given io.Writer.
func (s *Stack) Serialize(w io.Writer) errs.Error {
	err := romutil.SerializeU32(w, uint32(len(s.data)))
	if err != nil {
		return err
	}

	for _, v := range s.data {
		err = v.Serialize(w)
		if err != nil {
			return err
		}
	}
	return nil

}

// DeserializeStack deserializes a Stack from the given io.Reader.
func DeserializeStack(r io.Reader) (*Stack, errs.Error) {
	length, err := romutil.DeserializeU32(r)
	if err != nil {
		return nil, err
	}

	values := make([]bytecode.Value, length)
	for i := uint32(0); i < length; i++ {
		v, err := bytecode.DeserializeValue(r)
		if err != nil {
			return nil, err
		}
		values[i] = v
	}
	return &Stack{data: values}, nil
}

// StackView provides a read-write view into a Stack. It looks just like a
// Stack, but uses data owned by a real Stack, and behaves as if its base was at
// some arbitrary point within that backing Stack.
//
// It's assumed that all accesses made to this view are done while it is the
// topmost view created on the backing stack.
//
// This behavior matches the use case of call frames.
type StackView struct {
	stack *Stack
	base  int
}

// size returns the number of elements in the stack view.
func (s *StackView) size() int {
	return s.stack.size() - s.base
}

// top returns the value at the top of the stack view, without popping it.
func (s *StackView) top() bytecode.Value {
	return s.stack.top()
}

// push pushes a new value into the stack view.
func (s *StackView) push(v bytecode.Value) {
	s.stack.push(v)
}

// pop pops a value from the top of the stack view and returns it.
func (s *StackView) pop() bytecode.Value {
	return s.stack.pop()
}

// peek returns a value on the stack view that is a given distance from the top.
// Passing 0 means "give me the value on the top of the stack". The stack is not
// changed at all.
func (s *StackView) peek(distance int) bytecode.Value {
	return s.stack.peek(distance)
}

// at returns a value at a given index of the stack view. In other words,
// accesses the stack view as an array. The stack view is not changed at all.
func (s *StackView) at(index int) bytecode.Value {
	return s.stack.at(s.base + index)
}

// setAt sets the value at a given index of the stack view. In other words,
// accesses the stack view as an array.
func (s *StackView) setAt(index int, value bytecode.Value) {
	s.stack.setAt(s.base+index, value)
}
