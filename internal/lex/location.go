package lex

import "fmt"

type SourceLocation struct {
	Line   int32
	Column int32
}

func (l SourceLocation) String() string {
	return fmt.Sprintf("%d:%d", l.Line, l.Column)
}
