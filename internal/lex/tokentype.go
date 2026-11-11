package lex

type TokenType int

const (
	// Лексема конца файла.
	EndOfFile = TokenType(iota)

	// Лексема-ошибка.
	Error

	// Литералы.
	Identifier
	IntLiteral
	FloatLiteral
	StringLiteral

	// Ключевые слова miniGoOnGo.
	Func
	Main
	Int
	Float
	String
	Bool
	True
	False
	If
	Else
	While
	Print
	New
	Struct
	Nil

	// Операторы.
	Assign
	Equals
	NotEquals
	Less
	LessOrEquals
	Greater
	GreaterOrEquals
	And
	Or
	Not
	Plus
	Minus
	Multiply
	Divide
	Modulo

	// Прочие лексемы.
	LeftParen
	RightParen
	LeftBrace
	RightBrace
	LeftBracket
	RightBracket
	Comma
	Semicolon
	Dot
)

//go:generate go tool stringer -type=TokenType -output=tokentype_string.go
