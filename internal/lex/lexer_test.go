package lex

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type testCase struct {
	source   string
	expected []Token
}

func (c *testCase) name() string {
	return c.source
}

func tokenize(source string) []Token {
	lexer := NewLexer([]byte(source))
	tokens := make([]Token, 0, 16)
	// Жёсткий лимит: токенов не может быть больше, чем байт в исходнике + 1
	// (плюс EndOfFile). Защита от зацикливания лексера.
	maxTokens := len(source) + 1
	for range maxTokens {
		token := lexer.ParseToken()
		if token.Type == EndOfFile {
			return tokens
		}
		tokens = append(tokens, token)
	}
	panic(fmt.Errorf("too much tokens: source has only %d bytes", len(source)))
}

func TestLexer(t *testing.T) {
	cases := []testCase{
		{
			"name = testUser20",
			[]Token{
				{Identifier, stringValue("name"), SourceLocation{1, 1}},
				{Assign, Value{}, SourceLocation{1, 6}},
				{Identifier, stringValue("testUser20"), SourceLocation{1, 8}},
			},
		},
		{
			// Идентификатор, начинающийся с подчёркивания.
			"_isActive",
			[]Token{
				{Identifier, stringValue("_isActive"), SourceLocation{1, 1}},
			},
		},
		{
			// Идентификатор с подчёркиванием в середине.
			"my_var",
			[]Token{
				{Identifier, stringValue("my_var"), SourceLocation{1, 1}},
			},
		},
		{
			// Регистр важен: Name — не ключевое слово.
			"Name",
			[]Token{
				{Identifier, stringValue("Name"), SourceLocation{1, 1}},
			},
		},
		{
			// Кириллица не является допустимым символом идентификатора
			// по спецификации (только английские буквы).
			// Каждая кириллическая буква — одна руна, значит колонка
			// увеличивается на 1 (scanner считает колонки по рунам).
			"имя",
			[]Token{
				{Error, Value{}, SourceLocation{1, 1}},
				{Error, Value{}, SourceLocation{1, 2}},
				{Error, Value{}, SourceLocation{1, 3}},
			},
		},
		{
			"sum = 10 + 20262027",
			[]Token{
				{Identifier, stringValue("sum"), SourceLocation{1, 1}},
				{Assign, Value{}, SourceLocation{1, 5}},
				{IntLiteral, intValue(10), SourceLocation{1, 7}},
				{Plus, Value{}, SourceLocation{1, 10}},
				{IntLiteral, intValue(20262027), SourceLocation{1, 12}},
			},
		},
		{
			// Вещественные литералы.
			"3.14 0.5 100.0",
			[]Token{
				{FloatLiteral, floatValue(3.14), SourceLocation{1, 1}},
				{FloatLiteral, floatValue(0.5), SourceLocation{1, 6}},
				{FloatLiteral, floatValue(100.0), SourceLocation{1, 10}},
			},
		},
		{
			// Точка без цифры после — не часть числа,
			// но это валидный токен Dot.
			"3.abc",
			[]Token{
				{IntLiteral, intValue(3), SourceLocation{1, 1}},
				{Dot, Value{}, SourceLocation{1, 2}},
				{Identifier, stringValue("abc"), SourceLocation{1, 3}},
			},
		},
		{
			// Ведущий ноль в int — Error.
			"0123",
			[]Token{
				{Error, Value{}, SourceLocation{1, 1}},
			},
		},
		{
			// Ведущий ноль во float — Error.
			"01.5",
			[]Token{
				{Error, Value{}, SourceLocation{1, 1}},
			},
		},
		{
			// Ведущий ноль, точка без цифры после:
			// литерал поглощается до точки, затем '.' — Dot.
			"0123.",
			[]Token{
				{Error, Value{}, SourceLocation{1, 1}},
				{Dot, Value{}, SourceLocation{1, 5}},
			},
		},
		{
			// Литералы int в диапазоне int32.
			"0 1 2147483647 2147483648",
			[]Token{
				{IntLiteral, intValue(0), SourceLocation{1, 1}},
				{IntLiteral, intValue(1), SourceLocation{1, 3}},
				{IntLiteral, intValue(2147483647), SourceLocation{1, 5}},
				{Error, Value{}, SourceLocation{1, 16}},
			},
		},
		{
			// Знак минус — оператор, а не часть литерала.
			"-2147483647 -2147483648",
			[]Token{
				{Minus, Value{}, SourceLocation{1, 1}},
				{IntLiteral, intValue(2147483647), SourceLocation{1, 2}},
				{Minus, Value{}, SourceLocation{1, 13}},
				{Error, Value{}, SourceLocation{1, 14}},
			},
		},
		{
			// Пустая строка.
			`""`,
			[]Token{
				{StringLiteral, stringValue(""), SourceLocation{1, 1}},
			},
		},
		{
			// Строка с escape-последовательностями.
			`"a\nb\tc\\d\"e"`,
			[]Token{
				{StringLiteral, stringValue("a\nb\tc\\d\"e"), SourceLocation{1, 1}},
			},
		},
		{
			// Кириллица в строке (UTF-8).
			`"привет"`,
			[]Token{
				{StringLiteral, stringValue("привет"), SourceLocation{1, 1}},
			},
		},
		{
			// Реальный перенос строки внутри литерала запрещён.
			"\"a\nb\"",
			[]Token{
				{Error, Value{}, SourceLocation{1, 1}},
			},
		},
		{
			// Незакрытая строка — Error.
			`"unclosed`,
			[]Token{
				{Error, Value{}, SourceLocation{1, 1}},
			},
		},
		{
			// Неизвестный escape — Error.
			`"a\qb"`,
			[]Token{
				{Error, Value{}, SourceLocation{1, 1}},
			},
		},
		{
			"true false",
			[]Token{
				{True, Value{}, SourceLocation{1, 1}},
				{False, Value{}, SourceLocation{1, 6}},
			},
		},
		{
			// func main — точка входа miniGoOnGo.
			joinLines(
				"func main() {",
				"}",
			),
			[]Token{
				{Func, Value{}, SourceLocation{1, 1}},
				{Main, Value{}, SourceLocation{1, 6}},
				{LeftParen, Value{}, SourceLocation{1, 10}},
				{RightParen, Value{}, SourceLocation{1, 11}},
				{LeftBrace, Value{}, SourceLocation{1, 13}},
				{RightBrace, Value{}, SourceLocation{2, 1}},
			},
		},
		{
			// if / else / while / print.
			joinLines(
				"if (canRead) {",
				"    print(1);",
				"} else {",
				"    print(2);",
				"}",
			),
			[]Token{
				{If, Value{}, SourceLocation{1, 1}},
				{LeftParen, Value{}, SourceLocation{1, 4}},
				{Identifier, stringValue("canRead"), SourceLocation{1, 5}},
				{RightParen, Value{}, SourceLocation{1, 12}},
				{LeftBrace, Value{}, SourceLocation{1, 14}},
				{Print, Value{}, SourceLocation{2, 5}},
				{LeftParen, Value{}, SourceLocation{2, 10}},
				{IntLiteral, intValue(1), SourceLocation{2, 11}},
				{RightParen, Value{}, SourceLocation{2, 12}},
				{Semicolon, Value{}, SourceLocation{2, 13}},
				{RightBrace, Value{}, SourceLocation{3, 1}},
				{Else, Value{}, SourceLocation{3, 3}},
				{LeftBrace, Value{}, SourceLocation{3, 8}},
				{Print, Value{}, SourceLocation{4, 5}},
				{LeftParen, Value{}, SourceLocation{4, 10}},
				{IntLiteral, intValue(2), SourceLocation{4, 11}},
				{RightParen, Value{}, SourceLocation{4, 12}},
				{Semicolon, Value{}, SourceLocation{4, 13}},
				{RightBrace, Value{}, SourceLocation{5, 1}},
			},
		},
		{
			// while.
			joinLines(
				"while (canRead) {",
				"    print(1);",
				"}",
			),
			[]Token{
				{While, Value{}, SourceLocation{1, 1}},
				{LeftParen, Value{}, SourceLocation{1, 7}},
				{Identifier, stringValue("canRead"), SourceLocation{1, 8}},
				{RightParen, Value{}, SourceLocation{1, 15}},
				{LeftBrace, Value{}, SourceLocation{1, 17}},
				{Print, Value{}, SourceLocation{2, 5}},
				{LeftParen, Value{}, SourceLocation{2, 10}},
				{IntLiteral, intValue(1), SourceLocation{2, 11}},
				{RightParen, Value{}, SourceLocation{2, 12}},
				{Semicolon, Value{}, SourceLocation{2, 13}},
				{RightBrace, Value{}, SourceLocation{3, 1}},
			},
		},
		{
			// Объявления переменных всех четырёх базовых типов.
			joinLines(
				"int x = 10;",
				"float y = 3.14;",
				"string s = \"hello\";",
				"bool b = true;",
			),
			[]Token{
				{Int, Value{}, SourceLocation{1, 1}},
				{Identifier, stringValue("x"), SourceLocation{1, 5}},
				{Assign, Value{}, SourceLocation{1, 7}},
				{IntLiteral, intValue(10), SourceLocation{1, 9}},
				{Semicolon, Value{}, SourceLocation{1, 11}},
				{Float, Value{}, SourceLocation{2, 1}},
				{Identifier, stringValue("y"), SourceLocation{2, 7}},
				{Assign, Value{}, SourceLocation{2, 9}},
				{FloatLiteral, floatValue(3.14), SourceLocation{2, 11}},
				{Semicolon, Value{}, SourceLocation{2, 15}},
				{String, Value{}, SourceLocation{3, 1}},
				{Identifier, stringValue("s"), SourceLocation{3, 8}},
				{Assign, Value{}, SourceLocation{3, 10}},
				{StringLiteral, stringValue("hello"), SourceLocation{3, 12}},
				{Semicolon, Value{}, SourceLocation{3, 19}},
				{Bool, Value{}, SourceLocation{4, 1}},
				{Identifier, stringValue("b"), SourceLocation{4, 6}},
				{Assign, Value{}, SourceLocation{4, 8}},
				{True, Value{}, SourceLocation{4, 10}},
				{Semicolon, Value{}, SourceLocation{4, 14}},
			},
		},
		{
			"== != < <= >= >",
			[]Token{
				{Equals, Value{}, SourceLocation{1, 1}},
				{NotEquals, Value{}, SourceLocation{1, 4}},
				{Less, Value{}, SourceLocation{1, 7}},
				{LessOrEquals, Value{}, SourceLocation{1, 9}},
				{GreaterOrEquals, Value{}, SourceLocation{1, 12}},
				{Greater, Value{}, SourceLocation{1, 15}},
			},
		},
		{
			"+ - / * %",
			[]Token{
				{Plus, Value{}, SourceLocation{1, 1}},
				{Minus, Value{}, SourceLocation{1, 3}},
				{Divide, Value{}, SourceLocation{1, 5}},
				{Multiply, Value{}, SourceLocation{1, 7}},
				{Modulo, Value{}, SourceLocation{1, 9}},
			},
		},
		{
			"&& || !",
			[]Token{
				{And, Value{}, SourceLocation{1, 1}},
				{Or, Value{}, SourceLocation{1, 4}},
				{Not, Value{}, SourceLocation{1, 7}},
			},
		},
		{
			// Одиночные & и | — не операторы miniGoOnGo.
			"& & | |",
			[]Token{
				{Error, Value{}, SourceLocation{1, 1}},
				{Error, Value{}, SourceLocation{1, 3}},
				{Error, Value{}, SourceLocation{1, 5}},
				{Error, Value{}, SourceLocation{1, 7}},
			},
		},
		{
			"{ () }",
			[]Token{
				{LeftBrace, Value{}, SourceLocation{1, 1}},
				{LeftParen, Value{}, SourceLocation{1, 3}},
				{RightParen, Value{}, SourceLocation{1, 4}},
				{RightBrace, Value{}, SourceLocation{1, 6}},
			},
		},
		{
			// Запятая: лексерный тест (токен существует,
			// хотя грамматика print принимает одно выражение).
			"print(1, 2);",
			[]Token{
				{Print, Value{}, SourceLocation{1, 1}},
				{LeftParen, Value{}, SourceLocation{1, 6}},
				{IntLiteral, intValue(1), SourceLocation{1, 7}},
				{Comma, Value{}, SourceLocation{1, 8}},
				{IntLiteral, intValue(2), SourceLocation{1, 10}},
				{RightParen, Value{}, SourceLocation{1, 11}},
				{Semicolon, Value{}, SourceLocation{1, 12}},
			},
		},
		{
			// Комментарий в начале строки пропускается.
			joinLines(
				"// declare variable",
				"int x = 1;",
			),
			[]Token{
				{Int, Value{}, SourceLocation{2, 1}},
				{Identifier, stringValue("x"), SourceLocation{2, 5}},
				{Assign, Value{}, SourceLocation{2, 7}},
				{IntLiteral, intValue(1), SourceLocation{2, 9}},
				// Строка "int x = 1;" — 10 символов, ';' на 10-й колонке.
				{Semicolon, Value{}, SourceLocation{2, 10}},
			},
		},
		{
			// Комментарий после инструкции.
			"int x = 1; // comment",
			[]Token{
				{Int, Value{}, SourceLocation{1, 1}},
				{Identifier, stringValue("x"), SourceLocation{1, 5}},
				{Assign, Value{}, SourceLocation{1, 7}},
				{IntLiteral, intValue(1), SourceLocation{1, 9}},
				{Semicolon, Value{}, SourceLocation{1, 10}},
			},
		},
		{
			// Комментарий в конце файла без перевода строки.
			"int x = 1; // tail",
			[]Token{
				{Int, Value{}, SourceLocation{1, 1}},
				{Identifier, stringValue("x"), SourceLocation{1, 5}},
				{Assign, Value{}, SourceLocation{1, 7}},
				{IntLiteral, intValue(1), SourceLocation{1, 9}},
				{Semicolon, Value{}, SourceLocation{1, 10}},
			},
		},
		{
			// Только комментарий — пустой поток токенов.
			"// only comment",
			[]Token{},
		},
		{
			// Многострочных комментариев нет: "/*" — это Divide Multiply.
			"/* */",
			[]Token{
				{Divide, Value{}, SourceLocation{1, 1}},
				{Multiply, Value{}, SourceLocation{1, 2}},
				{Multiply, Value{}, SourceLocation{1, 4}},
				{Divide, Value{}, SourceLocation{1, 5}},
			},
		},
		{
			// Все три типа переносов эквивалентны: CRLF, CR, LF.
			"first\r\n" +
				"second\r" +
				"third\n" +
				"fourth",
			[]Token{
				{Identifier, stringValue("first"), SourceLocation{1, 1}},
				{Identifier, stringValue("second"), SourceLocation{2, 1}},
				{Identifier, stringValue("third"), SourceLocation{3, 1}},
				{Identifier, stringValue("fourth"), SourceLocation{4, 1}},
			},
		},
		{
			// Табуляция = 4 колонки.
			"\tfirst\n\tsecond",
			[]Token{
				{Identifier, stringValue("first"), SourceLocation{1, 5}},
				{Identifier, stringValue("second"), SourceLocation{2, 5}},
			},
		},
		{
			// Только пробельные символы (SP, HT, LF, CR) — пустой поток.
			" \t\n\r ",
			[]Token{},
		},
		{
			// Пустой исходник — пустой поток.
			"",
			[]Token{},
		},
		{
			// Квадратные скобки в типе массива.
			"int[] arr = new int[5];",
			[]Token{
				{Int, Value{}, SourceLocation{1, 1}},
				{LeftBracket, Value{}, SourceLocation{1, 4}},
				{RightBracket, Value{}, SourceLocation{1, 5}},
				{Identifier, stringValue("arr"), SourceLocation{1, 7}},
				{Assign, Value{}, SourceLocation{1, 11}},
				{New, Value{}, SourceLocation{1, 13}},
				{Int, Value{}, SourceLocation{1, 17}},
				{LeftBracket, Value{}, SourceLocation{1, 20}},
				{IntLiteral, intValue(5), SourceLocation{1, 21}},
				{RightBracket, Value{}, SourceLocation{1, 22}},
				{Semicolon, Value{}, SourceLocation{1, 23}},
			},
		},
		{
			// Создание массива без объявления переменной.
			"new int[10]",
			[]Token{
				{New, Value{}, SourceLocation{1, 1}},
				{Int, Value{}, SourceLocation{1, 5}},
				{LeftBracket, Value{}, SourceLocation{1, 8}},
				{IntLiteral, intValue(10), SourceLocation{1, 9}},
				{RightBracket, Value{}, SourceLocation{1, 11}},
			},
		},
		{
			// Индексация и поле length.
			"arr[0] = 1; arr.length",
			[]Token{
				{Identifier, stringValue("arr"), SourceLocation{1, 1}},
				{LeftBracket, Value{}, SourceLocation{1, 4}},
				{IntLiteral, intValue(0), SourceLocation{1, 5}},
				{RightBracket, Value{}, SourceLocation{1, 6}},
				{Assign, Value{}, SourceLocation{1, 8}},
				{IntLiteral, intValue(1), SourceLocation{1, 10}},
				{Semicolon, Value{}, SourceLocation{1, 11}},
				{Identifier, stringValue("arr"), SourceLocation{1, 13}},
				{Dot, Value{}, SourceLocation{1, 16}},
				{Identifier, stringValue("length"), SourceLocation{1, 17}},
			},
		},
		{
			// Литерал массива.
			"{1, 2, 3}",
			[]Token{
				{LeftBrace, Value{}, SourceLocation{1, 1}},
				{IntLiteral, intValue(1), SourceLocation{1, 2}},
				{Comma, Value{}, SourceLocation{1, 3}},
				{IntLiteral, intValue(2), SourceLocation{1, 5}},
				{Comma, Value{}, SourceLocation{1, 6}},
				{IntLiteral, intValue(3), SourceLocation{1, 8}},
				{RightBrace, Value{}, SourceLocation{1, 9}},
			},
		},
		{
			// Пустой литерал массива.
			"{}",
			[]Token{
				{LeftBrace, Value{}, SourceLocation{1, 1}},
				{RightBrace, Value{}, SourceLocation{1, 2}},
			},
		},
		{
			// Объявление структуры.
			joinLines(
				"struct Point {",
				"    int x;",
				"    int y;",
				"}",
			),
			[]Token{
				{Struct, Value{}, SourceLocation{1, 1}},
				{Identifier, stringValue("Point"), SourceLocation{1, 8}},
				{LeftBrace, Value{}, SourceLocation{1, 14}},
				{Int, Value{}, SourceLocation{2, 5}},
				{Identifier, stringValue("x"), SourceLocation{2, 9}},
				{Semicolon, Value{}, SourceLocation{2, 10}},
				{Int, Value{}, SourceLocation{3, 5}},
				{Identifier, stringValue("y"), SourceLocation{3, 9}},
				{Semicolon, Value{}, SourceLocation{3, 10}},
				{RightBrace, Value{}, SourceLocation{4, 1}},
			},
		},
		{
			// Литерал структуры и доступ к полю.
			"Point p = Point{1, 2}; print(p.x);",
			[]Token{
				{Identifier, stringValue("Point"), SourceLocation{1, 1}},
				{Identifier, stringValue("p"), SourceLocation{1, 7}},
				{Assign, Value{}, SourceLocation{1, 9}},
				{Identifier, stringValue("Point"), SourceLocation{1, 11}},
				{LeftBrace, Value{}, SourceLocation{1, 16}},
				{IntLiteral, intValue(1), SourceLocation{1, 17}},
				{Comma, Value{}, SourceLocation{1, 18}},
				{IntLiteral, intValue(2), SourceLocation{1, 20}},
				{RightBrace, Value{}, SourceLocation{1, 21}},
				{Semicolon, Value{}, SourceLocation{1, 22}},
				{Print, Value{}, SourceLocation{1, 24}},
				{LeftParen, Value{}, SourceLocation{1, 29}},
				{Identifier, stringValue("p"), SourceLocation{1, 30}},
				{Dot, Value{}, SourceLocation{1, 31}},
				{Identifier, stringValue("x"), SourceLocation{1, 32}},
				{RightParen, Value{}, SourceLocation{1, 33}},
				{Semicolon, Value{}, SourceLocation{1, 34}},
			},
		},
		{
			// Присваивание nil структуре.
			"Point p = nil;",
			[]Token{
				{Identifier, stringValue("Point"), SourceLocation{1, 1}},
				{Identifier, stringValue("p"), SourceLocation{1, 7}},
				{Assign, Value{}, SourceLocation{1, 9}},
				{Nil, Value{}, SourceLocation{1, 11}},
				{Semicolon, Value{}, SourceLocation{1, 14}},
			},
		},
		{
			// Сравнение с nil.
			"if (p == nil) { }",
			[]Token{
				{If, Value{}, SourceLocation{1, 1}},
				{LeftParen, Value{}, SourceLocation{1, 4}},
				{Identifier, stringValue("p"), SourceLocation{1, 5}},
				{Equals, Value{}, SourceLocation{1, 7}},
				{Nil, Value{}, SourceLocation{1, 10}},
				{RightParen, Value{}, SourceLocation{1, 13}},
				{LeftBrace, Value{}, SourceLocation{1, 15}},
				{RightBrace, Value{}, SourceLocation{1, 17}},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name(), func(t *testing.T) {
			actual := tokenize(c.source)
			assertEqualTokens(t, c.expected, actual)
		})
	}
}

func TestTokenStringMethod(t *testing.T) {
	code := "print(x)"
	tokenStrings := toStringSlice(tokenize(code))
	require.Equal(
		t,
		[]string{
			"Print(1:1)",
			"LeftParen(1:6)",
			"Identifier(1:7, x)",
			"RightParen(1:8)",
		},
		tokenStrings,
	)
}

// assertEqualTokens сравнивает две последовательности токенов,
// а если они не равны, то останавливает тест паникой и печатает разницу.
func assertEqualTokens(t *testing.T, expected, actual []Token) {
	if !slices.Equal(expected, actual) {
		expectedPrinted := toStringSlice(expected)
		actualPrinted := toStringSlice(actual)
		require.Equal(t, expectedPrinted, actualPrinted)
	}
}

func joinLines(lines ...string) string {
	return strings.Join(lines, "\n")
}

func toStringSlice[T fmt.Stringer](s []T) []string {
	result := make([]string, len(s))
	for i, v := range s {
		result[i] = v.String()
	}
	return result
}
