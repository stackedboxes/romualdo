/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2025 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package ast

// EventType identifies "events" that happen during the tree traversal. These
// events allow us to write code that gets executed at specific points during
// the traversal -- beyond the usual "enter" and "leave" points.
type EventType int

const (
	// EventAfterIfCondition is emitted right after the condition of an "if"
	// statement has been visited.
	EventAfterIfCondition EventType = iota

	// EventAfterThenBlock is emitted wight after the "then" block (that is, the
	// block executed when the "if" condition is true) has been visited.
	EventAfterThenBlock

	// EventBeforeElse is emitted right before we visit the "else" part of an
	// "if" statement. This is not emitted for "if" statements that don't have
	// an "else".
	EventBeforeElse

	// EventAfterElse is emitted right after we visit the "else" part of an "if"
	// statement. This is not emitted for "if" statements that don't have an
	// "else".
	EventAfterElse
)

// A Visitor has all the methods needed to traverse a Romualdo AST.
type Visitor interface {
	// Enter is called when entering a node during the traversal.
	Enter(node Node)

	// Event is called for special "events" during the tree traversal that we
	// might need to handle specially. The event argument is one of the Event*
	// constants.
	Event(node Node, event EventType)

	// Leave is called when leaving a node during the traversal.
	Leave(node Node)
}
