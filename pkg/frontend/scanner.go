/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2023 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package frontend

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// TODO: We should probably make the types here non-exported and have one
// exported function that gets called from scanCmd. Similar to what we have with
// the parser.

// ScannerMode represents the possible modes the scanner can work in.
type ScannerMode int

const (
	// ScannerModeCode means that the scanner is treating the input as source
	// code, i.e., like a traditional programming language.
	ScannerModeCode ScannerMode = iota

	// ScannerModeLecture means that the scanner is treating the input as text
	// meant to be `say`d; programming keywords generally need to be escaped.
	ScannerModeLecture
)

// A Scanner (AKA lexer) tokenizes Romualdo source code.
type Scanner struct {
	// source is the source code being scanned.
	source string

	// mode is the mode of operation the scanner is working in.
	mode ScannerMode

	// start points to the start of the token being currently scanned. It points
	// into source.
	start int

	// current points to the code unit we are currently looking at. It points
	// into source.
	current int

	// line holds the line number we are currently looking at.
	line int

	// tokenLexeme contains the current lexeme, as we scan it. For backslashed
	// tokens, this includes the backslash rune.
	//
	// This lexeme may be a transformed version of what is actually found in the
	// code. For example, a Lecture will have the prefix spaces removed from the
	// lexeme. So, I guess this may not fit the formal definition of a lexeme.
	//
	// TODO: Should probably use a strings.Builder. But also, as noted
	// elsewhere, should be used only if we cannot use a slice of source.
	tokenLexeme string

	// tokenLine contains the line number where the current token started.
	tokenLine int

	// spacePrefix is a stack of "space prefixes". The one at the top is used in
	// the current Lecture. A "space prefix" is the indentation common to every
	// line of a Lecture, which is discarded by the scanner.
	spacePrefix []string

	// startNewSpacePrefix is set to true to tell the scanner that we are at a
	// point in which we shall start a new space prefix.
	startNewSpacePrefix bool
}

//
// Public API
//

// NewScanner returns a new Scanner that will scan source.
func NewScanner(source string) *Scanner {
	return &Scanner{
		source:      source,
		mode:        ScannerModeCode,
		line:        1,
		spacePrefix: []string{""},
	}
}

// Token returns the next Token in the source code being scanned.
func (s *Scanner) Token() *Token {
	s.tokenLexeme = ""

	if s.isAtEnd() {
		return s.makeToken(TokenKindEOF)
	}

	switch s.mode {
	case ScannerModeCode:
		return s.codeModeToken()
	case ScannerModeLecture:
		ssp := s.startNewSpacePrefix
		s.startNewSpacePrefix = false
		return s.lectureModeToken(ssp)
	default:
		panic("Can't happen")
	}
}

// SetMode sets the Scanner's scanning mode.
func (s *Scanner) SetMode(mode ScannerMode) {
	s.mode = mode
}

// StartNewSpacePrefix tells the scanner that we are at a point in which we
// shall start a new space prefix.
func (s *Scanner) StartNewSpacePrefix() {
	s.startNewSpacePrefix = true
}

//
// Code Mode
//

// codeModeToken returns the next token, using the "code mode" scanning rules.
func (s *Scanner) codeModeToken() *Token {
	// Skip whitespace, handle EOF.
	s.skipWhitespace()
	s.start = s.current
	s.tokenLine = s.line
	if s.isAtEnd() {
		return s.makeToken(TokenKindEOF)
	}

	// Check for a backslashed token, which is also valid in code mode.
	if s.atBackslashedToken() {
		return s.backslashedToken()
	}

	// Handle normal (non-backslashed) code.
	r := s.advance()
	s.tokenLexeme += string(r)

	if unicode.IsLetter(r) {
		return s.scanIdentifier()
	}

	switch r {
	case '(':
		return s.makeToken(TokenKindLeftParen)
	case ')':
		return s.makeToken(TokenKindRightParen)
	case '[':
		return s.makeToken(TokenKindLeftSquare)
	case ']':
		return s.makeToken(TokenKindRightSquare)
	case '}':
		// This puts us back into Lecture mode.
		s.SetMode(ScannerModeLecture)
		return s.makeToken(TokenKindRightCurly)
	case ':':
		return s.makeToken(TokenKindColon)
	case ',':
		return s.makeToken(TokenKindComma)
	case '^':
		return s.makeToken(TokenKindHat)
	case '!':
		if s.match('=') {
			s.tokenLexeme += "="
			return s.makeToken(TokenKindBangEqual)
		}
		return s.errorToken("'!' must be followed by '='.")
	case '=':
		if s.match('=') {
			s.tokenLexeme += "="
			return s.makeToken(TokenKindEqualEqual)
		}
		return s.makeToken(TokenKindEqual)
	case '<':
		if s.match('=') {
			s.tokenLexeme += "="
			return s.makeToken(TokenKindLessEqual)
		}
		return s.makeToken(TokenKindLess)
	case '>':
		if s.match('=') {
			s.tokenLexeme += "="
			return s.makeToken(TokenKindGreaterEqual)
		}
		return s.makeToken(TokenKindGreater)
	case '"':
		// TODO: For now, strings are always double-quoted. We should probably
		// support single-quoted and back-quoted strings as well. Maybe with
		// different semantics, maybe just for giving choice of what must be
		// escaped. Open design point.
		return s.scanString()
	}

	// If we could not figure out what token that rune was supposed start, it's
	// an error.
	return s.errorToken("unexpected character '%c'.", r)
}

