/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2023 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package ast

// A Visitor has all the methods needed to traverse a Romualdo AST.
type Visitor interface {
	// Enter is called when entering a node during the traversal.
	Enter(node Node)

	// Leave is called when leaving a node during the traversal.
	Leave(node Node)
}
