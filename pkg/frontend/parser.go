/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2025 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package frontend

import (
	"fmt"
	"path/filepath"

	"github.com/stackedboxes/romualdo/pkg/ast"
	"github.com/stackedboxes/romualdo/pkg/errs"
	"github.com/stackedboxes/romualdo/pkg/romutil"
)

// precedence is the precedence of expressions.
type precedence int

// TODO: Need to explain, and probably understand, precNone better.
const (
	precNone       precedence = iota // Means: cannot be the "center" of an expression.
	precAssignment                   // =
	precOr                           // or
	precAnd                          // and
	precEquality                     // == !=
	precComparison                   // < > <= >=
	precTerm                         // + -
	precFactor                       // * /
	precBlend                        // ~ // TODO: Not sure about blend precedence or its operator
	PrecUnary                        // not -
	precPower                        // ^
	precCall                         // . ()
	precPrimary
)

// prefixParseFn is a function used to parse code for a certain kind of prefix
// expression. canAssign tells if the expression we are parsing accepts to be
// the target of an assignment.
type prefixParseFn = func(c *parser, canAssign bool) ast.Node

// infixParseFn is a function used to parse code for a certain kind of infix
// expression. lhs is the left-hand side expression previously parsed. canAssign
// tells if the expression we parsing accepts to be the target of an assignment.
type infixParseFn = func(c *parser, lhs ast.Node, canAssign bool) ast.Node

// parseRule encodes one rule of our Pratt parser.
type parseRule struct {
	prefix     prefixParseFn // For expressions using the token as a prefix operator.
	infix      infixParseFn  // For expressions using the token as an infix operator.
	precedence precedence    // When the token is used as a binary operator.
}

// parser is a parser for the Romualdo language. It converts source code into an
// AST.
type parser struct {
	// fileName contains the name of the file being parsed, from the root of the
	// Storyworld.
	fileName string

	// currentToken is the current token we are parsing.
	currentToken *Token

	// previousToken is the previous token we have parsed.
	previousToken *Token

	// errors contains the syntax errors we have found so far.
	errors *errs.CompileTimeCollection

	// panicMode indicates whether we are in panic mode. This has nothing to do
	// with Go panics. Right after finding a syntax error it is hard to generate
	// good error messages because the parser is "out of sync" with the code, so
	// we enter panic mode (during which we don't report any errors). Once we
	// find a "synchronization point", we leave panic mode.
	panicMode bool

	// scanner is the Scanner from where we get our tokens.
	scanner *Scanner
}

// newParser returns a new parser that will parse source. fileName must be
// relative to the Storyworld root; it is used for things like deriving the
// package path, debugging, and error reporting.
func newParser(fileName, source string) *parser {
	return &parser{
		fileName: fileName,
		errors:   &errs.CompileTimeCollection{},
		scanner:  NewScanner(source),
	}
}

// parse parses p.scanner.source and returns the root of the resulting AST.
func (p *parser) parse() (*ast.SourceFile, error) {
	sf := &ast.SourceFile{}

	p.advance()

	for !p.match(TokenKindEOF) {
		decl := p.declaration()
		if p.hadError() {
			return nil, p.errors
		}

		sf.Declarations = append(sf.Declarations, decl)
	}

	return sf, nil
}

// parsePrecedence parses and generates the AST for expressions with a
// precedence level equal to or greater than prec.
func (p *parser) parsePrecedence(prec precedence) ast.Node {
	p.advance()
	prefixRule := rules[p.previousToken.Kind].prefix
	if prefixRule == nil {
		p.errorAtPrevious("Expected expression.")
		return nil
	}

	canAssign := prec <= precAssignment
	node := prefixRule(p, canAssign)

	for prec <= rules[p.currentToken.Kind].precedence {
		p.advance()
		infixRule := rules[p.previousToken.Kind].infix
		node = infixRule(p, node, canAssign)
	}

	// TODO: Not dealing with assignments yet.
	// if canAssign && p.match(TokenKindEqual) {
	// 	p.error("Invalid assignment target.")
	// }

	return node
}

