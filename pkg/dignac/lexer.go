package dignac

import (
	"fmt"
	"strings"
	"unicode"
)

// Lexer faz a análise léxica do código-fonte MSX-BASIC Dignified
type Lexer struct {
	source []rune
	pos    int
	line   int
	col    int
}

// NewLexer cria uma nova instância de Lexer para o texto fornecido
func NewLexer(input string) *Lexer {
	return &Lexer{
		source: []rune(input),
		pos:    0,
		line:   1,
		col:    1,
	}
}

func (l *Lexer) current() rune {
	if l.pos >= len(l.source) {
		return 0
	}
	return l.source[l.pos]
}

func (l *Lexer) peek() rune {
	if l.pos+1 >= len(l.source) {
		return 0
	}
	return l.source[l.pos+1]
}

func (l *Lexer) advance() rune {
	ch := l.current()
	l.pos++
	if ch == '\n' {
		l.line++
		l.col = 1
	} else {
		l.col++
	}
	return ch
}

func (l *Lexer) skipWhitespaceExceptNewline() {
	for l.pos < len(l.source) {
		ch := l.current()
		if ch == ' ' || ch == '\t' || ch == '\r' {
			l.advance()
		} else {
			break
		}
	}
}

func (l *Lexer) skipToEndOfLine() {
	for l.pos < len(l.source) {
		if l.current() == '\n' {
			break
		}
		l.advance()
	}
}

// NextToken retorna o próximo token léxico do código fonte
func (l *Lexer) NextToken() (Token, error) {
	for {
		l.skipWhitespaceExceptNewline()

		if l.pos >= len(l.source) {
			return Token{Type: TokenEOF, Line: l.line, Column: l.col}, nil
		}

		ch := l.current()
		startLine := l.line
		startCol := l.col

		// Comentário com apóstrofo (')
		if ch == '\'' {
			l.skipToEndOfLine()
			continue
		}

		// Quebra de linha ou ':' funcionam como delimitadores de comando no BASIC
		if ch == '\n' || ch == ':' {
			l.advance()
			// Ignora quebras ou dois-pontos repetidos em sequência
			for l.pos < len(l.source) {
				l.skipWhitespaceExceptNewline()
				c := l.current()
				if c == '\n' || c == ':' {
					l.advance()
				} else if c == '\'' {
					l.skipToEndOfLine()
				} else {
					break
				}
			}
			return Token{Type: TokenNewline, Value: "\n", Line: startLine, Column: startCol}, nil
		}

		// Strings entre aspas
		if ch == '"' {
			return l.lexString()
		}

		// Números hexadecimais (&H... ou $...)
		if ch == '&' {
			next := l.peek()
			if next == 'H' || next == 'h' {
				return l.lexHexNumber("&H")
			}
		}
		if ch == '$' && isHexDigit(l.peek()) {
			return l.lexHexNumber("$")
		}

		// Números decimais
		if unicode.IsDigit(ch) {
			return l.lexNumber()
		}

		// Identificadores e palavras-chave
		if unicode.IsLetter(ch) || ch == '_' {
			tok, err := l.lexIdentOrKeyword()
			if err != nil {
				return Token{}, err
			}
			// Se o identificador for REM, trata o resto da linha como comentário
			if tok.Type == TokenIdent && strings.ToUpper(tok.Value) == "REM" {
				l.skipToEndOfLine()
				continue
			}
			return tok, nil
		}

		// Operadores relacionais e de atribuição
		if ch == '<' {
			l.advance()
			if l.current() == '>' {
				l.advance()
				return Token{Type: TokenNotEqual, Value: "<>", Line: startLine, Column: startCol}, nil
			}
			if l.current() == '=' {
				l.advance()
				return Token{Type: TokenLessEq, Value: "<=", Line: startLine, Column: startCol}, nil
			}
			return Token{Type: TokenLess, Value: "<", Line: startLine, Column: startCol}, nil
		}

		if ch == '>' {
			l.advance()
			if l.current() == '=' {
				l.advance()
				return Token{Type: TokenGreaterEq, Value: ">=", Line: startLine, Column: startCol}, nil
			}
			return Token{Type: TokenGreater, Value: ">", Line: startLine, Column: startCol}, nil
		}

		if ch == '=' {
			l.advance()
			return Token{Type: TokenEqual, Value: "=", Line: startLine, Column: startCol}, nil
		}

		// Operadores de 1 caractere
		l.advance()
		switch ch {
		case '+':
			return Token{Type: TokenPlus, Value: "+", Line: startLine, Column: startCol}, nil
		case '-':
			return Token{Type: TokenMinus, Value: "-", Line: startLine, Column: startCol}, nil
		case '*':
			return Token{Type: TokenMul, Value: "*", Line: startLine, Column: startCol}, nil
		case '/':
			return Token{Type: TokenDiv, Value: "/", Line: startLine, Column: startCol}, nil
		case '\\':
			return Token{Type: TokenDiv, Value: "\\", Line: startLine, Column: startCol}, nil
		case '(':
			return Token{Type: TokenLParen, Value: "(", Line: startLine, Column: startCol}, nil
		case ')':
			return Token{Type: TokenRParen, Value: ")", Line: startLine, Column: startCol}, nil
		case ',':
			return Token{Type: TokenComma, Value: ",", Line: startLine, Column: startCol}, nil
		case ';':
			return Token{Type: TokenSemi, Value: ";", Line: startLine, Column: startCol}, nil
		default:
			return Token{Type: TokenError, Value: string(ch), Line: startLine, Column: startCol}, fmt.Errorf("caractere inesperado '%c' em %d:%d", ch, startLine, startCol)
		}
	}
}

