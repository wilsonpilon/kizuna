package wirth80

import "fmt"

// TokenType representa a categoria de um token da linguagem Pascal
type TokenType int

const (
	TokenEOF TokenType = iota
	TokenError

	// Literais e Identificadores
	TokenIdent
	TokenNumber
	TokenString

	// Palavras-chave
	TokenProgram
	TokenBegin
	TokenEnd
	TokenVar
	TokenInteger
	TokenChar
	TokenBoolean
	TokenStringKw
	TokenWrite
	TokenWriteLn
	TokenIf
	TokenThen
	TokenElse
	TokenWhile
	TokenDo
	TokenFor
	TokenTo
	TokenDownto

	// Operadores e Delimitadores
	TokenAssign    // :=
	TokenPlus      // +
	TokenMinus     // -
	TokenMul       // *
	TokenDiv       // div ou /
	TokenEqual     // =
	TokenNotEqual  // <>
	TokenLess      // <
	TokenLessEq    // <=
	TokenGreater   // >
	TokenGreaterEq // >=
	TokenLParen    // (
	TokenRParen    // )
	TokenSemi      // ;
	TokenColon     // :
	TokenComma     // ,
	TokenDot       // .
)

var tokenNames = map[TokenType]string{
	TokenEOF:       "EOF",
	TokenError:     "Error",
	TokenIdent:     "Identifier",
	TokenNumber:    "Number",
	TokenString:    "String",
	TokenProgram:   "program",
	TokenBegin:     "begin",
	TokenEnd:       "end",
	TokenVar:       "var",
	TokenInteger:   "Integer",
	TokenChar:      "Char",
	TokenBoolean:   "Boolean",
	TokenStringKw:  "String",
	TokenWrite:     "Write",
	TokenWriteLn:   "WriteLn",
	TokenIf:        "if",
	TokenThen:      "then",
	TokenElse:      "else",
	TokenWhile:     "while",
	TokenDo:        "do",
	TokenFor:       "for",
	TokenTo:        "to",
	TokenDownto:    "downto",
	TokenAssign:    ":=",
	TokenPlus:      "+",
	TokenMinus:     "-",
	TokenMul:       "*",
	TokenDiv:       "div",
	TokenEqual:     "=",
	TokenNotEqual:  "<>",
	TokenLess:      "<",
	TokenLessEq:    "<=",
	TokenGreater:   ">",
	TokenGreaterEq: ">=",
	TokenLParen:    "(",
	TokenRParen:    ")",
	TokenSemi:      ";",
	TokenColon:     ":",
	TokenComma:     ",",
	TokenDot:       ".",
}

func (t TokenType) String() string {
	if name, ok := tokenNames[t]; ok {
		return name
	}
	return fmt.Sprintf("Token(%d)", int(t))
}

// Token representa uma unidade léxica produzida pelo Lexer
type Token struct {
	Type   TokenType
	Value  string
	Number int64
	Line   int
	Column int
}

func (t Token) String() string {
	return fmt.Sprintf("[%s '%s' L%d:C%d]", t.Type, t.Value, t.Line, t.Column)
}
