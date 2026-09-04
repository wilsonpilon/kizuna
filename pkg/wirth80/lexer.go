package wirth80

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

var keywords = map[string]TokenType{
	"PROGRAM":   TokenProgram,
	"BEGIN":     TokenBegin,
	"END":       TokenEnd,
	"VAR":       TokenVar,
	"INTEGER":   TokenInteger,
	"CHAR":      TokenChar,
	"BOOLEAN":   TokenBoolean,
	"STRING":    TokenStringKw,
	"WRITE":     TokenWrite,
	"WRITELN":   TokenWriteLn,
	"IF":        TokenIf,
	"THEN":      TokenThen,
	"ELSE":      TokenElse,
	"WHILE":     TokenWhile,
	"DO":        TokenDo,
	"FOR":       TokenFor,
	"TO":        TokenTo,
	"DOWNTO":    TokenDownto,
	"DIV":       TokenDiv,
}

// Lexer realiza a análise léxica de código-fonte Pascal
type Lexer struct {
	src    []rune
	pos    int
	line   int
	column int
}

// NewLexer cria uma nova instância de Lexer para o código-fonte fornecido
func NewLexer(source string) *Lexer {
	return &Lexer{
		src:    []rune(source),
		pos:    0,
		line:   1,
		column: 1,
	}
}

func (l *Lexer) current() rune {
	if l.pos >= len(l.src) {
		return 0
	}
	return l.src[l.pos]
}

func (l *Lexer) peek() rune {
	if l.pos+1 >= len(l.src) {
		return 0
	}
	return l.src[l.pos+1]
}

func (l *Lexer) advance() rune {
	ch := l.current()
	l.pos++
	if ch == '\n' {
		l.line++
		l.column = 1
	} else {
		l.column++
	}
	return ch
}

func (l *Lexer) skipWhitespaceAndComments() error {
	for {
		ch := l.current()
		if ch == 0 {
			break
		}

		// Espaços em branco
		if unicode.IsSpace(ch) {
			l.advance()
			continue
		}

		// Comentário { ... }
		if ch == '{' {
			l.advance()
			for l.current() != 0 && l.current() != '}' {
				l.advance()
			}
			if l.current() == '}' {
				l.advance()
			}
			continue
		}

		// Comentário (* ... *)
		if ch == '(' && l.peek() == '*' {
			l.advance() // (
			l.advance() // *
			for l.current() != 0 {
				if l.current() == '*' && l.peek() == ')' {
					l.advance() // *
					l.advance() // )
					break
				}
				l.advance()
			}
			continue
		}

		// Comentário // ... até o fim da linha
		if ch == '/' && l.peek() == '/' {
			l.advance()
			l.advance()
			for l.current() != 0 && l.current() != '\n' {
				l.advance()
			}
			continue
		}

		break
	}
	return nil
}

// NextToken retorna o próximo token da fonte
func (l *Lexer) NextToken() (Token, error) {
	if err := l.skipWhitespaceAndComments(); err != nil {
		return Token{}, err
	}

	startLine := l.line
	startCol := l.column
	ch := l.current()

	if ch == 0 {
		return Token{Type: TokenEOF, Line: startLine, Column: startCol}, nil
	}

	// Identificador ou palavra-chave
	if unicode.IsLetter(ch) || ch == '_' {
		var sb strings.Builder
		for unicode.IsLetter(l.current()) || unicode.IsDigit(l.current()) || l.current() == '_' {
			sb.WriteRune(l.advance())
		}
		text := sb.String()
		upper := strings.ToUpper(text)
		if tt, ok := keywords[upper]; ok {
			return Token{Type: tt, Value: text, Line: startLine, Column: startCol}, nil
		}
		return Token{Type: TokenIdent, Value: text, Line: startLine, Column: startCol}, nil
	}

	// Número decimal ou hexadecimal ($FF)
	if unicode.IsDigit(ch) || ch == '$' {
		var sb strings.Builder
		isHex := false
		if ch == '$' {
			isHex = true
			l.advance()
		}
		for {
			cur := l.current()
			if isHex {
				if unicode.IsDigit(cur) || (cur >= 'a' && cur <= 'f') || (cur >= 'A' && cur <= 'F') {
					sb.WriteRune(l.advance())
				} else {
					break
				}
			} else {
				if unicode.IsDigit(cur) {
					sb.WriteRune(l.advance())
				} else {
					break
				}
			}
		}
		strVal := sb.String()
		var num int64
		var err error
		if isHex {
			num, err = strconv.ParseInt(strVal, 16, 64)
		} else {
			num, err = strconv.ParseInt(strVal, 10, 64)
		}
		if err != nil {
			return Token{}, fmt.Errorf("número inválido na linha %d:%d: '%s'", startLine, startCol, strVal)
		}
		return Token{Type: TokenNumber, Value: strVal, Number: num, Line: startLine, Column: startCol}, nil
	}

	// String literal ('exemplo' ou '' para aspa simples)
	if ch == '\'' {
		l.advance() // abre aspa
		var sb strings.Builder
		for {
			cur := l.current()
			if cur == 0 {
				return Token{}, fmt.Errorf("string não terminada iniciada na linha %d:%d", startLine, startCol)
			}
			if cur == '\'' {
				l.advance()
				if l.current() == '\'' {
					// Duas aspas seguidas representam um literal de aspa
					sb.WriteRune('\'')
					l.advance()
				} else {
					// Fim da string
					break
				}
			} else {
				sb.WriteRune(l.advance())
			}
		}
		return Token{Type: TokenString, Value: sb.String(), Line: startLine, Column: startCol}, nil
	}

	// Operadores de 2 caracteres
	if ch == ':' && l.peek() == '=' {
		l.advance()
		l.advance()
		return Token{Type: TokenAssign, Value: ":=", Line: startLine, Column: startCol}, nil
	}
	if ch == '<' && l.peek() == '>' {
		l.advance()
		l.advance()
		return Token{Type: TokenNotEqual, Value: "<>", Line: startLine, Column: startCol}, nil
	}
	if ch == '<' && l.peek() == '=' {
		l.advance()
		l.advance()
		return Token{Type: TokenLessEq, Value: "<=", Line: startLine, Column: startCol}, nil
	}
	if ch == '>' && l.peek() == '=' {
		l.advance()
		l.advance()
		return Token{Type: TokenGreaterEq, Value: ">=", Line: startLine, Column: startCol}, nil
	}

	// Operadores e delimitadores simples
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
	case '=':
		return Token{Type: TokenEqual, Value: "=", Line: startLine, Column: startCol}, nil
	case '<':
		return Token{Type: TokenLess, Value: "<", Line: startLine, Column: startCol}, nil
	case '>':
		return Token{Type: TokenGreater, Value: ">", Line: startLine, Column: startCol}, nil
	case '(':
		return Token{Type: TokenLParen, Value: "(", Line: startLine, Column: startCol}, nil
	case ')':
		return Token{Type: TokenRParen, Value: ")", Line: startLine, Column: startCol}, nil
	case ';':
		return Token{Type: TokenSemi, Value: ";", Line: startLine, Column: startCol}, nil
	case ':':
		return Token{Type: TokenColon, Value: ":", Line: startLine, Column: startCol}, nil
	case ',':
		return Token{Type: TokenComma, Value: ",", Line: startLine, Column: startCol}, nil
	case '.':
		return Token{Type: TokenDot, Value: ".", Line: startLine, Column: startCol}, nil
	default:
		return Token{}, fmt.Errorf("caractere inesperado '%c' (0x%02X) na linha %d:%d", ch, ch, startLine, startCol)
	}
}
