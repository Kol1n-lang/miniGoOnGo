package lex

import "fmt"

type Token struct {
	Type     TokenType
	Value    Value
	Location SourceLocation
}

func (t Token) String() string {
	if t.Value.data != nil {
		return fmt.Sprintf("%s(%s, %s)", t.Type.String(), t.Location.String(), t.Value.String())
	}
	return fmt.Sprintf("%s(%s)", t.Type.String(), t.Location.String())
}
