package kaji80

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"
)

// =============================================================================
// KIZUNA KAJI80 - Avaliador de expressões numéricas em tempo de montagem
//
// Suporta aritmética, bits, lógica e funções matemáticas sobre números puros
// (literais, PI, EQU já definidas, variáveis reatribuíveis "Nome = expr").
// NÃO suporta aritmética envolvendo um rótulo/símbolo relocável (EXTERN ou
// label do próprio módulo) -- "EQU MinhaConst + Rotulo" continua um erro
// claro, decisão deliberada: o endereço de um símbolo EXTERN só é conhecido
// depois da linkagem pelo MUSUBI, e misturar isso com aritmética exigiria
// estender o formato de relocation do .MOB pra carregar um deslocamento
// (addend) -- mudança de formato maior, fora do escopo desta leva. Mesma
// razão pela qual "Label+N" (bug já documentado em sessões anteriores)
// continua não suportado.
//
// Precedência igual ao C (do mais apertado pro mais solto):
//   unário (- + NOT ~)  >  * / MOD  >  + -  >  << >>  >  < <= > >=
//   >  == !=  >  &  >  ^  >  |  >  &&  >  ||
//
// Nota sobre '%': o KAJI80 já usa '%' como prefixo de literal binário
// (%1010, documentado em docs/manual-assembly.md) -- diferente do asMSX,
// que usa '%' como operador de módulo. Pra não quebrar essa sintaxe já
// estabelecida, o módulo aqui é a palavra-chave MOD, igual ao que
// DIGNAC/WIRTH80 já usam (TokenMod) -- consistência entre as linguagens da
// toolchain, não uma cópia literal do asMSX ("nossa própria sintaxe").
// =============================================================================

type exprTokKind int

const (
	etNum exprTokKind = iota
	etIdent
	etPlus
	etMinus
	etStar
	etSlash
	etMod
	etShl
	etShr
	etPipe
	etAmp
	etCaret
	etTilde
	etNot
	etOrOr
	etAndAnd
	etEqEq
	etNotEq
	etLt
	etLtEq
	etGt
	etGtEq
	etLParen
	etRParen
	etComma
	etEOF
)

type exprTok struct {
	kind exprTokKind
	text string
	num  float64
}

