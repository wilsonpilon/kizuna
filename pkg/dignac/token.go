package dignac

import "fmt"

// TokenType representa o tipo de token da linguagem MSX-BASIC Dignified
type TokenType int

const (
	TokenEOF TokenType = iota
	TokenError
	TokenNewline // Nova linha ou ':' separador de comandos

	// Literais e Identificadores
	TokenIdent
	TokenNumber
	TokenString

	// Palavras-chave de Estrutura e Modularidade
	TokenModule
	TokenEnd
	TokenBank
	TokenPublic
	TokenExtern
	TokenProcedure
	TokenSub
	TokenFunction
	TokenLocal
	TokenDim
	TokenAs
	TokenInteger
	TokenStringKw
	TokenBoolean
	TokenReturn
	TokenExit

	// Controle de Fluxo
	TokenFor
	TokenTo
	TokenStep
	TokenNext
	TokenIf
	TokenThen
	TokenElse
	TokenWhile
	TokenWend
	TokenDo
	TokenLoop

	// Comandos de I/O e Sistema
	TokenPrint
	TokenInput
	TokenCls
	TokenBeep
	TokenScreen
	TokenColor

	// Primitivas Gráficas
	TokenLine
	TokenPset
	TokenB
	TokenBf

	// Operadores Aritméticos e Lógicos
	TokenPlus     // +
	TokenMinus    // -
	TokenMul      // *
	TokenDiv      // / ou \
	TokenMod      // MOD
	TokenAnd      // AND
	TokenOr       // OR
	TokenNot      // NOT
	TokenXor      // XOR
	TokenEqual    // =
	TokenNotEqual // <>
	TokenLess     // <
	TokenLessEq   // <=
	TokenGreater  // >
	TokenGreaterEq// >=

	// Delimitadores
	TokenLParen   // (
	TokenRParen   // )
	TokenComma    // ,
	TokenColon    // :
	TokenSemi     // ;
)

var tokenNames = map[TokenType]string{
	TokenEOF:       "EOF",
	TokenError:     "Error",
	TokenNewline:   "Newline",
	TokenIdent:     "Identifier",
	TokenNumber:    "Number",
	TokenString:    "String",
	TokenModule:    "MODULE",
	TokenEnd:       "END",
	TokenBank:      "BANK",
	TokenPublic:    "PUBLIC",
	TokenExtern:    "EXTERN",
	TokenProcedure: "PROCEDURE",
	TokenSub:       "SUB",
	TokenFunction:  "FUNCTION",
	TokenLocal:     "LOCAL",
	TokenDim:       "DIM",
	TokenAs:        "AS",
	TokenInteger:   "INTEGER",
	TokenStringKw:  "STRING",
	TokenBoolean:   "BOOLEAN",
	TokenReturn:    "RETURN",
	TokenExit:      "EXIT",
	TokenFor:       "FOR",
	TokenTo:        "TO",
	TokenStep:      "STEP",
	TokenNext:      "NEXT",
	TokenIf:        "IF",
	TokenThen:      "THEN",
	TokenElse:      "ELSE",
	TokenWhile:     "WHILE",
	TokenWend:      "WEND",
	TokenDo:        "DO",
	TokenLoop:      "LOOP",
	TokenPrint:     "PRINT",
	TokenInput:     "INPUT",
	TokenCls:       "CLS",
	TokenBeep:      "BEEP",
	TokenScreen:    "SCREEN",
	TokenColor:     "COLOR",
	TokenLine:      "LINE",
	TokenPset:      "PSET",
	TokenB:         "B",
	TokenBf:        "BF",
	TokenPlus:      "+",
	TokenMinus:     "-",
	TokenMul:       "*",
	TokenDiv:       "/",
	TokenMod:       "MOD",
	TokenAnd:       "AND",
	TokenOr:        "OR",
	TokenNot:       "NOT",
	TokenXor:       "XOR",
	TokenEqual:     "=",
	TokenNotEqual:  "<>",
	TokenLess:      "<",
	TokenLessEq:    "<=",
	TokenGreater:   ">",
	TokenGreaterEq: ">=",
	TokenLParen:    "(",
	TokenRParen:    ")",
	TokenComma:     ",",
	TokenColon:     ":",
	TokenSemi:      ";",
}

func (t TokenType) String() string {
	if s, ok := tokenNames[t]; ok {
		return s
	}
	return fmt.Sprintf("Token(%d)", t)
}

// Token representa uma unidade léxica com sua posição no código fonte
type Token struct {
	Type   TokenType
	Value  string
	Line   int
	Column int
}

func (t Token) String() string {
	return fmt.Sprintf("%s(%q) em %d:%d", t.Type, t.Value, t.Line, t.Column)
}
