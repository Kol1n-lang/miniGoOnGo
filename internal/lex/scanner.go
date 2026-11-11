package lex

import "unicode/utf8"

const tabWidth = 4

type scanner struct {
	source       []byte
	position     int
	location     SourceLocation
	ignoreNextLF bool
}

func (s *scanner) peek() (rune, bool) {
	c, size := utf8.DecodeRune(s.source[s.position:])
	if size == 0 {
		return 0, true
	}

	return c, false
}

func (s *scanner) peekN(n int) (rune, bool) {
	if n <= 0 {
		return 0, true
	}

	tail := s.source[s.position:]
	for n > 1 {
		_, size := utf8.DecodeRune(tail)
		if size == 0 {
			return 0, true
		}
		tail = tail[size:]
		n--
	}

	c, size := utf8.DecodeRune(tail)
	if size == 0 {
		return 0, true
	}

	return c, false
}

func (s *scanner) advance() {
	c, size := utf8.DecodeRune(s.source[s.position:])
	if size == 0 {
		return
	}

	s.position += size

	switch c {
	case '\n':
		if s.ignoreNextLF {
			s.ignoreNextLF = false
		} else {
			s.location.Line++
			s.location.Column = 1
		}

	case '\r':
		s.location.Line++
		s.location.Column = 1
		s.ignoreNextLF = true

	case '\t':
		s.location.Column += tabWidth
		s.ignoreNextLF = false

	default:
		s.location.Column++
		s.ignoreNextLF = false
	}
}