// skipWhitespace skips all whitespace and comments, leaving the s.current index
// pointing to the start of a non-space, non-comment rune.
func (s *Scanner) skipWhitespace() {
	for {
		if s.isAtEnd() {
			return
		}

		r, width := utf8.DecodeRuneInString(s.source[s.current:])

		if r == '\\' && s.peekNext() == '#' {
			// Comments
			s.skipComment()
		} else if r == '\n' {
			// Line breaks
			s.line++
			s.current += width
		} else if r == '\r' {
			// Ignore carriage returns
			s.current += width
		} else if unicode.IsSpace(r) {
			// Other whitespace
			s.current += width
		} else {
			return
		}
	}
}

//
// Lecture Mode
//

// lectureModeToken returns the next Token, using the "lecture mode" scanning
// rules. newSpacePrefix tells if we are at a point in which we shall start a
// new space prefix.
func (s *Scanner) lectureModeToken(newSpacePrefix bool) *Token {
	s.tokenLine = s.line
	if newSpacePrefix {
		s.skipHorizontalWhitespace()
	}
	s.start = s.current
	if s.isAtEnd() {
		return s.makeToken(TokenKindEOF)
	}

	if s.peek() == '\r' {
		// Ignore carriage returns
		s.advance()
	}

	// Check for a backslashed token.
	if s.atBackslashedToken() {
		return s.backslashedToken()
	}

	// If starting with a line break we need some special handling. First, we
	// ignore this line break (don't add it to the lexeme). Then, also ignore
	// any horizontal whitespace. But remember this amount of horizontal
	// whitespace, because it will be ignored from every subsequent line. This
	// is to allow nice indentation of source code without adding a lot of
	// spaces to the lexemes.
	if s.peek() == '\n' {
		s.advance()
		s.line += 1
		s.tokenLine += 1

		// If we are in a point where we shall start a new space prefix, we
		// obtain the new space prefix with `s.skipHorizontalWhitespace()` and
		// push it onto the stack of space prefixes.
		if newSpacePrefix {
			s.spacePrefixPush(s.skipHorizontalWhitespace())
		}
	}

	sp := ""
	if len(s.spacePrefix) > 0 {
		sp = s.spacePrefixTop()
	}
	if ok, errToken := s.isSpacePrefixValid(sp); !ok {
		return errToken
	}

	// Now we are finally at a point where real text could exist.
	for {
		// EOF ends the Lecture token. But if we already have read some text,
		// return it as a Lecture token. (The EOF will be returned as the next
		// token.)
		if s.isAtEnd() {
			if len(s.tokenLexeme) > 0 {
				return s.makeToken(TokenKindLecture)
			}
			return s.makeToken(TokenKindEOF)
		}

		// A backslashed token also ends the Lecture token. Handling is
		// analogous to the EOF case above.
		if s.atBackslashedToken() {
			if len(s.tokenLexeme) > 0 {
				return s.makeToken(TokenKindLecture)
			}
			return s.backslashedToken()
		}

		r := s.advance()
		switch r {
		case '\\':
			next := s.peek()
			if next == '\\' {
				// Skip one of the backslashes in the pair.
				s.advance()
				s.tokenLexeme += string('\\')
			} else if next == '#' {
				s.skipComment()
			}
		case '\n':
			s.line += 1
			s.tokenLexeme += string('\n')
			if s.isAtEnd() || s.atBackslashedToken() {
				continue
			}
			if errToken := s.skipSpacePrefix(s.spacePrefixTop()); errToken != nil {
				// We failed to match the space prefix. This is not necessarily
				// an error: if the token right ahead is an `end` token, we
				// shall use instead of erroring out.
				if s.current < 1 || s.current+2 > len(s.source) {
					return errToken
				}
				if s.source[s.current-1:s.current+2] == "end" {
					// We have an `end` token ahead. Let's return the Lecture we
					// just read, and set everything up so that the `end` token
					// is returned next.
					s.tokenLexeme = s.tokenLexeme[0:len(s.tokenLexeme)]
					tok := s.makeToken(TokenKindLecture)
					s.current -= 1 // the `e` of `end` was consumed; undo that
					s.SetMode(ScannerModeCode)
					s.spacePrefixPop()
					return tok
				}
				return errToken
			}
		case '\r':
			// Ignore carriage returns

		case '{':
			s.tokenLexeme += "{"

			if s.tokenLexeme != "{" {
				// We have a `{` token ahead, but already scanned a Lecture.
				// Let's return this Lecture and set everything up so that the
				// `{` token is returned next.
				s.tokenLexeme = s.tokenLexeme[0 : len(s.tokenLexeme)-1] // Ignore the `{`.
				s.current -= 1                                          // The `{` was consumed; undo that.
				tok := s.makeToken(TokenKindLecture)
				return tok
			}

			// And if we got here, the `{` token is the one to return. Starting
			// from here, we want to be in code mode.
			s.SetMode(ScannerModeCode)
			return s.makeToken(TokenKindLeftCurly)

		default:
			s.tokenLexeme += string(r)
		}
	}
}

