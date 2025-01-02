/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2025 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

// The frontend package contains everything needed to transform Romualdo source
// code into an Abstract Syntax Tree (AST). Focus on the verb, "transform",
// because we are talking about the operations. The AST-related types are
// defined in the ast package.
//
// Highlights here are the scanner (lexer) and the parser.
package frontend
