package kaji80

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// Lexer realiza a análise léxica de código assembly Z80.
type Lexer struct {
	src    []rune
	pos    int
	line   int
	col    int
	length int
}

// NewLexer cria um novo Lexer.
func NewLexer(input string) *Lexer {
	runes := []rune(input)
	return &Lexer{
		src:    runes,
		pos:    0,
		line:   1,
		col:    1,
		length: len(runes),
	}
}

func (l *Lexer) current() rune {
	if l.pos >= l.length {
		return 0
	}
	return l.src[l.pos]
}

func (l *Lexer) peek() rune {
	if l.pos+1 >= l.length {
		return 0
	}
	return l.src[l.pos+1]
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

// NextToken retorna o próximo token.
func (l *Lexer) NextToken() (Token, error) {
	for l.pos < l.length {
		ch := l.current()

		// Ignorar espaços em branco (exceto nova linha)
		if ch == ' ' || ch == '\t' || ch == '\r' {
			l.advance()
			continue
		}

		// Comentários até o fim da linha
		if ch == ';' {
			for l.pos < l.length && l.current() != '\n' {
				l.advance()
			}
			continue
		}

		startLine := l.line
		startCol := l.col

		if ch == '\n' {
			l.advance()
			return Token{Type: TokenNewline, Value: "\n", Line: startLine, Col: startCol}, nil
		}

		if ch == ',' {
			l.advance()
			return Token{Type: TokenComma, Value: ",", Line: startLine, Col: startCol}, nil
		}

		if ch == ':' {
			l.advance()
			return Token{Type: TokenColon, Value: ":", Line: startLine, Col: startCol}, nil
		}

		if ch == '(' {
			l.advance()
			return Token{Type: TokenLParen, Value: "(", Line: startLine, Col: startCol}, nil
		}

		if ch == ')' {
			l.advance()
			return Token{Type: TokenRParen, Value: ")", Line: startLine, Col: startCol}, nil
		}

		if ch == '+' {
			l.advance()
			return Token{Type: TokenPlus, Value: "+", Line: startLine, Col: startCol}, nil
		}

		if ch == '-' {
			l.advance()
			return Token{Type: TokenMinus, Value: "-", Line: startLine, Col: startCol}, nil
		}

		// Strings entre aspas simples ou duplas
		if ch == '"' || ch == '\'' {
			strVal, err := l.readString(ch)
			if err != nil {
				return Token{}, err
			}
			return Token{Type: TokenString, Value: strVal, Line: startLine, Col: startCol}, nil
		}

		// Números iniciados por $, # (hex) ou % (bin)
		if ch == '$' || ch == '#' {
			l.advance()
			hexStr := l.readHexDigits()
			if hexStr == "" {
				return Token{}, fmt.Errorf("linha %d:%d: dígito hexadecimal esperado após '%c'", startLine, startCol, ch)
			}
			val, err := strconv.ParseInt(hexStr, 16, 64)
			if err != nil {
				return Token{}, fmt.Errorf("linha %d:%d: número hexadecimal inválido: %w", startLine, startCol, err)
			}
			return Token{Type: TokenNumber, Value: hexStr, Number: val, Line: startLine, Col: startCol}, nil
		}

		if ch == '%' {
			l.advance()
			binStr := l.readBinDigits()
			if binStr == "" {
				return Token{}, fmt.Errorf("linha %d:%d: dígito binário esperado após '%%'", startLine, startCol)
			}
			val, err := strconv.ParseInt(binStr, 2, 64)
			if err != nil {
				return Token{}, fmt.Errorf("linha %d:%d: número binário inválido: %w", startLine, startCol, err)
			}
			return Token{Type: TokenNumber, Value: binStr, Number: val, Line: startLine, Col: startCol}, nil
		}

		// Números iniciando com dígito decimal (0-9)
		if unicode.IsDigit(ch) {
			numToken, err := l.readNumber()
			if err != nil {
				return Token{}, err
			}
			numToken.Line = startLine
			numToken.Col = startCol
			return numToken, nil
		}

		// Identificadores (labels, mnemônicos, registradores)
		if isIdentStart(ch) {
			ident := l.readIdentifier()
			return Token{Type: TokenIdentifier, Value: ident, Line: startLine, Col: startCol}, nil
		}

		return Token{}, fmt.Errorf("linha %d:%d: caractere inesperado '%c'", startLine, startCol, ch)
	}

	return Token{Type: TokenEOF, Value: "", Line: l.line, Col: l.col}, nil
}

func isIdentStart(ch rune) bool {
	return unicode.IsLetter(ch) || ch == '_' || ch == '.' || ch == '?' || ch == '@'
}

func isIdentPart(ch rune) bool {
	return unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_' || ch == '.' || ch == '?' || ch == '@'
}

func (l *Lexer) readIdentifier() string {
	start := l.pos
	for l.pos < l.length && isIdentPart(l.current()) {
		l.advance()
	}
	return string(l.src[start:l.pos])
}

func (l *Lexer) readHexDigits() string {
	start := l.pos
	for l.pos < l.length && isHexRune(l.current()) {
		l.advance()
	}
	return string(l.src[start:l.pos])
}

func (l *Lexer) readBinDigits() string {
	start := l.pos
	for l.pos < l.length && (l.current() == '0' || l.current() == '1') {
		l.advance()
	}
	return string(l.src[start:l.pos])
}

func isHexRune(ch rune) bool {
	return (ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')
}

func (l *Lexer) readNumber() (Token, error) {
	start := l.pos
	// Ler sequência de dígitos alfanuméricos para suportar sufixos como 0FFH, 1010B, etc.
	for l.pos < l.length && (isHexRune(l.current()) || l.current() == 'h' || l.current() == 'H' || l.current() == 'b' || l.current() == 'B' || l.current() == 'x' || l.current() == 'X') {
		l.advance()
	}
	raw := string(l.src[start:l.pos])

	// Verificar prefixo 0x ou 0b
	if strings.HasPrefix(strings.ToLower(raw), "0x") {
		val, err := strconv.ParseInt(raw[2:], 16, 64)
		if err != nil {
			return Token{}, fmt.Errorf("número hexadecimal inválido %q", raw)
		}
		return Token{Type: TokenNumber, Value: raw, Number: val}, nil
	}
	if strings.HasPrefix(strings.ToLower(raw), "0b") {
		val, err := strconv.ParseInt(raw[2:], 2, 64)
		if err != nil {
			return Token{}, fmt.Errorf("número binário inválido %q", raw)
		}
		return Token{Type: TokenNumber, Value: raw, Number: val}, nil
	}

	// Verificar sufixo 'H' (hexadecimal)
	if strings.HasSuffix(strings.ToLower(raw), "h") {
		hexPart := raw[:len(raw)-1]
		val, err := strconv.ParseInt(hexPart, 16, 64)
		if err != nil {
			return Token{}, fmt.Errorf("número hexadecimal inválido %q", raw)
		}
		return Token{Type: TokenNumber, Value: raw, Number: val}, nil
	}

	// Verificar sufixo 'B' (binário) se todos forem 0 e 1
	if strings.HasSuffix(strings.ToLower(raw), "b") {
		binPart := raw[:len(raw)-1]
		allBin := true
		for _, r := range binPart {
			if r != '0' && r != '1' {
				allBin = false
				break
			}
		}
		if allBin && len(binPart) > 0 {
			val, err := strconv.ParseInt(binPart, 2, 64)
			if err == nil {
				return Token{Type: TokenNumber, Value: raw, Number: val}, nil
			}
		}
	}

	// Tentar decimal comum
	val, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return Token{}, fmt.Errorf("número inválido %q", raw)
	}
	return Token{Type: TokenNumber, Value: raw, Number: val}, nil
}

func (l *Lexer) readString(quote rune) (string, error) {
	l.advance() // Pula abertura
	var sb strings.Builder
	for l.pos < l.length {
		ch := l.advance()
		if ch == quote {
			return sb.String(), nil
		}
		if ch == '\n' {
			return "", fmt.Errorf("linha %d: string não terminada antes da nova linha", l.line-1)
		}
		if ch == '\\' {
			if l.pos >= l.length {
				return "", fmt.Errorf("escape não terminado")
			}
			esc := l.advance()
			switch esc {
			case 'n':
				sb.WriteRune('\n')
			case 'r':
				sb.WriteRune('\r')
			case 't':
				sb.WriteRune('\t')
			case '0':
				sb.WriteByte(0)
			case '\\':
				sb.WriteRune('\\')
			case '\'':
				sb.WriteRune('\'')
			case '"':
				sb.WriteRune('"')
			default:
				sb.WriteRune(esc)
			}
		} else {
			sb.WriteRune(ch)
		}
	}
	return "", fmt.Errorf("string não terminada")
}