// tokenizeExpr faz sua própria varredura independente do lexer principal do
// KAJI80 -- os operandos já chegam reconstruídos como texto puro por
// parseLine, então não há necessidade (nem vantagem) de reusar o stream de
// tokens original.
func tokenizeExpr(s string) ([]exprTok, error) {
	var toks []exprTok
	i := 0
	n := len(s)
	for i < n {
		c := s[i]
		switch {
		case c == ' ' || c == '\t':
			i++
		case c == '(':
			toks = append(toks, exprTok{kind: etLParen, text: "("})
			i++
		case c == ')':
			toks = append(toks, exprTok{kind: etRParen, text: ")"})
			i++
		case c == ',':
			toks = append(toks, exprTok{kind: etComma, text: ","})
			i++
		case c == '+':
			toks = append(toks, exprTok{kind: etPlus, text: "+"})
			i++
		case c == '-':
			toks = append(toks, exprTok{kind: etMinus, text: "-"})
			i++
		case c == '*':
			toks = append(toks, exprTok{kind: etStar, text: "*"})
			i++
		case c == '/':
			toks = append(toks, exprTok{kind: etSlash, text: "/"})
			i++
		case c == '~':
			toks = append(toks, exprTok{kind: etTilde, text: "~"})
			i++
		case c == '^':
			toks = append(toks, exprTok{kind: etCaret, text: "^"})
			i++
		case c == '<':
			if i+1 < n && s[i+1] == '<' {
				toks = append(toks, exprTok{kind: etShl, text: "<<"})
				i += 2
			} else if i+1 < n && s[i+1] == '=' {
				toks = append(toks, exprTok{kind: etLtEq, text: "<="})
				i += 2
			} else {
				toks = append(toks, exprTok{kind: etLt, text: "<"})
				i++
			}
		case c == '>':
			if i+1 < n && s[i+1] == '>' {
				toks = append(toks, exprTok{kind: etShr, text: ">>"})
				i += 2
			} else if i+1 < n && s[i+1] == '=' {
				toks = append(toks, exprTok{kind: etGtEq, text: ">="})
				i += 2
			} else {
				toks = append(toks, exprTok{kind: etGt, text: ">"})
				i++
			}
		case c == '=':
			if i+1 < n && s[i+1] == '=' {
				toks = append(toks, exprTok{kind: etEqEq, text: "=="})
				i += 2
			} else {
				return nil, fmt.Errorf("'=' isolado não é válido dentro de uma expressão (use '==' pra comparar)")
			}
		case c == '!':
			if i+1 < n && s[i+1] == '=' {
				toks = append(toks, exprTok{kind: etNotEq, text: "!="})
				i += 2
			} else {
				return nil, fmt.Errorf("'!' isolado não é um operador válido (use NOT ou '!=')")
			}
		case c == '|':
			if i+1 < n && s[i+1] == '|' {
				toks = append(toks, exprTok{kind: etOrOr, text: "||"})
				i += 2
			} else {
				toks = append(toks, exprTok{kind: etPipe, text: "|"})
				i++
			}
		case c == '&':
			if i+1 < n && s[i+1] == '&' {
				toks = append(toks, exprTok{kind: etAndAnd, text: "&&"})
				i += 2
			} else {
				toks = append(toks, exprTok{kind: etAmp, text: "&"})
				i++
			}
		case c >= '0' && c <= '9':
			start := i
			isFloat := false
			for i < n && isExprNumChar(s[i]) {
				i++
			}
			// Ponto decimal (float) -- só se seguido de dígito, pra não
			// engolir um '.' que na verdade começa outra coisa.
			if i < n && s[i] == '.' && i+1 < n && s[i+1] >= '0' && s[i+1] <= '9' {
				isFloat = true
				i++
				for i < n && s[i] >= '0' && s[i] <= '9' {
					i++
				}
			}
			raw := s[start:i]
			val, err := parseExprNumber(raw, isFloat)
			if err != nil {
				return nil, err
			}
			toks = append(toks, exprTok{kind: etNum, num: val, text: raw})
		case c == '\'':
			// Literal de caractere: 'A'
			if i+2 >= n || s[i+2] != '\'' {
				return nil, fmt.Errorf("literal de caractere mal formado em %q (esperado 'X')", s)
			}
			toks = append(toks, exprTok{kind: etNum, num: float64(s[i+1]), text: s[i : i+3]})
			i += 3
		case c == '$':
			// $HEX (convenção do KAJI80 -- ver manual-assembly.md)
			start := i
			i++
			for i < n && isHexRune(rune(s[i])) {
				i++
			}
			if i == start+1 {
				return nil, fmt.Errorf("dígito hexadecimal esperado após '$' em %q", s)
			}
			val, err := strconv.ParseInt(s[start+1:i], 16, 64)
			if err != nil {
				return nil, fmt.Errorf("número hexadecimal inválido: %q", s[start:i])
			}
			toks = append(toks, exprTok{kind: etNum, num: float64(val), text: s[start:i]})
		case isExprIdentStart(c):
			start := i
			for i < n && isExprIdentPart(s[i]) {
				i++
			}
			word := s[start:i]
			switch strings.ToUpper(word) {
			case "MOD":
				toks = append(toks, exprTok{kind: etMod, text: word})
			case "NOT":
				toks = append(toks, exprTok{kind: etNot, text: word})
			default:
				toks = append(toks, exprTok{kind: etIdent, text: word})
			}
		default:
			return nil, fmt.Errorf("caractere inesperado %q na expressão %q", c, s)
		}
	}
	toks = append(toks, exprTok{kind: etEOF})
	return toks, nil
}

