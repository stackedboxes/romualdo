# Implementation

## Error handling strategy

Simple, just a little help from the type system to ensure we exit with the
proper status:

* Exported functions on all packages shall only return `err.Error`s.
    * These know their proper exit code.
    * And also allow testing code to know what value would be returned to the
      shell just by looking at the error.
* Commands use the `reportAndExitOnError()` and `reportAndExit()` helpers to
  handle error reporting.

## Adding something to the language

This is unverified, but let's see. In order to not have to relearn this for the
next I spend 5 months without looking at this code, here are some common steps
to make a language change:

* If a new token is needed:
    * In `pkg/frontend/token.go`: Add a new constant and update
      `TokenKind.String()` accordingly.
    * In `pkg/frontend/scanner.go`: Add new token to `lexemeToTokenKind`.
* If a new AST node is needed.
    * Add a new AST `Node` subtype at `pkg/ast/nodes.go`.
    * Change the parser at `pkg/frontend/parser.go`; some new function will
      return a node of this new type.
    * Implement how to interpret this node in `pkg/twi/interpreter.go`.
    * Generate code for this node type in `pkg/backend/pass_two.go`.
* Maybe add some new semantic checks at `pkg/frontend/semantic_checker.go`.
* Maybe add some new type checks at `pkg/frontend/type_checker.go`.
* Add the new AST node to the AST printer.
* If a new opcode is needed:
    * Document it at `doc/instruction_set.md`.
    * Add the opcode constant to `pkg/bytecode/opcodes.go`.
    * Emit this new opcode somewhere in `pkg/backend/pass_two.go`.
    * Add code to interpret it at `pkg/vm/vm.go`.
    * Add code to disassemble it in `pkg/bytecode/disassembler.go`.
