/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2023 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package frontend

// TokenKind represents the type of a token. I would call this tokenType if
// "type" wasn't a reserved word in Go. So, there we have it: "TokenKind".
type TokenKind int

const (
	// Single-character tokens
	TokenKindLeftParen  TokenKind = iota // (
	TokenKindRightParen                  // )
	TokenKindComma                       // ,
	TokenKindColon                       // :

	// Literals
	TokenKindIdentifier
	TokenKindText // Text to be sent to the host

	// Keywords
	TokenKindBNum     // bnum
	TokenKindBool     // bool
	TokenKindEnd      // end
	TokenKindFloat    // float
	TokenKindFunction // function
	TokenKindInt      // int
	TokenKindPassage  // passage
	TokenKindString   // string
	TokenKindVoid     // void

	// Special tokens
	TokenKindError
	TokenKindEOF // end-of-file

	// Not really a token
	TokenKindCount
)

// String converts a tokenKind to its string representation. Returns an empty
// string if an invalid kind value is passed.
func (kind TokenKind) String() string {
	switch kind {
	case TokenKindLeftParen:
		return "TokenKindLeftParen"
	case TokenKindRightParen:
		return "TokenKindRightParen"
	case TokenKindComma:
		return "TokenKindComma"
	case TokenKindColon:
		return "TokenKindColon"
	case TokenKindIdentifier:
		return "TokenKindIdentifier"
	case TokenKindText:
		return "TokenKindText"
	case TokenKindBNum:
		return "TokenKindBNum"
	case TokenKindBool:
		return "TokenKindBool"
	case TokenKindEnd:
		return "TokenKindEnd"
	case TokenKindFloat:
		return "TokenKindFloat"
	case TokenKindFunction:
		return "TokenKindFunction"
	case TokenKindInt:
		return "TokenKindInt"
	case TokenKindPassage:
		return "TokenKindPassage"
	case TokenKindString:
		return "TokenKindString"
	case TokenKindVoid:
		return "TokenKindVoid"
	case TokenKindError:
		return "TokenKindError"
	case TokenKindEOF:
		return "TokenKindEOF"
	}
	return ""
}

// A Token is a token -- you know, one of these thingies the scanner generates
// and the compiler consumes.
type Token struct {
	// Kind is the Kind of the token.
	Kind TokenKind

	// Lexeme is the text that makes up the token. It usually is just a slice of
	// the source code string, but there are exceptions. Error tokens, for
	// instance, will use this to store the error message as new string. And
	// some token kinds (Text tokens, in particular) can do quite a bit of
	// pre-processing to the input before turning it into a Lexeme (for example,
	// removing blanks used for indentation).
	Lexeme string

	// Line is the number where the token came from. In case of multiline
	// tokens, it refers to the first Line.
	Line int
}

// IsBackslashed checks if the token is escaped by a backslash.
func (t *Token) IsBackslashed() bool {
	return len(t.Lexeme) > 0 && t.Lexeme[0] == '\\'
}
