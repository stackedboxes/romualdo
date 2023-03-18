/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2023 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package frontend

import (
	"fmt"
	"os"

	"github.com/stackedboxes/romualdo/pkg/ast"
)

// parser is a parser for the Romualdo language. It converts source code into an
// AST.
type parser struct {
	// currentToken is the current token we are parsing.
	currentToken *Token

	// previousToken is the previous token we have parsed.
	previousToken *Token

	// hadError indicates whether we found at least one syntax error.
	hadError bool

	// panicMode indicates whether we are in panic mode. This has nothing to do
	// with Go panics. Right after finding a syntax error it is hard to generate
	// good error messages because the parser is "out of sync" with the code, so
	// we enter panic mode (during which we don't report any errors). Once we
	// find a "synchronization point", we leave panic mode.
	panicMode bool

	// scanner is the Scanner from where we get our tokens.
	scanner *Scanner
}

// newParser returns a new parser that will parse source.
func newParser(source string) *parser {
	return &parser{
		scanner: NewScanner(source),
	}
}

// parse parses source and returns the root of the resulting AST. Returns nil in
// case of error.
func (p *parser) parse() *ast.SourceFile {
	sf := &ast.SourceFile{}

	p.advance()

	for !p.match(TokenKindEOF) {
		decl := p.declaration()
		if p.hadError {
			return nil
		}

		sf.Declarations = append(sf.Declarations, decl)
	}

	return sf
}

//
// Parsing building blocks
//

// advance advances the parser by one token. This will report errors for each
// error token found; callers will only see the non-error tokens.
func (p *parser) advance() {
	p.previousToken = p.currentToken

	for {
		p.currentToken = p.scanner.Token()
		if p.currentToken.Kind != TokenKindError {
			break
		}

		p.errorAtCurrent(p.currentToken.Lexeme)
	}
}

// check checks if the current token is of a given kind.
func (p *parser) check(kind TokenKind) bool {
	return p.currentToken.Kind == kind
}

// match consumes the current token if it is of a given type and returns true;
// otherwise, it simply returns false without consuming any token.
func (p *parser) match(kind TokenKind) bool {
	if !p.check(kind) {
		return false
	}
	p.advance()
	return true
}

// consume consumes the current token (and advances the parser), assuming it is
// of a given kind. If it is not of this kind, reports this is an error with a
// given error message.
func (p *parser) consume(kind TokenKind, message string) {
	if p.currentToken.Kind == kind {
		p.advance()
		return
	}

	p.errorAtCurrent(message)
}

//
// Parsing of grammar rules (thinks that return Nodes)
//

// Parses any kind of top-level declaration, like functions and passages.
func (p *parser) declaration() ast.Node {
	if p.match(TokenKindFunction) {
		return p.functionDecl()
	} else if p.match(TokenKindPassage) {
		return p.passageDecl()
	} else {
		p.errorAtCurrent("Expected a declaration.")
		return nil
	}
}

// functionDecl parses a function declaration. The "function" token must have
// been just consumed.
func (p *parser) functionDecl() *ast.ProcDecl {
	proc := &ast.ProcDecl{
		Kind: ast.ProcKindFunction,
	}

	p.consume(TokenKindIdentifier, "Expected the function name.")
	proc.Name = p.previousToken.Lexeme

	p.consume(
		TokenKindLeftParen,
		fmt.Sprintf("Expected '(' after the function name %q.", proc.Name))
	proc.Parameters = p.parseParameterList()
	p.consume(TokenKindColon, "Expected ':' after parameter list.")

	proc.ReturnType = p.parseType()

	proc.Body = p.block()

	return proc
}

// passageDecl parses a passage declaration. The "passage" token must have been
// just consumed.
func (p *parser) passageDecl() *ast.ProcDecl {
	proc := &ast.ProcDecl{
		Kind: ast.ProcKindPassage,
	}

	p.consume(TokenKindIdentifier, "Expected the passage name.")
	proc.Name = p.previousToken.Lexeme

	p.consume(
		TokenKindLeftParen,
		fmt.Sprintf("Expected '(' after the passage name %q.", proc.Name))
	proc.Parameters = p.parseParameterList()
	p.consume(TokenKindColon, "Expected ':' after parameter list.")

	// Consume the last token of the type only after switching to text mode,
	// because we want advance() to parse the next token already in text mode.
	proc.ReturnType = p.parseTypeNoConsume()
	p.scanner.SetMode(ScannerModeText)
	p.advance()

	// As above, make sure we switch back to code mode before parsing the first
	// token after the block.
	proc.Body = p.blockNoConsume()
	p.scanner.SetMode(ScannerModeCode)
	p.advance()

	return proc
}

