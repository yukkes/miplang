package lexer

import (
	"fmt"
	"unicode"
)

type Kind string

const (
	EOF          Kind = "EOF"
	Keyword      Kind = "KEYWORD"
	Identifier   Kind = "IDENTIFIER"
	Number       Kind = "NUMBER"
	Semicolon    Kind = ";"
	Colon        Kind = ":"
	Comma        Kind = ","
	LBracket     Kind = "["
	RBracket     Kind = "]"
	LBrace       Kind = "{"
	RBrace       Kind = "}"
	LParen       Kind = "("
	RParen       Kind = ")"
	Plus         Kind = "+"
	Minus        Kind = "-"
	Star         Kind = "*"
	LessEqual    Kind = "<="
	GreaterEqual Kind = ">="
	EqualEqual   Kind = "=="
)

type Token struct {
	Kind    Kind
	Literal string
	Offset  int
	Line    int
	Column  int
}

var keywords = map[string]struct{}{
	"set": {}, "param": {}, "var": {}, "integer": {}, "binary": {},
	"constraint": {}, "minimize": {}, "maximize": {}, "sum": {}, "in": {},
}

func Lex(src string) ([]Token, error) {
	var out []Token
	line, col := 1, 1
	for i := 0; i < len(src); {
		c := src[i]
		if c == ' ' || c == '\t' || c == '\r' {
			i++
			col++
			continue
		}
		if c == '\n' {
			i++
			line++
			col = 1
			continue
		}
		if c == '/' && i+1 < len(src) && src[i+1] == '/' {
			for i < len(src) && src[i] != '\n' {
				i++
				col++
			}
			continue
		}
		start, startLine, startCol := i, line, col
		if isIdentStart(c) {
			i++
			col++
			for i < len(src) && isIdentContinue(src[i]) {
				i++
				col++
			}
			lit := src[start:i]
			kind := Identifier
			if _, ok := keywords[lit]; ok {
				kind = Keyword
			}
			out = append(out, Token{Kind: kind, Literal: lit, Offset: start, Line: startLine, Column: startCol})
			continue
		}
		if isDigit(c) || (c == '.' && i+1 < len(src) && isDigit(src[i+1])) {
			i++
			col++
			for i < len(src) && (isDigit(src[i]) || src[i] == '.') {
				i++
				col++
			}
			out = append(out, Token{Kind: Number, Literal: src[start:i], Offset: start, Line: startLine, Column: startCol})
			continue
		}
		if i+1 < len(src) {
			two := src[i : i+2]
			var kind Kind
			switch two {
			case "<=":
				kind = LessEqual
			case ">=":
				kind = GreaterEqual
			case "==":
				kind = EqualEqual
			}
			if kind != "" {
				out = append(out, Token{Kind: kind, Literal: two, Offset: start, Line: line, Column: col})
				i += 2
				col += 2
				continue
			}
		}
		var kind Kind
		switch c {
		case ';':
			kind = Semicolon
		case ':':
			kind = Colon
		case ',':
			kind = Comma
		case '[':
			kind = LBracket
		case ']':
			kind = RBracket
		case '{':
			kind = LBrace
		case '}':
			kind = RBrace
		case '(':
			kind = LParen
		case ')':
			kind = RParen
		case '+':
			kind = Plus
		case '-':
			kind = Minus
		case '*':
			kind = Star
		default:
			return nil, fmt.Errorf("unexpected character %q at %d:%d", c, line, col)
		}
		out = append(out, Token{Kind: kind, Literal: string(c), Offset: start, Line: line, Column: col})
		i++
		col++
	}
	out = append(out, Token{Kind: EOF, Offset: len(src), Line: line, Column: col})
	return out, nil
}

func isIdentStart(c byte) bool    { return c == '_' || unicode.IsLetter(rune(c)) }
func isIdentContinue(c byte) bool { return isIdentStart(c) || isDigit(c) }
func isDigit(c byte) bool         { return c >= '0' && c <= '9' }
