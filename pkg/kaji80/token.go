package kaji80

import "fmt"

// TokenType representa o tipo de token no assembly Z80.
type TokenType int

const (
	TokenEOF TokenType = iota
	TokenNewline
	TokenIdentifier // Instruções, registradores, labels, etc.
	TokenNumber     // Literais numéricos (10, 0x10, 10h, $10, %1010)
	TokenString     // Literais de texto ("msg", 'msg')
	TokenComma      // ,
	TokenColon      // :
	TokenLParen     // (
	TokenRParen     // )
	TokenPlus       // +
	TokenMinus      // -
)

func (t TokenType) String() string {
	switch t {
	case TokenEOF:
		return "EOF"
	case TokenNewline:
		return "NEWLINE"
	case TokenIdentifier:
		return "IDENTIFIER"
	case TokenNumber:
		return "NUMBER"
	case TokenString:
		return "STRING"
	case TokenComma:
		return "COMMA"
	case TokenColon:
		return "COLON"
	case TokenLParen:
		return "LPAREN"
	case TokenRParen:
		return "RPAREN"
	case TokenPlus:
		return "PLUS"
	case TokenMinus:
		return "MINUS"
	default:
		return fmt.Sprintf("TOKEN(%d)", t)
	}
}

// Token representa uma unidade léxica com posição no arquivo fonte.
type Token struct {
	Type   TokenType
	Value  string
	Number int64
	Line   int
	Col    int
}