// packagePath returns the package path of the file being parsed.
func (p *parser) packagePath() string {
	result := "/" + filepath.Dir(p.fileName)
	return filepath.Clean(result)
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
func (p *parser) consume(kind TokenKind, format string, a ...any) {
	if p.currentToken.Kind == kind {
		p.advance()
		return
	}

	p.errorAtCurrent(format, a...)
}

//
// Parsing of grammar rules (things that return Nodes)
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
func (p *parser) functionDecl() *ast.ProcedureDecl {
	proc := &ast.ProcedureDecl{
		BaseNode: ast.BaseNode{
			SrcFile:    p.fileName,
			LineNumber: p.previousToken.Line,
		},
		Kind:    ast.ProcKindFunction,
		Package: p.packagePath(),
	}

	p.consume(TokenKindIdentifier, "Expected the function name.")
	proc.Name = p.previousToken.Lexeme

	p.consume(TokenKindLeftParen, "Expected '(' after the function name '%v'.", proc.Name)
	proc.Parameters = p.parseParameterList()
	p.consume(TokenKindColon, "Expected ':' after parameter list.")

	proc.ReturnType = p.parseType()

	proc.Body = p.block()

	return proc
}

// passageDecl parses a passage declaration. The "passage" token must have been
// just consumed.
func (p *parser) passageDecl() *ast.ProcedureDecl {
	proc := &ast.ProcedureDecl{
		Kind:    ast.ProcKindPassage,
		Package: p.packagePath(),
		BaseNode: ast.BaseNode{
			SrcFile:    p.fileName,
			LineNumber: p.previousToken.Line,
		},
	}

	p.consume(TokenKindIdentifier, "Expected the passage name.")
	proc.Name = p.previousToken.Lexeme

	p.consume(TokenKindLeftParen, "Expected '(' after the passage name '%v'.", proc.Name)
	proc.Parameters = p.parseParameterList()
	p.consume(TokenKindColon, "Expected ':' after parameter list.")

	// Consume the last token of the type only after switching to lecture mode,
	// because we want advance() to parse the next token already in lecture
	// mode.
	proc.ReturnType = p.parseTypeNoConsume()
	p.scanner.SetMode(ScannerModeLecture)
	p.scanner.StartNewSpacePrefix()
	p.advance()

	// As above, make sure we switch back to code mode before parsing the first
	// token after the block.
	proc.Body = p.blockNoConsume()
	p.scanner.SetMode(ScannerModeCode)
	p.advance()

	return proc
}

// ifStatement parses an if statement. The if keyword is expected to have just
// been consumed.
func (p *parser) ifStatement() ast.Node {
	n := &ast.IfStmt{
		BaseNode: ast.BaseNode{
			SrcFile:    p.fileName,
			LineNumber: p.previousToken.Line,
		},
	}

	n.Condition = p.expression()
	p.consume(TokenKindThen, "Expected 'then' after condition.")

	thenBlock := &ast.Block{
		BaseNode: ast.BaseNode{
			SrcFile:    p.fileName,
			LineNumber: p.previousToken.Line,
		},
	}

	for !(p.check(TokenKindEnd) || p.check(TokenKindElse) || p.check(TokenKindElseif)) && !p.check(TokenKindEOF) {
		stmt := p.statement()
		thenBlock.Statements = append(thenBlock.Statements, stmt)
	}
	n.Then = thenBlock

	switch {
	case p.match(TokenKindEnd):
		n.Else = nil

	case p.match(TokenKindElse):
		elseBlock := &ast.Block{
			BaseNode: ast.BaseNode{
				SrcFile:    p.fileName,
				LineNumber: p.previousToken.Line,
			},
		}
		for !p.check(TokenKindEnd) && !p.check(TokenKindEOF) {
			stmt := p.statement()
			elseBlock.Statements = append(elseBlock.Statements, stmt)
		}
		p.consume(TokenKindEnd, fmt.Sprintf("Expected: 'end' to close 'if' statement started at line %v.", n.LineNumber))
		n.Else = elseBlock

	case p.match(TokenKindElseif):
		n.Else = p.ifStatement()

	default:
		p.errorAtPrevious(fmt.Sprintf("Unterminated 'if' statement at line %v.", n.LineNumber))
	}
	return n
}

// listen parses a listen expression. The "listen" token is expected to have
// been just consumed.
func (p *parser) listen(canAssign bool) ast.Node {
	options := p.parsePrecedence(precPrimary)

	return &ast.Listen{
		BaseNode: ast.BaseNode{
			SrcFile:    p.fileName,
			LineNumber: p.previousToken.Line,
		},
		Options: options,
	}
}

// expression parses an expression.
func (p *parser) expression() ast.Node {
	return p.parsePrecedence(precAssignment)
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
		if p.scanner.mode == ScannerModeLecture {
			closingKeyword = "\\end"
		}
		p.errorAtCurrent("Expected '%v' to end the block started at line %v.", closingKeyword, blockLine)
		return nil
	}
	return block
}

