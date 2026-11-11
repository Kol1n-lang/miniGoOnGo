package lex

import (
	"strings"
)

var keywords = map[string]TokenType{
	"func":   Func,
	"main":   Main,
	"int":    Int,
	"float":  Float,
	"string": String,
	"bool":   Bool,
	"true":   True,
	"false":  False,
	"if":     If,
	"else":   Else,
	"while":  While,
	"print":  Print,
	"new":    New,
	"struct": Struct,
	"nil":    Nil,
}

func NewLexer(source []byte) *Lexer {
	return &Lexer{
		scanner: scanner{
			source: source,
			location: SourceLocation{
				Line:   1,
				Column: 1,
			},
		},
	}
}

type Lexer struct {
	scanner       scanner
	startLocation SourceLocation
}

func (l *Lexer) ParseToken() Token {
	if tokenType, stopped := l.skipCommentsAndWhitespace(); stopped {
		return l.token(tokenType)
	}

	l.startLocation = l.scanner.location

	c, _ := l.scanner.peek()

	if isLetterASCII(c) || c == '_' {
		return l.parseIdentifierOrKeyword()
	}

	if isDigitASCII(c) {
		return l.parseNumberLiteral()
	}

	if c == '"' {
		return l.parseStringLiteral()
	}

	switch c {
	case '=':
		l.scanner.advance()
		if next, _ := l.scanner.peek(); next == '=' {
			l.scanner.advance()
			return l.token(Equals)
		}
		return l.token(Assign)

	case '+':
		l.scanner.advance()
		return l.token(Plus)

	case '-':
		l.scanner.advance()
		return l.token(Minus)

	case '*':
		l.scanner.advance()
		return l.token(Multiply)

	case '/':
		l.scanner.advance()
		return l.token(Divide)

	case '%':
		l.scanner.advance()
		return l.token(Modulo)

	case '{':
		l.scanner.advance()
		return l.token(LeftBrace)

	case '}':
		l.scanner.advance()
		return l.token(RightBrace)

	case '(':
		l.scanner.advance()
		return l.token(LeftParen)

	case ')':
		l.scanner.advance()
		return l.token(RightParen)

	case ';':
		l.scanner.advance()
		return l.token(Semicolon)

	case ',':
		l.scanner.advance()
		return l.token(Comma)

	case '[':
		l.scanner.advance()
		return l.token(LeftBracket)

	case ']':
		l.scanner.advance()
		return l.token(RightBracket)

	case '.':
		l.scanner.advance()
		return l.token(Dot)

	case '<':
		l.scanner.advance()
		if next, _ := l.scanner.peek(); next == '=' {
			l.scanner.advance()
			return l.token(LessOrEquals)
		}
		return l.token(Less)

	case '>':
		l.scanner.advance()
		if next, _ := l.scanner.peek(); next == '=' {
			l.scanner.advance()
			return l.token(GreaterOrEquals)
		}
		return l.token(Greater)

	case '!':
		l.scanner.advance()
		if next, _ := l.scanner.peek(); next == '=' {
			l.scanner.advance()
			return l.token(NotEquals)
		}
		return l.token(Not)

	case '&':
		if next, _ := l.scanner.peekN(2); next == '&' {
			l.scanner.advance()
			l.scanner.advance()
			return l.token(And)
		}

	case '|':
		if next, _ := l.scanner.peekN(2); next == '|' {
			l.scanner.advance()
			l.scanner.advance()
			return l.token(Or)
		}
	}

	l.scanner.advance()
	return l.token(Error)
}

func (l *Lexer) parseIdentifierOrKeyword() Token {
	word := make([]byte, 0, 16)

	for {
		c, isEnd := l.scanner.peek()

		if isEnd {
			break
		}

		isIdentifierTail := isLetterASCII(c) ||
			isDigitASCII(c) ||
			c == '_'

		if !isIdentifierTail {
			break
		}

		word = append(word, byte(c))
		l.scanner.advance()
	}

	tokenType, ok := keywords[string(word)]
	if ok {
		return l.token(tokenType)
	}

	return l.tokenWithValue(Identifier, stringValue(string(word)))
}