func isExprNumChar(c byte) bool {
	return (c >= '0' && c <= '9') || c == 'h' || c == 'H' || c == 'x' || c == 'X' ||
		(c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F') || c == 'o' || c == 'O' || c == 'b' || c == 'B'
}

func isExprIdentStart(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_'
}

func isExprIdentPart(c byte) bool {
	return isExprIdentStart(c) || (c >= '0' && c <= '9')
}

// parseExprNumber entende os mesmos formatos já documentados pro KAJI80
// (decimal, hex 0x../..h, binário ..b) mais octal (..o, convenção sufixo
// apenas -- o prefixo "0" puro do asMSX é ambíguo com decimal e o próprio
// asMSX desaconselha a forma "O" maiúscula por confundir com zero) e ponto
// flutuante decimal (necessário pra funções trigonométricas fazerem
// sentido, ex. sin(pi*45.0/180.0)).
func parseExprNumber(raw string, isFloat bool) (float64, error) {
	if isFloat {
		v, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return 0, fmt.Errorf("número decimal inválido: %q", raw)
		}
		return v, nil
	}
	lower := strings.ToLower(raw)
	switch {
	case strings.HasPrefix(lower, "0x"):
		v, err := strconv.ParseInt(raw[2:], 16, 64)
		if err != nil {
			return 0, fmt.Errorf("número hexadecimal inválido: %q", raw)
		}
		return float64(v), nil
	case strings.HasSuffix(lower, "h"):
		v, err := strconv.ParseInt(raw[:len(raw)-1], 16, 64)
		if err != nil {
			return 0, fmt.Errorf("número hexadecimal inválido: %q", raw)
		}
		return float64(v), nil
	case strings.HasSuffix(lower, "b"):
		v, err := strconv.ParseInt(raw[:len(raw)-1], 2, 64)
		if err != nil {
			return 0, fmt.Errorf("número binário inválido: %q", raw)
		}
		return float64(v), nil
	case strings.HasSuffix(lower, "o"):
		v, err := strconv.ParseInt(raw[:len(raw)-1], 8, 64)
		if err != nil {
			return 0, fmt.Errorf("número octal inválido: %q", raw)
		}
		return float64(v), nil
	default:
		v, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("número decimal inválido: %q", raw)
		}
		return float64(v), nil
	}
}

// exprParser é um parser recursivo-descendente clássico com um nível de
// função por precedência (precedence climbing).
type exprParser struct {
	toks []exprTok
	pos  int
	asm  *Assembler
}

func (p *exprParser) cur() exprTok { return p.toks[p.pos] }
func (p *exprParser) advance() exprTok {
	t := p.toks[p.pos]
	if p.pos < len(p.toks)-1 {
		p.pos++
	}
	return t
}

func (p *exprParser) parseLogicalOr() (float64, error) {
	left, err := p.parseLogicalAnd()
	if err != nil {
		return 0, err
	}
	for p.cur().kind == etOrOr {
		p.advance()
		right, err := p.parseLogicalAnd()
		if err != nil {
			return 0, err
		}
		left = boolToFloat(left != 0 || right != 0)
	}
	return left, nil
}

func (p *exprParser) parseLogicalAnd() (float64, error) {
	left, err := p.parseBitOr()
	if err != nil {
		return 0, err
	}
	for p.cur().kind == etAndAnd {
		p.advance()
		right, err := p.parseBitOr()
		if err != nil {
			return 0, err
		}
		left = boolToFloat(left != 0 && right != 0)
	}
	return left, nil
}

func (p *exprParser) parseBitOr() (float64, error) {
	left, err := p.parseBitXor()
	if err != nil {
		return 0, err
	}
	for p.cur().kind == etPipe {
		p.advance()
		right, err := p.parseBitXor()
		if err != nil {
			return 0, err
		}
		left = float64(int64(left) | int64(right))
	}
	return left, nil
}

func (p *exprParser) parseBitXor() (float64, error) {
	left, err := p.parseBitAnd()
	if err != nil {
		return 0, err
	}
	for p.cur().kind == etCaret {
		p.advance()
		right, err := p.parseBitAnd()
		if err != nil {
			return 0, err
		}
		left = float64(int64(left) ^ int64(right))
	}
	return left, nil
}

func (p *exprParser) parseBitAnd() (float64, error) {
	left, err := p.parseEquality()
	if err != nil {
		return 0, err
	}
	for p.cur().kind == etAmp {
		p.advance()
		right, err := p.parseEquality()
		if err != nil {
			return 0, err
		}
		left = float64(int64(left) & int64(right))
	}
	return left, nil
}

func (p *exprParser) parseEquality() (float64, error) {
	left, err := p.parseRelational()
	if err != nil {
		return 0, err
	}
	for p.cur().kind == etEqEq || p.cur().kind == etNotEq {
		op := p.advance().kind
		right, err := p.parseRelational()
		if err != nil {
			return 0, err
		}
		if op == etEqEq {
			left = boolToFloat(left == right)
		} else {
			left = boolToFloat(left != right)
		}
	}
	return left, nil
}

