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

	// Operadores de expressão (avaliador de constantes em tempo de
	// montagem -- ver expr.go). Não existiam antes porque nada no KAJI80
	// suportava expressões aritméticas em operandos até esta leva.
	TokenStar   // *
	TokenSlash  // /
	TokenShl    // <<
	TokenShr    // >>
	TokenPipe   // |
	TokenAmp    // &
	TokenCaret  // ^
	TokenTilde  // ~
	TokenOrOr   // ||
	TokenAndAnd // &&
	TokenEqEq   // ==
	TokenNotEq  // !=
	TokenLt     // <
	TokenLtEq   // <=
	TokenGt     // >
	TokenGtEq   // >=
	TokenAssign // = (fora de EQU: "Nome = expressão", variável reatribuível)
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
	case TokenStar:
		return "STAR"
	case TokenSlash:
		return "SLASH"
	case TokenShl:
		return "SHL"
	case TokenShr:
		return "SHR"
	case TokenPipe:
		return "PIPE"
	case TokenAmp:
		return "AMP"
	case TokenCaret:
		return "CARET"
	case TokenTilde:
		return "TILDE"
	case TokenOrOr:
		return "OROR"
	case TokenAndAnd:
		return "ANDAND"
	case TokenEqEq:
		return "EQEQ"
	case TokenNotEq:
		return "NOTEQ"
	case TokenLt:
		return "LT"
	case TokenLtEq:
		return "LTEQ"
	case TokenGt:
		return "GT"
	case TokenGtEq:
		return "GTEQ"
	case TokenAssign:
		return "ASSIGN"
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