// skipHorizontalWhitespace skips horizontal whitespace, returns a string with
// whatever was skipped.
func (s *Scanner) skipHorizontalWhitespace() string {
	start := s.current
	for {
		if s.isAtEnd() {
			return s.source[start:s.current]
		}

		r := s.peek()
		if r != ' ' && r != '\t' {
			return s.source[start:s.current]
		}
		s.advance()
	}
}

// isSpacePrefixValid checks if prefix is a valid space prefix. In other words,
// checks if prefix is valid as the ignored indentation before the relevant
// contents of a Lecture line starts. It assumes, though, that all runes in
// prefix were already checked to be horizontal whitespace. If the result is
// false, it additionally returns an appropriate error token.
func (s *Scanner) isSpacePrefixValid(prefix string) (bool, *Token) {
	if strings.Contains(prefix, " ") && strings.Contains(prefix, "\t") {
		return false, s.errorToken("space prefix cannot contain mixed spaces and tabs.")
	}
	return true, nil
}

// skipSpacePrefix skips over the space prefix given by prefix. If the space
// prefix actually found in the input does not match prefix, returns an error
// Token. Otherwise, returns nil.
func (s *Scanner) skipSpacePrefix(prefix string) *Token {
	for _, r := range prefix {
		if s.isAtEnd() || s.advance() != r {
			return s.errorToken("expected the same space prefix as the previous line.")
		}
	}
	return nil
}

//
// Mode-independent
//

// skipComment skips until the end of the current comment.
func (s *Scanner) skipComment() {
	for s.peek() != '\n' && !s.isAtEnd() {
		_, width := utf8.DecodeRuneInString(s.source[s.current:])
		s.current += width
	}
}

// match checks if the next rune matches the expected one. If it does, the
// scanner consumes the rune and returns true. If not, the scanner leaves the
// rune there (not consuming it) and returns false.
func (s *Scanner) match(expected rune) bool {
	if s.isAtEnd() {
		return false
	}

	currentRune, width := utf8.DecodeRuneInString(s.source[s.current:])

	if currentRune != expected {
		return false
	}

	s.current += width
	return true
}

// atBackslashedToken checks if we are at the start of a backslashed token.
//
// This doesn't check if the token is valid backslashed token: it could be
// something invalid as well, like `\not_an_actual_keyword`. In other words, if
// this returns true, we either have a backslashed token or an error right ahead
// of us.
func (s *Scanner) atBackslashedToken() bool {
	if s.peek() == '\\' {
		next := s.peekNext()
		// Ignore end-of-file (0), escaped backslashes, and comments.
		return next != 0 && next != '\\' && next != '#'
	}
	return false
}

// backslashedToken scans and returns the next Token, which is assumed to be a
// backslashed token (as tested by atBackslashedToken()).
func (s *Scanner) backslashedToken() *Token {
	// Consume the backslash.
	r := s.advance()
	s.tokenLexeme += string(r)

	// Look for the keyword itself.
	tok := s.scanIdentifier()
	if tok.Kind == TokenKindIdentifier {
		// There's no such thing as a "backslashed identifier", only backslashed
		// keywords. Ergo, this must be an error.
		return s.errorToken("Unknown keyword: '%v'", tok.Lexeme[1:])
	}
	return tok
}

// scanIdentifier scans and returns a Token that is either an identifier or a
// keyword. This function doesn't do any special handling for backslashed
// keywords, but if the backslash has already been consumed, it will work just
// fine.
func (s *Scanner) scanIdentifier() *Token {
	for {
		r := s.peek()
		if r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r) {
			next := s.advance()
			s.tokenLexeme += string(next)
		} else {
			break
		}
	}
	return s.makeToken(s.identifierKind())
}