func (l *Lexer) parseNumberLiteral() Token {
	buffer := make([]byte, 0, 16)

	for {
		c, isEnd := l.scanner.peek()

		if isEnd || !isDigitASCII(c) {
			break
		}

		buffer = append(buffer, byte(c))
		l.scanner.advance()
	}

	if len(buffer) > 1 && buffer[0] == '0' {
		l.skipRemainingNumber()
		return l.token(Error)
	}

	c, isEnd := l.scanner.peek()

	if !isEnd && c == '.' {
		next, nextIsEnd := l.scanner.peekN(2)

		if !nextIsEnd && isDigitASCII(next) {
			buffer = append(buffer, '.')
			l.scanner.advance()

			for {
				c, isEnd = l.scanner.peek()

				if isEnd || !isDigitASCII(c) {
					break
				}

				buffer = append(buffer, byte(c))
				l.scanner.advance()
			}

			value, err := parseFloatValue(buffer)
			if err != nil {
				return l.token(Error)
			}

			return l.tokenWithValue(FloatLiteral, value)
		}
	}

	value, err := parseIntValue(buffer)
	if err != nil {
		return l.token(Error)
	}

	return l.tokenWithValue(IntLiteral, value)
}

func (l *Lexer) skipRemainingNumber() {
	for {
		c, isEnd := l.scanner.peek()
		if isEnd || !isDigitASCII(c) {
			break
		}
		l.scanner.advance()
	}

	if c, isEnd := l.scanner.peek(); !isEnd && c == '.' {
		next, nextIsEnd := l.scanner.peekN(2)
		if !nextIsEnd && isDigitASCII(next) {
			l.scanner.advance() // '.'
			for {
				c, isEnd := l.scanner.peek()
				if isEnd || !isDigitASCII(c) {
					break
				}
				l.scanner.advance()
			}
		}
	}
}

func (l *Lexer) parseStringLiteral() Token {
	l.scanner.advance()

	var builder strings.Builder

	for {
		c, isEnd := l.scanner.peek()

		if isEnd {
			return l.token(Error)
		}

		if c == '"' {
			l.scanner.advance()
			return l.tokenWithValue(StringLiteral, stringValue(builder.String()))
		}

		if c == '\n' || c == '\r' {
			// Незакрытая строка: поглощаем до закрывающей '"' или EOF,
			// чтобы вернуть ровно один Error на весь литерал.
			for {
				c, isEnd := l.scanner.peek()
				if isEnd {
					return l.token(Error)
				}
				l.scanner.advance()
				if c == '"' {
					return l.token(Error)
				}
			}
		}

		if c == '\\' {
			l.scanner.advance()

			next, nextIsEnd := l.scanner.peek()
			if nextIsEnd {
				return l.token(Error)
			}

			switch next {
			case '"':
				builder.WriteByte('"')
			case '\\':
				builder.WriteByte('\\')
			case 'n':
				builder.WriteByte('\n')
			case 't':
				builder.WriteByte('\t')
			case 'r':
				builder.WriteByte('\r')
			default:
				// Неизвестный escape: поглощаем до закрывающей '"' или EOF.
				for {
					c, isEnd := l.scanner.peek()
					if isEnd {
						return l.token(Error)
					}
					l.scanner.advance()
					if c == '"' {
						return l.token(Error)
					}
				}
			}

			l.scanner.advance()
			continue
		}

		builder.WriteRune(c)
		l.scanner.advance()
	}
}

func (l *Lexer) token(tokenType TokenType) Token {
	return Token{
		Type:     tokenType,
		Location: l.startLocation,
	}
}

func (l *Lexer) tokenWithValue(tokenType TokenType, value Value) Token {
	return Token{
		Type:     tokenType,
		Value:    value,
		Location: l.startLocation,
	}
}

func (l *Lexer) skipCommentsAndWhitespace() (tokenType TokenType, stopped bool) {
	for {
		c, isEnd := l.scanner.peek()

		if isEnd {
			l.startLocation = l.scanner.location
			return EndOfFile, true
		}

		if isWhitespace(c) {
			l.scanner.advance()
			continue
		}

		if c == '/' {
			next, _ := l.scanner.peekN(2)

			if next == '/' {
				l.startLocation = l.scanner.location

				l.scanner.advance()
				l.scanner.advance()

				for {
					c, isEnd = l.scanner.peek()

					if isEnd || c == '\n' || c == '\r' {
						break
					}

					l.scanner.advance()
				}

				continue
			}
		}

		break
	}

	return Error, false
}

func isLetterASCII(c rune) bool {
	return c >= 'a' && c <= 'z' ||
		c >= 'A' && c <= 'Z'
}

func isDigitASCII(c rune) bool {
	return c >= '0' && c <= '9'
}

func isWhitespace(c rune) bool {
	return c == ' ' ||
		c == '\t' ||
		c == '\r' ||
		c == '\n'
}