// statement parses a statement. The current token is expected to be the first
// token of the statement.
func (p *parser) statement() ast.Node {
	switch {
	case p.match(TokenKindLecture):
		// Lectures are handled as statements because they have this dual nature
		// of being both literals and statements ("say" statements, to be
		// precise).
		return &ast.Lecture{
			BaseNode: ast.BaseNode{
				SrcFile:    p.fileName,
				LineNumber: p.previousToken.Line,
			},
			Text: p.previousToken.Lexeme,
		}

	case p.match(TokenKindLeftCurly):
		curlies := &ast.Curlies{
			BaseNode: ast.BaseNode{
				SrcFile:    p.fileName,
				LineNumber: p.previousToken.Line,
			},
		}
		curlies.Expr = p.expression()
		p.consume(TokenKindRightCurly, "Expected `}` to close the curlies started at line %v.", curlies.LineNumber)
		return curlies

	case p.match(TokenKindIf):
		return p.ifStatement()

	case p.check(TokenKindSay):
		// Notice the use of check() instead of match() above to avoid
		// prematurely consuming the next token. That's because a "say" token
		// makes us switch to lecture mode and we want the next token to be
		// handled already in lecture mode.

		// Switch the scanner to lecture mode, because a Lecture is what we
		// expect between a `say`/`end` pair.
		p.scanner.SetMode(ScannerModeLecture)
		p.scanner.StartNewSpacePrefix()

		// Now that we are in lecture mode, we can consume the "say" token.
		p.advance()

		say := &ast.Say{
			BaseNode: ast.BaseNode{
				SrcFile:    p.fileName,
				LineNumber: p.previousToken.Line,
			},
		}

		if p.match(TokenKindLecture) {
			say.Lectures = append(
				say.Lectures,
				&ast.Lecture{
					BaseNode: ast.BaseNode{
						SrcFile:    p.fileName,
						LineNumber: p.previousToken.Line,
					},
					Text: p.previousToken.Lexeme,
				},
			)
		}

		p.consume(TokenKindEnd, "Expected `end` to close the `say` statement started at line %v.", say.LineNumber)

		return say

	default:
		expr := p.expression()
		return &ast.ExpressionStmt{
			BaseNode: ast.BaseNode{
				SrcFile:    p.fileName,
				LineNumber: p.previousToken.Line,
			},
			Expr: expr,
		}
	}
}

// boolLiteral parses a literal Boolean value. The corresponding keyword is
// expected to have been just consumed.
func (p *parser) boolLiteral(canAssign bool) ast.Node {
	if p.previousToken.Kind != TokenKindTrue && p.previousToken.Kind != TokenKindFalse {
		panic(fmt.Sprintf("Unexpected token type on boolLiteral: %v", p.previousToken.Kind))
	}

	return &ast.BoolLiteral{
		BaseNode: ast.BaseNode{
			SrcFile:    p.fileName,
			LineNumber: p.previousToken.Line,
		},
		Value: p.previousToken.Kind == TokenKindTrue,
	}
}

// stringLiteral parses a string literal. The string literal token is expected
// to have been just consumed.
func (p *parser) stringLiteral(canAssign bool) ast.Node {
	// TODO: Assuming strings are always quoted by one single char each side.
	// May change in the future.
	value := p.previousToken.Lexeme[1 : len(p.previousToken.Lexeme)-1] // remove the quotes

	return &ast.StringLiteral{
		BaseNode: ast.BaseNode{
			SrcFile:    p.fileName,
			LineNumber: p.previousToken.Line,
		},
		Value: value,
	}
}