// scanString scans a string token.
func (s *Scanner) scanString() *Token {
	for s.peek() != '"' && !s.isAtEnd() {
		if s.peek() == '\n' {
			s.line++
		}
		r := s.advance()
		s.tokenLexeme += string(r)
	}

	if s.isAtEnd() {
		return s.errorToken("Unterminated string.")
	}

	// The closing quote.
	r := s.advance()
	s.tokenLexeme += string(r)
	return s.makeToken(TokenKindStringLiteral)
}

//
// Space prefix stack
//

// spacePrefixPush pushes a new space prefix onto the stack of space prefixes.
func (s *Scanner) spacePrefixPush(prefix string) {
	s.spacePrefix = append(s.spacePrefix, prefix)
}

// spacePrefixPop pops the space prefix on the top of the stack of space
// prefixes, and returns it for convenience. It soes not check for overflow.
func (s *Scanner) spacePrefixPop() string {
	i := len(s.spacePrefix) - 1
	result := s.spacePrefix[i]
	s.spacePrefix = s.spacePrefix[:i]
	return result
}

// spacePrefixTop pops the space prefix on the top of the stack of space
// prefixes, and returns it for convenience. It soes not check for overflow.
func (s *Scanner) spacePrefixTop() string {
	i := len(s.spacePrefix) - 1
	return s.spacePrefix[i]
}

//
// Helpers
//

// advance returns the next rune in the input source and advance the s.current
// index so that it points to the start of the next rune.
func (s *Scanner) advance() rune {
	r, width := utf8.DecodeRuneInString(s.source[s.current:])
	s.current += width
	return r
}

// isAtEnd checks if the scanner reached the end of the input. Specifically,
// this means that s.current is pointing just beyond the last valid index of
// s.source.
func (s *Scanner) isAtEnd() bool {
	return s.current == len(s.source)
}

// peek returns the current rune from the input without advancing the s.current
// pointer.
func (s *Scanner) peek() rune {
	r, _ := utf8.DecodeRuneInString(s.source[s.current:])
	return r
}

// peekNext returns the next rune from the input (one rune past s.current)
// without the advancing s.current pointer. Returns 0 if this would be beyond
// the end of the input.
func (s *Scanner) peekNext() rune {
	if s.current >= len(s.source)-1 {
		return 0
	}
	_, width := utf8.DecodeRuneInString(s.source[s.current:])
	r, _ := utf8.DecodeRuneInString(s.source[s.current+width:])
	return r
}

// makeToken returns a token of a given kind.
func (s *Scanner) makeToken(kind TokenKind) *Token {
	// TODO: Is there a better way to deal with the lexeme? I'd like to keep
	// using a slice of s.source when possible, but also to avoid creating the
	// s.tokenLexeme stuff as much as possible. Can I hide all details behind
	// methods that create and use s.tokenLexeme only when needed? Things to try
	// when we have much more things working (otherwise we can't have any
	// meaningful CPU or memory usage measurements).
	lexeme := s.source[s.start:s.current]
	if s.tokenLexeme != lexeme {
		lexeme = s.tokenLexeme
	}
	return &Token{
		Kind:   kind,
		Lexeme: lexeme,
		Line:   s.tokenLine,
	}
}

// errorToken returns a new token of kind TokenKindError containing a given
// error message.
func (s *Scanner) errorToken(format string, a ...any) *Token {
	return &Token{
		Kind:   TokenKindError,
		Lexeme: fmt.Sprintf(format, a...),
		Line:   s.line,
	}
}

// identifierKind returns the token kind corresponding to the current token
// (assumed to be either a keyword or an identifier).
func (s *Scanner) identifierKind() TokenKind {
	lexeme := s.tokenLexeme
	if len(s.tokenLexeme) > 0 && s.tokenLexeme[0] == '\\' {
		// This is a backslashed token; ignore the backslash for the purposes of
		// keyword matching.
		lexeme = lexeme[1:]
	}

	// TODO: Someday it would be interesting to measure how this compares
	// performance-wise with a handcrafted matching function.
	kind, ok := lexemeToTokenKind[lexeme]
	if ok {
		return kind
	}
	return TokenKindIdentifier
}

// lexemeToTokenKind maps the keyword lexeme to its corresponding token kind.
var lexemeToTokenKind = map[string]TokenKind{
	"bnum":     TokenKindBNum,
	"bool":     TokenKindBool,
	"else":     TokenKindElse,
	"elseif":   TokenKindElseif,
	"end":      TokenKindEnd,
	"false":    TokenKindFalse,
	"float":    TokenKindFloat,
	"function": TokenKindFunction,
	"if":       TokenKindIf,
	"int":      TokenKindInt,
	"listen":   TokenKindListen,
	"passage":  TokenKindPassage,
	"say":      TokenKindSay,
	"string":   TokenKindString,
	"then":     TokenKindThen,
	"true":     TokenKindTrue,
	"void":     TokenKindVoid,
}