func (l *Lexer) lexString() (Token, error) {
	startLine := l.line
	startCol := l.col
	l.advance() // consome aspas inicial

	var sb strings.Builder
	for {
		if l.pos >= len(l.source) {
			return Token{}, fmt.Errorf("string literal não finalizada em %d:%d", startLine, startCol)
		}
		ch := l.advance()
		if ch == '"' {
			// Verifica se tem aspas duplas escapadas ("")
			if l.current() == '"' {
				sb.WriteRune('"')
				l.advance()
				continue
			}
			break
		}
		if ch == '\n' {
			return Token{}, fmt.Errorf("quebra de linha dentro de string literal em %d:%d", startLine, startCol)
		}
		sb.WriteRune(ch)
	}

	return Token{
		Type:   TokenString,
		Value:  sb.String(),
		Line:   startLine,
		Column: startCol,
	}, nil
}

func (l *Lexer) lexNumber() (Token, error) {
	startLine := l.line
	startCol := l.col
	var sb strings.Builder

	for l.pos < len(l.source) && unicode.IsDigit(l.current()) {
		sb.WriteRune(l.advance())
	}

	return Token{
		Type:   TokenNumber,
		Value:  sb.String(),
		Line:   startLine,
		Column: startCol,
	}, nil
}

func (l *Lexer) lexHexNumber(prefix string) (Token, error) {
	startLine := l.line
	startCol := l.col

	// Consome o prefixo
	for i := 0; i < len(prefix); i++ {
		l.advance()
	}

	var sb strings.Builder
	for l.pos < len(l.source) && isHexDigit(l.current()) {
		sb.WriteRune(l.advance())
	}

	if sb.Len() == 0 {
		return Token{}, fmt.Errorf("número hexadecimal inválido em %d:%d", startLine, startCol)
	}

	return Token{
		Type:   TokenNumber,
		Value:  "&H" + strings.ToUpper(sb.String()),
		Line:   startLine,
		Column: startCol,
	}, nil
}

func isHexDigit(ch rune) bool {
	return unicode.IsDigit(ch) ||
		(ch >= 'a' && ch <= 'f') ||
		(ch >= 'A' && ch <= 'F')
}

func (l *Lexer) lexIdentOrKeyword() (Token, error) {
	startLine := l.line
	startCol := l.col
	var sb strings.Builder

	for l.pos < len(l.source) {
		ch := l.current()
		if unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_' {
			sb.WriteRune(l.advance())
		} else {
			break
		}
	}

	// Permite sufixos de tipo clássicos do BASIC: %, $, !, #
	if l.pos < len(l.source) {
		ch := l.current()
		if ch == '%' || ch == '$' || ch == '!' || ch == '#' {
			sb.WriteRune(l.advance())
		}
	}

	val := sb.String()
	upperVal := strings.ToUpper(val)

	// Verifica se é uma palavra-chave
	kwMap := map[string]TokenType{
		"MODULE":    TokenModule,
		"END":       TokenEnd,
		"BANK":      TokenBank,
		"PUBLIC":    TokenPublic,
		"EXTERN":    TokenExtern,
		"PROCEDURE": TokenProcedure,
		"SUB":       TokenSub,
		"FUNCTION":  TokenFunction,
		"LOCAL":     TokenLocal,
		"DIM":       TokenDim,
		"AS":        TokenAs,
		"INTEGER":   TokenInteger,
		"STRING":    TokenStringKw,
		"BOOLEAN":   TokenBoolean,
		"RETURN":    TokenReturn,
		"EXIT":      TokenExit,
		"FOR":       TokenFor,
		"TO":        TokenTo,
		"STEP":      TokenStep,
		"NEXT":      TokenNext,
		"IF":        TokenIf,
		"THEN":      TokenThen,
		"ELSE":      TokenElse,
		"WHILE":     TokenWhile,
		"WEND":      TokenWend,
		"DO":        TokenDo,
		"LOOP":      TokenLoop,
		"PRINT":     TokenPrint,
		"INPUT":     TokenInput,
		"CLS":       TokenCls,
		"BEEP":      TokenBeep,
		"SCREEN":    TokenScreen,
		"COLOR":     TokenColor,
		"LINE":      TokenLine,
		"PSET":      TokenPset,
		"B":         TokenB,
		"BF":        TokenBf,
		"MOD":       TokenMod,
		"AND":       TokenAnd,
		"OR":        TokenOr,
		"NOT":       TokenNot,
		"XOR":       TokenXor,
	}

	if tokType, ok := kwMap[upperVal]; ok {
		return Token{
			Type:   tokType,
			Value:  val,
			Line:   startLine,
			Column: startCol,
		}, nil
	}

	return Token{
		Type:   TokenIdent,
		Value:  val,
		Line:   startLine,
		Column: startCol,
	}, nil
}
