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

		// Números hexadecimais (&H... ou $...), octais (&O...) e binários (&B...)
		if ch == '&' {
			next := l.peek()
			switch next {
			case 'H', 'h':
				return l.lexHexNumber("&H")
			case 'O', 'o':
				return l.lexRadixNumber("&O", isOctalDigit)
			case 'B', 'b':
				return l.lexRadixNumber("&B", isBinaryDigit)
			}
		}
		if ch == '$' && isHexDigit(l.peek()) {
			return l.lexHexNumber("$")
		}

		// '#' -- número de arquivo (OPEN ... AS #1 / PRINT #1, / CLOSE #1)
		if ch == '#' {
			l.advance()
			return Token{Type: TokenHash, Value: "#", Line: startLine, Column: startCol}, nil
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

// lexNumber lê um número decimal, que pode ser um INTEGER simples ("42") ou
// um literal SINGLE/DOUBLE de ponto flutuante: parte fracionária ("3.14"),
// notação de expoente com "e"/"E" (implica SINGLE) ou "d"/"D" (implica
// DOUBLE), com sinal opcional ("1.5e+10", "3.14159d-5"), e/ou um sufixo de
// tipo explícito "!" (SINGLE) ou "#" (DOUBLE) no final ("42!", "42#"). Sem
// nenhum desses marcadores, o literal é um INTEGER (TokenNumber); com
// qualquer um deles, vira um TokenFloat cujo Value é o número no formato
// aceito por strconv.ParseFloat mais um marcador de precisão ('S'/'D') no
// último caractere, que parsePrimary (parser.go) lê e remove.
func (l *Lexer) lexNumber() (Token, error) {
	startLine := l.line
	startCol := l.col
	var sb strings.Builder

	for l.pos < len(l.source) && unicode.IsDigit(l.current()) {
		sb.WriteRune(l.advance())
	}

	isFloat := false
	isDouble := false

	if l.current() == '.' && unicode.IsDigit(l.peek()) {
		isFloat = true
		sb.WriteRune(l.advance())
		for l.pos < len(l.source) && unicode.IsDigit(l.current()) {
			sb.WriteRune(l.advance())
		}
	}

	if c := l.current(); c == 'e' || c == 'E' || c == 'd' || c == 'D' {
		savedPos, savedLine, savedCol := l.pos, l.line, l.col
		expMarker := c
		l.advance()
		sign := ""
		if l.current() == '+' || l.current() == '-' {
			sign = string(l.current())
			l.advance()
		}
		if unicode.IsDigit(l.current()) {
			isFloat = true
			if expMarker == 'd' || expMarker == 'D' {
				isDouble = true
			}
			sb.WriteRune('E') // normaliza p/ 'E' -- ParseFloat do Go só aceita 'e'/'E'
			sb.WriteString(sign)
			for l.pos < len(l.source) && unicode.IsDigit(l.current()) {
				sb.WriteRune(l.advance())
			}
		} else {
			// Não era um expoente de verdade (ex: "d" logo após um número
			// sem dígito nenhum depois) -- devolve a posição, esse
			// caractere pertence ao próximo token.
			l.pos, l.line, l.col = savedPos, savedLine, savedCol
		}
	}

	switch l.current() {
	case '!':
		isFloat = true
		l.advance()
	case '#':
		isFloat = true
		isDouble = true
		l.advance()
	}

	if !isFloat {
		return Token{
			Type:   TokenNumber,
			Value:  sb.String(),
			Line:   startLine,
			Column: startCol,
		}, nil
	}

	marker := "S"
	if isDouble {
		marker = "D"
	}
	return Token{
		Type:   TokenFloat,
		Value:  sb.String() + marker,
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

// lexRadixNumber generaliza lexHexNumber para os literais octal (&O) e
// binário (&B) -- mesmo padrão de normalizar o prefixo + dígitos em
// maiúsculas no Value do token, que parsePrimary reconhece pelo prefixo.
func (l *Lexer) lexRadixNumber(prefix string, isDigit func(rune) bool) (Token, error) {
	startLine := l.line
	startCol := l.col

	for i := 0; i < len(prefix); i++ {
		l.advance()
	}

	var sb strings.Builder
	for l.pos < len(l.source) && isDigit(l.current()) {
		sb.WriteRune(l.advance())
	}

	if sb.Len() == 0 {
		return Token{}, fmt.Errorf("número %s inválido em %d:%d", prefix, startLine, startCol)
	}

	return Token{
		Type:   TokenNumber,
		Value:  prefix + sb.String(),
		Line:   startLine,
		Column: startCol,
	}, nil
}

func isHexDigit(ch rune) bool {
	return unicode.IsDigit(ch) ||
		(ch >= 'a' && ch <= 'f') ||
		(ch >= 'A' && ch <= 'F')
}

func isOctalDigit(ch rune) bool {
	return ch >= '0' && ch <= '7'
}

func isBinaryDigit(ch rune) bool {
	return ch == '0' || ch == '1'
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
		"SINGLE":    TokenSingle,
		"DOUBLE":    TokenDouble,
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
		"PUT":       TokenPut,
		"SPRITE":    TokenSprite,
		"PATTERN":   TokenPattern,
		"OFF":       TokenOff,
		"PLAY":      TokenPlay,
		"OPEN":      TokenOpen,
		"CLOSE":     TokenClose,
		"OUTPUT":    TokenOutput,
		"APPEND":    TokenAppend,
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
