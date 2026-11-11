package lex

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestScanner(t *testing.T) {
	s := &scanner{
		source: []byte("the" + "\n" + "end"),
		location: SourceLocation{
			Line:   1,
			Column: 1,
		},
	}

	checkPosition(t, s, 1, 1)
	checkPeek(t, s, 't', false)
	checkPeekN(t, s, 1, 't', false)
	checkPeekN(t, s, 2, 'h', false)
	checkPeekN(t, s, 3, 'e', false)

	s.advance()
	checkPosition(t, s, 1, 2)

	checkPeek(t, s, 'h', false)
	checkPeekN(t, s, 1, 'h', false)
	checkPeekN(t, s, 2, 'e', false)
	checkPeekN(t, s, 3, '\n', false)

	s.advance()
	checkPosition(t, s, 1, 3)
	checkPeek(t, s, 'e', false)

	s.advance()
	checkPosition(t, s, 1, 4)
	checkPeek(t, s, '\n', false)

	s.advance()
	checkPosition(t, s, 2, 1)

	checkPeek(t, s, 'e', false)
	checkPeekN(t, s, 1, 'e', false)
	checkPeekN(t, s, 2, 'n', false)
	checkPeekN(t, s, 3, 'd', false)
	checkPeekN(t, s, 4, 0, true)

	s.advance()
	checkPosition(t, s, 2, 2)
	checkPeek(t, s, 'n', false)

	s.advance()
	checkPosition(t, s, 2, 3)
	checkPeek(t, s, 'd', false)

	s.advance()
	checkPosition(t, s, 2, 4)
	checkPeek(t, s, 0, true)

	s.advance()
	checkPosition(t, s, 2, 4)
	checkPeek(t, s, 0, true)
	checkPeekN(t, s, 1, 0, true)
	checkPeekN(t, s, 2, 0, true)
}

func checkPosition(t *testing.T, s *scanner, expectedLine int32, expectedColumn int32) {
	require.Equal(t, expectedLine, s.location.Line)
	require.Equal(t, expectedColumn, s.location.Column)
}

func checkPeek(t *testing.T, s *scanner, expectedRune rune, expectedIsEnd bool) {
	c, isEnd := s.peek()
	require.Equal(t, c, expectedRune)
	require.Equal(t, isEnd, expectedIsEnd)
}

func checkPeekN(t *testing.T, s *scanner, n int, expectedRune rune, expectedIsEnd bool) {
	c, isEnd := s.peekN(n)
	require.Equal(t, c, expectedRune)
	require.Equal(t, isEnd, expectedIsEnd)
}
