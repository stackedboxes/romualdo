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

// ScannerMode represents the possible modes the scanner can work in.
type ScannerMode int

const (
	// ScannerModeCode means that the scanner is treating the input as source
	// code, i.e., like a traditional programming language.
	ScannerModeCode = iota

	// ScannerModeText means that the scanner is treating the input as text
	// meant to be sent to the host; programming keywords generally need to be
	// escaped.
	ScannerModeText
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
	// code. For example, a Text segment will have the prefix spaces removed
	// from the lexeme. So, I guess this may not fit the formal definition of a
	// lexeme.
	//
	// TODO: Should probably use a strings.Builder. But also, as noted
	// elsewhere, should be used only if we cannot use a slice of source.
	tokenLexeme string

	// tokenLine contains the line number where the current token started.
	tokenLine int
}

//
// Public API
//

// NewScanner returns a new Scanner that will scan source.
func NewScanner(source string) *Scanner {
	return &Scanner{
		source: source,
		mode:   ScannerModeCode,
		line:   1,
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
		return s.textModeToken()
	case ScannerModeText:
		return s.codeModeToken()
	default:
		panic("Can't happen")
	}
}

// SetMode sets the Scanner's scanning mode.
func (s *Scanner) SetMode(mode ScannerMode) {
	s.mode = mode
}

//
// Code mode
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
	case ':':
		return s.makeToken(TokenKindColon)
	case ',':
		return s.makeToken(TokenKindComma)
	}

	// If we could not figure out what token that rune was supposed start, it's
	// an error.
	return s.errorToken(fmt.Sprintf("Unexpected character %q.", r))
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
		} else if unicode.IsSpace(r) {
			// Other whitespace
			s.current += width
		} else {
			return
		}
	}
}

//
// Text mode
//

// textModeToken returns the next Token, using the "text mode" scanning rules.
func (s *Scanner) textModeToken() *Token {
	s.tokenLine = s.line
	s.skipHorizontalWhitespace()
	s.start = s.current
	if s.isAtEnd() {
		return s.makeToken(TokenKindEOF)
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
	spacePrefix := ""
	if s.peek() == '\n' {
		s.advance()
		s.line += 1
		s.tokenLine += 1
		s.skipHorizontalWhitespace()
	}

	if ok, errToken := s.isSpacePrefixValid(spacePrefix); !ok {
		return errToken
	}

	// Now we are finally at a point where real text could exist.
	for {
		// EOF ends the Text token. But if we already have read some text,
		// return it as a Text token. (The EOF will be returned later on, as the
		// next token is requested.)
		if s.isAtEnd() {
			if len(s.tokenLexeme) > 0 {
				return s.makeToken(TokenKindText)
			}
			return s.makeToken(TokenKindEOF)
		}

		// A backslashed token also ends the Text token. Handling is analogous
		// to the EOF case above.
		if s.atBackslashedToken() {
			if len(s.tokenLexeme) > 0 {
				return s.makeToken(TokenKindText)
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
			if errToken := s.skipSpacePrefix(spacePrefix); errToken != nil {
				return errToken
			}
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
// checks if prefix is valid as the ignored indentation before a Text portion
// starts. It assumes, though, that all runes in prefix were already checked to
// be horizontal whitespace. If the result is false, it additionally returns an
// appropriate error token.
func (s *Scanner) isSpacePrefixValid(prefix string) (bool, *Token) {
	if strings.Contains(prefix, " ") && strings.Contains(prefix, "\t") {
		return false, s.errorToken("Space prefix cannot contain mixed spaces and tabs.")
	}
	return true, nil
}

// skipSpacePrefix skips over the space prefix given by prefix. If the space
// prefix actually found in the input does not match prefix, returns an error
// Token. Otherwise, returns nil.
func (s *Scanner) skipSpacePrefix(prefix string) *Token {
	for _, r := range prefix {
		if s.isAtEnd() || s.advance() != r {
			return s.errorToken("Expected the same space prefix as the previous line.")
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
		return s.errorToken(fmt.Sprintf("Unknown keyword: %q", tok.Lexeme))
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
func (s *Scanner) errorToken(message string) *Token {
	return &Token{
		Kind:   TokenKindError,
		Lexeme: message,
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
	"end":      TokenKindEnd,
	"float":    TokenKindFloat,
	"function": TokenKindFunction,
	"int":      TokenKindInt,
	"passage":  TokenKindPassage,
	"string":   TokenKindString,
	"void":     TokenKindVoid,
}