// block parses a block of code. Whatever keyword(s) started the block should
// have been just consumed.
func (p *parser) block() *ast.Block {
	result := p.blockNoConsume()
	p.advance()
	return result
}

// blockNoConsume is like block, but doesn't consume the token that closes the
// block.
func (p *parser) blockNoConsume() *ast.Block {
	block := &ast.Block{}

	blockLine := p.previousToken.Line

	for !p.check(TokenKindEnd) && !p.check(TokenKindEOF) {
		stmt := p.statement()
		block.Statements = append(block.Statements, stmt)
	}

	if p.currentToken.Kind != TokenKindEnd {
		closingKeyword := "end"
		if p.scanner.mode == ScannerModeText {
			closingKeyword = "\\end"
		}
		p.errorAtCurrent(
			fmt.Sprintf("Expected %q to end the block started at line %v.",
				closingKeyword, blockLine))
		return nil
	}
	return block
}

// statement parses a statement. The current token is expected to be the first
// token of the statement.
func (p *parser) statement() ast.Node {
	p.advance()

	// For now, we know only about Text statements.
	switch p.previousToken.Kind {
	case TokenKindText:
		return &ast.Text{
			Text: p.previousToken.Lexeme,
		}
	default:
		p.errorAtPrevious("Expected statement.")
		return nil
	}
}

//
// Parsing helpers (return things other than Nodes)
//

// parseType parses a type. The first token of the type is supposed to be the
// current token.
func (p *parser) parseType() ast.TypeTag {
	result := p.parseTypeNoConsume()
	p.advance()
	return result
}

// parseTypeNoConsume is like parseType, but doesn't consume the last token of
// the type.
func (p *parser) parseTypeNoConsume() ast.TypeTag {
	switch p.currentToken.Kind {
	case TokenKindInt:
		return ast.TypeInt
	case TokenKindFloat:
		return ast.TypeFloat
	case TokenKindBNum:
		return ast.TypeBNum
	case TokenKindString:
		return ast.TypeString
	case TokenKindBool:
		return ast.TypeBool
	case TokenKindVoid:
		return ast.TypeVoid
	default:
		p.errorAtCurrent("Expected type.")
		return ast.TypeInvalid
	}
}

// parseParameterList parses a list of parameters. The left parenthesis is
// supposed to have just been consumed.
func (p *parser) parseParameterList() []ast.Parameter {
	params := []ast.Parameter{}
	if p.check(TokenKindRightParen) {
		p.advance()
		return params
	}

	for {
		p.consume(TokenKindIdentifier, "Expected the parameter name.")
		name := p.previousToken.Lexeme
		p.consume(TokenKindColon, "Expected ':' after parameter name.")
		theType := p.parseType()
		if theType == ast.TypeVoid {
			p.errorAtPrevious("Cannot use 'void' as a parameter type.")
		}
		params = append(params, ast.Parameter{Name: name, Type: theType})

		if !p.match(TokenKindComma) {
			break
		}
	}

	p.consume(
		TokenKindRightParen,
		"Expected ',' to introduce new parameter or ')' to close parameter list.",
	)

	return params
}

//
// Error reporting
//

// TODO: Those error reporting funcs should accept formatting, right?

// errorAtCurrent reports an error at the current (c.currentToken) token.
func (p *parser) errorAtCurrent(message string) {
	p.errorAt(p.currentToken, message)
}

// error reports an error at the token we just consumed (c.previousToken).
func (p *parser) errorAtPrevious(message string) {
	p.errorAt(p.previousToken, message)
}

// errorAt reports an error at a given token, with a given error message.
func (p *parser) errorAt(tok *Token, message string) {
	if p.panicMode {
		return
	}

	p.panicMode = true

	fmt.Fprintf(os.Stderr, "[line %v] Error", tok.Line)

	switch tok.Kind {
	case TokenKindEOF:
		fmt.Fprintf(os.Stderr, " at end")
	case TokenKindError:
		// Nothing.
	default:
		fmt.Fprintf(os.Stderr, " at %q", tok.Lexeme)
	}

	fmt.Fprintf(os.Stderr, ": %v\n", message)
	p.hadError = true
}