// binary parses a binary operator expression. The left operand and the operator
// token are expected to have been just consumed.
func (p *parser) binary(lhs ast.Node, canAssign bool) ast.Node {
	// Remember the operator.
	operatorKind := p.previousToken.Kind
	operatorLexeme := p.previousToken.Lexeme
	operatorLine := p.previousToken.Line

	// Parse the right operand.
	var rhs ast.Node
	rule := rules[operatorKind]
	if operatorKind == TokenKindHat {
		rhs = p.parsePrecedence(rule.precedence)
	} else {
		rhs = p.parsePrecedence(rule.precedence + 1)
	}

	return &ast.Binary{
		BaseNode: ast.BaseNode{
			SrcFile:    p.fileName,
			LineNumber: operatorLine,
		},
		Operator: operatorLexeme,
		LHS:      lhs,
		RHS:      rhs,
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

// hadError checks if this parser has hit some syntax error already.
func (p *parser) hadError() bool {
	return !p.errors.IsEmpty()
}

// errorAtCurrent reports an error at the current (c.currentToken) token.
func (p *parser) errorAtCurrent(format string, a ...any) {
	p.errorAt(p.currentToken, format, a...)
}

// error reports an error at the token we just consumed (c.previousToken).
func (p *parser) errorAtPrevious(format string, a ...any) {
	p.errorAt(p.previousToken, format, a...)
}

// errorAt reports an error at a given token, with a given error message.
func (p *parser) errorAt(tok *Token, format string, a ...any) {
	if p.panicMode {
		return
	}

	err := &errs.CompileTime{
		Message:  fmt.Sprintf(format, a...),
		FileName: p.fileName,
		Line:     tok.Line,
	}

	p.panicMode = true

	switch tok.Kind {
	case TokenKindEOF:
		err.Lexeme = "end of file"
	case TokenKindError:
		err.Lexeme = ""
	default:
		err.Lexeme = romutil.FormatTextForDisplay(tok.Lexeme)
	}

	p.errors.Add(err)
}

//
// Rules table
//

// rules is the table of parsing rules for our Pratt parser.
var rules []parseRule

func init() {
	initRules()
}

// initRules initializes the rules array.
//
// Using block comments to convince gofmt to keep things aligned is ugly as
// hell.
func initRules() {
	rules = make([]parseRule, TokenKindCount)

	//                                     prefix                                      infix                          precedence
	//                                    ---------------------------------------     --------------------------     --------------
	rules[TokenKindLeftParen] = /*     */ parseRule{nil /*                        */, nil /*                     */, precCall}
	rules[TokenKindRightParen] = /*    */ parseRule{nil /*                        */, nil /*                     */, precNone}
	rules[TokenKindComma] = /*         */ parseRule{nil /*                        */, nil /*                     */, precNone}
	rules[TokenKindColon] = /*         */ parseRule{nil /*                        */, nil /*                     */, precNone}
	rules[TokenKindHat] = /*           */ parseRule{nil /*                        */, nil /*                     */, precNone}

	rules[TokenKindEqual] = /*         */ parseRule{nil /*                        */, nil /*                     */, precNone}
	rules[TokenKindEqualEqual] = /*    */ parseRule{nil /*                        */, (*parser).binary /*        */, precEquality}
	rules[TokenKindBangEqual] = /*     */ parseRule{nil /*                        */, (*parser).binary /*        */, precEquality}
	rules[TokenKindGreater] = /*       */ parseRule{nil /*                        */, nil /*                     */, precNone}
	rules[TokenKindGreater] = /*       */ parseRule{nil /*                        */, nil /*                     */, precNone}
	rules[TokenKindLess] = /*          */ parseRule{nil /*                        */, nil /*                     */, precNone}
	rules[TokenKindLessEqual] = /*     */ parseRule{nil /*                        */, nil /*                     */, precNone}

	rules[TokenKindIdentifier] = /*    */ parseRule{nil /*                        */, nil /*                     */, precNone}
	rules[TokenKindLecture] = /*       */ parseRule{nil /*                        */, nil /*                     */, precNone}
	rules[TokenKindStringLiteral] = /* */ parseRule{(*parser).stringLiteral /*    */, nil /*                     */, precNone}

	rules[TokenKindBNum] = /*          */ parseRule{nil /*                        */, nil /*                     */, precNone}
	rules[TokenKindBool] = /*          */ parseRule{nil /*                        */, nil /*                     */, precNone}
	rules[TokenKindElse] = /*          */ parseRule{nil /*                        */, nil /*                     */, precNone}
	rules[TokenKindElseif] = /*        */ parseRule{nil /*                        */, nil /*                     */, precNone}
	rules[TokenKindEnd] = /*           */ parseRule{nil /*                        */, nil /*                     */, precNone}
	rules[TokenKindFalse] = /*         */ parseRule{(*parser).boolLiteral /*      */, nil /*                     */, precNone}
	rules[TokenKindFloat] = /*         */ parseRule{nil /*                        */, nil /*                     */, precNone}
	rules[TokenKindFunction] = /*      */ parseRule{nil /*                        */, nil /*                     */, precNone}
	rules[TokenKindIf] = /*            */ parseRule{nil /*                        */, nil /*                     */, precNone}
	rules[TokenKindInt] = /*           */ parseRule{nil /*                        */, nil /*                     */, precNone}
	rules[TokenKindListen] = /*        */ parseRule{(*parser).listen /*           */, nil /*                     */, precNone}
	rules[TokenKindPassage] = /*       */ parseRule{nil /*                        */, nil /*                     */, precNone}
	rules[TokenKindSay] = /*           */ parseRule{nil /*                        */, nil /*                     */, precNone}
	rules[TokenKindString] = /*        */ parseRule{nil /*                        */, nil /*                     */, precNone}
	rules[TokenKindThen] = /*          */ parseRule{nil /*                        */, nil /*                     */, precNone}
	rules[TokenKindTrue] = /*          */ parseRule{(*parser).boolLiteral /*      */, nil /*                     */, precNone}
	rules[TokenKindVoid] = /*          */ parseRule{nil /*                        */, nil /*                     */, precNone}

	rules[TokenKindError] = /*         */ parseRule{nil /*                        */, nil /*                     */, precNone}
	rules[TokenKindEOF] = /*           */ parseRule{nil /*                        */, nil /*                     */, precNone}
}
