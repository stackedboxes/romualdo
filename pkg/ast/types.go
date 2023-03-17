/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2023 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package ast

// A TypeTag identifies a type as seen by the Romualdo Language (that is,
// ignoring the specificities of user-defined types).
type TypeTag int

const (
	// TypeInvalid is used to represent an invalid type. This is used internally
	// by the compiler, not something that would be ever found in a valid
	// Romualdo storyworld).
	TypeInvalid TypeTag = -1

	// TypeVoid identifies a void type (or rather non-type).
	TypeVoid = iota

	// TypeInt identifies an integer number type, AKA int.
	TypeInt

	// TypeFloat identifies a floating-point number type, AKA float.
	TypeFloat

	// TypeBNum identifies a bounded number number type, AKA bnum.
	TypeBNum

	// TypeBool identifies a Boolean type, AKA bool.
	TypeBool

	// TypeString identifies a string type.
	TypeString
)