func (p *exprParser) parseRelational() (float64, error) {
	left, err := p.parseShift()
	if err != nil {
		return 0, err
	}
	for {
		switch p.cur().kind {
		case etLt:
			p.advance()
			right, err := p.parseShift()
			if err != nil {
				return 0, err
			}
			left = boolToFloat(left < right)
		case etLtEq:
			p.advance()
			right, err := p.parseShift()
			if err != nil {
				return 0, err
			}
			left = boolToFloat(left <= right)
		case etGt:
			p.advance()
			right, err := p.parseShift()
			if err != nil {
				return 0, err
			}
			left = boolToFloat(left > right)
		case etGtEq:
			p.advance()
			right, err := p.parseShift()
			if err != nil {
				return 0, err
			}
			left = boolToFloat(left >= right)
		default:
			return left, nil
		}
	}
}

func (p *exprParser) parseShift() (float64, error) {
	left, err := p.parseAdditive()
	if err != nil {
		return 0, err
	}
	for p.cur().kind == etShl || p.cur().kind == etShr {
		op := p.advance().kind
		right, err := p.parseAdditive()
		if err != nil {
			return 0, err
		}
		if op == etShl {
			left = float64(int64(left) << uint(int64(right)))
		} else {
			left = float64(int64(left) >> uint(int64(right)))
		}
	}
	return left, nil
}

func (p *exprParser) parseAdditive() (float64, error) {
	left, err := p.parseMultiplicative()
	if err != nil {
		return 0, err
	}
	for p.cur().kind == etPlus || p.cur().kind == etMinus {
		op := p.advance().kind
		right, err := p.parseMultiplicative()
		if err != nil {
			return 0, err
		}
		if op == etPlus {
			left += right
		} else {
			left -= right
		}
	}
	return left, nil
}

func (p *exprParser) parseMultiplicative() (float64, error) {
	left, err := p.parseUnary()
	if err != nil {
		return 0, err
	}
	for p.cur().kind == etStar || p.cur().kind == etSlash || p.cur().kind == etMod {
		op := p.advance().kind
		right, err := p.parseUnary()
		if err != nil {
			return 0, err
		}
		switch op {
		case etStar:
			left *= right
		case etSlash:
			if right == 0 {
				return 0, fmt.Errorf("divisão por zero na expressão")
			}
			left /= right
		case etMod:
			if int64(right) == 0 {
				return 0, fmt.Errorf("módulo por zero na expressão")
			}
			left = float64(int64(left) % int64(right))
		}
	}
	return left, nil
}

func (p *exprParser) parseUnary() (float64, error) {
	switch p.cur().kind {
	case etMinus:
		p.advance()
		v, err := p.parseUnary()
		if err != nil {
			return 0, err
		}
		return -v, nil
	case etPlus:
		p.advance()
		return p.parseUnary()
	case etTilde:
		p.advance()
		v, err := p.parseUnary()
		if err != nil {
			return 0, err
		}
		return float64(^int64(v)), nil
	case etNot:
		p.advance()
		v, err := p.parseUnary()
		if err != nil {
			return 0, err
		}
		return boolToFloat(v == 0), nil
	default:
		return p.parsePrimary()
	}
}

func (p *exprParser) parsePrimary() (float64, error) {
	tok := p.cur()
	switch tok.kind {
	case etNum:
		p.advance()
		return tok.num, nil
	case etLParen:
		p.advance()
		v, err := p.parseLogicalOr()
		if err != nil {
			return 0, err
		}
		if p.cur().kind != etRParen {
			return 0, fmt.Errorf("esperado ')' na expressão")
		}
		p.advance()
		return v, nil
	case etIdent:
		name := tok.text
		p.advance()
		if p.cur().kind == etLParen {
			return p.parseFuncCall(name)
		}
		return p.lookupIdent(name)
	default:
		return 0, fmt.Errorf("token inesperado %q na expressão", tok.text)
	}
}

