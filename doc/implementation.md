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

## Lecture x Code, Parser x Scanner

This section should be more complete, but for now here are some quick points
about consequences of having two different modes (Lecture and Code):

* The scanner and the parser are not completely independent. The parser changes
  the scanner behavior.
* How this happens is not always very clear, and probably not very consistent
  right now.
* In this regard, there are two important points in the Scanner API:
    * `Scanner.SetMode()`: The parser uses this to change between the two modes.
      For example, after seeing a `say` token, the parser switches the scanning
      mode to Lecture mode.
    * `Scanner.StartNewSpacePrefix()`: This is a bit trickier. The scanner
      handles the space prefix ignoring, but it doesn't know when a new space
      prefix needs to be used. For example, we start a new space prefix after a
      `say` statement, but not after Curlies. So, the parser tells the scanner
      when a new space prefix is due.
        * And notice that we have a stack of space prefixes, because we need to
          handle nested things like `say` statements inside double curlies.
