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
	TokenFloat // literal SINGLE/DOUBLE (ponto/expoente/sufixo != INTEGER puro)
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
	TokenSingle
	TokenDouble
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

	// Sprites
	TokenPut
	TokenSprite
	TokenPattern
	TokenOff

	// Música (PLAY / MML)
	TokenPlay

	// Arquivos
	TokenOpen
	TokenClose
	TokenOutput
	TokenAppend
	TokenHash // '#'

	// Operadores Aritméticos e Lógicos
	TokenPlus      // +
	TokenMinus     // -
	TokenMul       // *
	TokenDiv       // / ou \
	TokenMod       // MOD
	TokenAnd       // AND
	TokenOr        // OR
	TokenNot       // NOT
	TokenXor       // XOR
	TokenEqual     // =
	TokenNotEqual  // <>
	TokenLess      // <
	TokenLessEq    // <=
	TokenGreater   // >
	TokenGreaterEq // >=

	// Delimitadores
	TokenLParen // (
	TokenRParen // )
	TokenComma  // ,
	TokenColon  // :
	TokenSemi   // ;
)

var tokenNames = map[TokenType]string{
	TokenEOF:       "EOF",
	TokenError:     "Error",
	TokenNewline:   "Newline",
	TokenIdent:     "Identifier",
	TokenNumber:    "Number",
	TokenFloat:     "Float",
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
	TokenSingle:    "SINGLE",
	TokenDouble:    "DOUBLE",
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
	TokenPut:       "PUT",
	TokenSprite:    "SPRITE",
	TokenPattern:   "PATTERN",
	TokenOff:       "OFF",
	TokenPlay:      "PLAY",
	TokenOpen:      "OPEN",
	TokenClose:     "CLOSE",
	TokenOutput:    "OUTPUT",
	TokenAppend:    "APPEND",
	TokenHash:      "#",
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