func (p *exprParser) parseArgs() ([]float64, error) {
	var args []float64
	if p.cur().kind == etRParen {
		return args, nil
	}
	for {
		v, err := p.parseLogicalOr()
		if err != nil {
			return nil, err
		}
		args = append(args, v)
		if p.cur().kind == etComma {
			p.advance()
			continue
		}
		break
	}
	return args, nil
}

func (p *exprParser) parseFuncCall(name string) (float64, error) {
	p.advance() // consome '('
	args, err := p.parseArgs()
	if err != nil {
		return 0, err
	}
	if p.cur().kind != etRParen {
		return 0, fmt.Errorf("esperado ')' depois dos argumentos de %s(...)", name)
	}
	p.advance()
	return callExprFunc(name, args)
}

func (p *exprParser) lookupIdent(name string) (float64, error) {
	if strings.EqualFold(name, "PI") {
		return math.Pi, nil
	}
	if p.asm != nil {
		if v, ok := p.asm.lookupNumericName(name); ok {
			return v, nil
		}
	}
	return 0, fmt.Errorf("'%s' não é uma constante, variável ou função conhecida (aritmética com rótulo/símbolo EXTERN não é suportada)", name)
}

func boolToFloat(b bool) float64 {
	if b {
		return 1
	}
	return 0
}

// exprFuncArity documenta a assinatura de cada função suportada, usado só
// pra gerar mensagens de erro claras (a checagem de fato é feita em
// callExprFunc, já que Go não tem sobrecarga por aridade).
var exprFuncArity = map[string]int{
	"SIN": 1, "COS": 1, "TAN": 1, "ASIN": 1, "ACOS": 1, "ATAN": 1,
	"SQR": 1, "SQRT": 1, "EXP": 1, "LOG": 1, "LN": 1, "ABS": 1,
	"INT": 1, "FIX": 1, "RANDOM": 1,
	"POW": 2, "FIXMUL": 2, "FIXDIV": 2,
}

// callExprFunc despacha as funções matemáticas suportadas. FIX converte um
// float pra ponto fixo 8.8 (1 byte inteiro + 1 byte fração, convenção comum
// em jogos Z80) -- o valor devolvido já é o inteiro de 16 bits pronto pra
// DW. FIXMUL/FIXDIV operam sobre valores JÁ em ponto fixo 8.8 (ex.:
// FIXMUL(FIX(1.5), FIX(2.0))), não sobre floats crus.
func callExprFunc(name string, args []float64) (float64, error) {
	upper := strings.ToUpper(name)
	arity, ok := exprFuncArity[upper]
	if !ok {
		return 0, fmt.Errorf("função desconhecida '%s(...)' numa expressão", name)
	}
	if len(args) != arity {
		return 0, fmt.Errorf("%s(...) espera %d argumento(s), recebeu %d", upper, arity, len(args))
	}
	switch upper {
	case "SIN":
		return math.Sin(args[0]), nil
	case "COS":
		return math.Cos(args[0]), nil
	case "TAN":
		return math.Tan(args[0]), nil
	case "ASIN":
		return math.Asin(args[0]), nil
	case "ACOS":
		return math.Acos(args[0]), nil
	case "ATAN":
		return math.Atan(args[0]), nil
	case "SQR":
		return args[0] * args[0], nil
	case "SQRT":
		if args[0] < 0 {
			return 0, fmt.Errorf("SQRT de número negativo")
		}
		return math.Sqrt(args[0]), nil
	case "EXP":
		return math.Exp(args[0]), nil
	case "LOG":
		if args[0] <= 0 {
			return 0, fmt.Errorf("LOG de número não-positivo")
		}
		return math.Log10(args[0]), nil
	case "LN":
		if args[0] <= 0 {
			return 0, fmt.Errorf("LN de número não-positivo")
		}
		return math.Log(args[0]), nil
	case "ABS":
		return math.Abs(args[0]), nil
	case "INT":
		return math.Trunc(args[0]), nil
	case "FIX":
		return math.Round(args[0] * 256), nil
	case "FIXMUL":
		return math.Round((args[0] * args[1]) / 256), nil
	case "FIXDIV":
		if args[1] == 0 {
			return 0, fmt.Errorf("FIXDIV por zero")
		}
		return math.Round((args[0] * 256) / args[1]), nil
	case "POW":
		return math.Pow(args[0], args[1]), nil
	case "RANDOM":
		n := int64(args[0])
		if n <= 0 {
			return 0, fmt.Errorf("RANDOM(n) precisa de n > 0")
		}
		return float64(globalExprRand.Int63n(n)), nil
	}
	return 0, fmt.Errorf("função '%s' declarada mas não implementada", upper)
}

// globalExprRand é a fonte de números pseudoaleatórios de RANDOM(n) --
// semeada por execução de montagem (não em runtime: é uma constante
// embutida no binário, igual ao .RANDOM do asMSX). SeedExprRandom existe
// pra testes determinísticos.
var globalExprRand = rand.New(rand.NewSource(1))

// SeedExprRandom fixa a semente de RANDOM(n) -- usado por testes que
// precisam de resultado determinístico. Chamadas reais de montagem podem
// ignorar isso (a semente default já é fixa por simplicidade; ficou de
// fora da leva a decisão de semear por horário real, já que RANDOM(n)
// nesta leva serve principalmente pra tabelas de "ruído" em tempo de
// montagem, não pra imprevisibilidade entre builds).
func SeedExprRandom(seed int64) {
	globalExprRand = rand.New(rand.NewSource(seed))
}

// EvalExpr avalia uma expressão numérica pura. ok=false (err=nil) significa
// "a entrada inteira é um único identificador não reconhecido como
// constante/variável/PI" -- o chamador deve tratar como um nome de símbolo
// comum (rótulo/EXTERN), não como expressão. Qualquer outro problema
// (sintaxe malformada, ou um nome desconhecido usado DENTRO de uma
// expressão maior, ex. "1 + RotuloQualquer") é um erro real: misturar
// aritmética com símbolo relocável não é suportado, e a mensagem deixa
// isso claro em vez de silenciosamente virar zero ou um símbolo fantasma
// (mesma classe de bug já corrigida nesta sessão em outros pontos do
// KAJI80 -- ver auditoria de emitAddressOrReloc/encodeAlu8).
func (a *Assembler) EvalExpr(s string) (float64, bool, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, false, fmt.Errorf("expressão vazia")
	}
	toks, err := tokenizeExpr(s)
	if err != nil {
		return 0, false, err
	}
	// Um único identificador solto (ex.: "MinhaLabel") -- se não for uma
	// constante/variável conhecida, deixa o chamador tratar como símbolo.
	if len(toks) == 2 && toks[0].kind == etIdent {
		if !strings.EqualFold(toks[0].text, "PI") {
			if _, ok := a.lookupNumericName(toks[0].text); !ok {
				return 0, false, nil
			}
		}
	}
	p := &exprParser{toks: toks, pos: 0, asm: a}
	val, err := p.parseLogicalOr()
	if err != nil {
		return 0, false, err
	}
	if p.cur().kind != etEOF {
		return 0, false, fmt.Errorf("token inesperado '%s' na expressão %q", p.cur().text, s)
	}
	return val, true, nil
}

// lookupNumericName procura um nome em constantes EQU e variáveis
// reatribuíveis (Nome = expressão) -- os dois únicos tipos de nome que são
// puros números conhecidos em tempo de montagem, nunca um rótulo/EXTERN
// relocável.
func (a *Assembler) lookupNumericName(name string) (float64, bool) {
	if v, ok := a.constants[name]; ok {
		return float64(v), true
	}
	if a.variables != nil {
		if v, ok := a.variables[name]; ok {
			return v, true
		}
	}
	// Nome pré-definido (BIOS/BDOS/BIOSVARS): último recurso, o código do
	// usuário (rótulo do arquivo ou EXTERN) sempre tem prioridade.
	if _, isLocal := a.symbols[name]; !isLocal {
		if _, isExtern := a.externs[name]; !isExtern {
			// Literal hexadecimal com sufixo h sem o 0 na frente (F0h, FFh):
			// o lexer o vê como identificador. Só vale se o nome NÃO foi
			// declarado como rótulo/EXTERN (Each, Beach seriam números).
			if len(name) > 1 && (name[len(name)-1] == 'h' || name[len(name)-1] == 'H') {
				if hv, err := strconv.ParseUint(name[:len(name)-1], 16, 32); err == nil {
					return float64(hv), true
				}
			}
			if v, ok := lookupPredefined(name); ok {
				return float64(v), true
			}
		}
	}
	return 0, false
}
