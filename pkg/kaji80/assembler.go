package kaji80

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/wilsonpilon/kizuna/pkg/mob"
)

// Registradores de 8 bits: B=0, C=1, D=2, E=3, H=4, L=5, (HL)=6, A=7
var reg8Map = map[string]uint8{
	"B": 0, "C": 1, "D": 2, "E": 3, "H": 4, "L": 5, "A": 7,
}

// Pares de registradores de 16 bits: BC=0, DE=1, HL=2, SP=3
var reg16Map = map[string]uint8{
	"BC": 0, "DE": 1, "HL": 2, "SP": 3,
}

// Pares para PUSH/POP: BC=0, DE=1, HL=2, AF=3
var reg16PushPopMap = map[string]uint8{
	"BC": 0, "DE": 1, "HL": 2, "AF": 3,
}

// Condições para JR/CALL/JP/RET: NZ=0, Z=1, NC=2, C=3, PO=4, PE=5, P=6, M=7
var condMap = map[string]uint8{
	"NZ": 0, "Z": 1, "NC": 2, "C": 3,
	"PO": 4, "PE": 5, "P": 6, "M": 7,
}

// ALU ops: ADD=0, ADC=1, SUB=2, SBC=3, AND=4, XOR=5, OR=6, CP=7
var aluMap = map[string]uint8{
	"ADD": 0, "ADC": 1, "SUB": 2, "SBC": 3,
	"AND": 4, "XOR": 5, "OR": 6, "CP": 7,
}

// Linha intermediária parsed
type parsedLine struct {
	lineNum   int
	label     string
	mnemonic  string
	operands  []string
	rawTokens []Token
}

// Assembler é o montador Z80 do KAJI80.
type Assembler struct {
	moduleName string
	bank       uint8
	publics    map[string]bool
	externs    map[string]bool
	constants  map[string]int64  // Constantes definidas por EQU
	variables  map[string]float64 // Variáveis reatribuíveis ("Nome = expressão")
	symbols    map[string]uint16 // label -> offset no segmento
	codeBytes  []byte
	relocs     []tempReloc

	baseDir     string            // diretório base pra resolver caminhos de INCBIN
	incbinCache map[string][]byte // evita reler o mesmo arquivo entre Pass 1/2
}

// SetBaseDir define o diretório usado pra resolver caminhos relativos de
// INCBIN (ex.: "sprites.bin" ao lado do .asm) -- o chamador (cmd/kaji80)
// passa o diretório do arquivo-fonte. Se nunca chamado, caminhos relativos
// de INCBIN são resolvidos a partir do diretório de trabalho do processo.
func (a *Assembler) SetBaseDir(dir string) {
	a.baseDir = dir
}

type tempReloc struct {
	offset     uint16
	symbolName string
	relocType  mob.RelocType
}

// NewAssembler cria uma nova instância de Assembler.
func NewAssembler() *Assembler {
	return &Assembler{
		moduleName: "MAIN",
		bank:       0,
		publics:    make(map[string]bool),
		externs:    make(map[string]bool),
		constants:  make(map[string]int64),
		variables:  make(map[string]float64),
		symbols:    make(map[string]uint16),
		codeBytes:  make([]byte, 0),
		relocs:     make([]tempReloc, 0),
	}
}

// Assemble monta o código-fonte assembly Z80 e produz um ObjectFile .MOB.
func (a *Assembler) Assemble(source string) (*mob.ObjectFile, error) {
	lines, err := a.tokenizeLines(source)
	if err != nil {
		return nil, err
	}

	// Passo 1: Descobrir Module, Bank, Publics, Externs e calcular tamanhos de instruções
	// para resolver posições dos labels locais.
	labelOffsets := make(map[string]uint16)
	currentOffset := uint16(0)

	for _, line := range lines {
		if line.label != "" {
			labelOffsets[line.label] = currentOffset
		}

		if line.mnemonic == "" {
			continue
		}

		upperMnem := strings.ToUpper(line.mnemonic)
		switch upperMnem {
		case "MODULE":
			if len(line.operands) > 0 {
				a.moduleName = line.operands[0]
			}
		case "BANK":
			if len(line.operands) > 0 {
				var b int
				_, _ = fmt.Sscanf(line.operands[0], "%d", &b)
				a.bank = uint8(b)
			}
		case "PUBLIC":
			for _, op := range line.operands {
				a.publics[op] = true
			}
		case "EXTERN":
			for _, op := range line.operands {
				a.externs[op] = true
			}
		case "ENDMOD", "END":
			// Fim
		case "ORG":
			if len(line.operands) > 0 {
				var orgVal uint16
				_, _ = fmt.Sscanf(line.operands[0], "%v", &orgVal)
				currentOffset = orgVal
			}
		case "EQU":
			if line.label != "" && len(line.operands) > 0 {
				val, ok, err := a.EvalExpr(line.operands[0])
				if err != nil {
					return nil, fmt.Errorf("linha %d: EQU %s: %w", line.lineNum, line.label, err)
				}
				if !ok {
					return nil, fmt.Errorf("linha %d: EQU %s: '%s' não é uma expressão numérica válida (EQU não aceita nome de símbolo/rótulo)", line.lineNum, line.label, line.operands[0])
				}
				a.constants[line.label] = int64(math.Round(val))
			}
		case "ASSIGN":
			// "Nome = expressão" -- variável reatribuível, distinta de EQU
			// (que só permite definir uma vez). Avaliada de novo no Pass 2
			// (ver encodeInstruction) na mesma ordem sequencial, porque o
			// valor pode mudar várias vezes ao longo do arquivo e DB/DW
			// precisam enxergar o valor certo NO PONTO em que aparecem, não
			// o valor final depois do Pass 1 inteiro já ter rodado.
			if line.label != "" && len(line.operands) > 0 {
				val, ok, err := a.EvalExpr(line.operands[0])
				if err != nil {
					return nil, fmt.Errorf("linha %d: %s = ...: %w", line.lineNum, line.label, err)
				}
				if !ok {
					return nil, fmt.Errorf("linha %d: %s = ...: '%s' não é uma expressão numérica válida", line.lineNum, line.label, line.operands[0])
				}
				a.variables[line.label] = val
			}
		default:
			// Instrução ou diretiva de dados: estimar tamanho
			size, err := a.estimateSize(upperMnem, line.operands, line.rawTokens)
			if err != nil {
				return nil, fmt.Errorf("linha %d: erro ao analisar instrução '%s': %w", line.lineNum, line.mnemonic, err)
			}
			currentOffset += size
		}
	}

	a.symbols = labelOffsets

	// Passo 2: Codificar instruções e gerar relocações. "variables" (Nome =
	// expressão) precisa ser reconstruída do zero aqui -- o Pass 1 já a
	// deixou no valor FINAL depois de percorrer o arquivo inteiro, mas o
	// Pass 2 precisa dos valores INTERMEDIÁRIOS, no ponto exato em que cada
	// DB/DW aparece (ex.: X=0 / DB X / X=X+1 / DB X -- os dois DB devem
	// emitir valores diferentes). "ASSIGN" por isso NÃO está na lista de
	// diretivas "já tratadas" abaixo -- tem seu próprio case mais adiante,
	// que reavalia e atualiza a.variables na mesma ordem sequencial.
	a.variables = make(map[string]float64)
	a.codeBytes = make([]byte, 0, currentOffset)
	for _, line := range lines {
		if line.mnemonic == "" {
			continue
		}
		upperMnem := strings.ToUpper(line.mnemonic)
		switch upperMnem {
		case "MODULE", "BANK", "PUBLIC", "EXTERN", "ENDMOD", "END", "ORG", "EQU":
			// Diretivas já tratadas
			continue
		default:
			// Verificação de consistência interna: o tamanho que o Pass 1
			// estimou para esta linha (usado para calcular o endereço de
			// TODO rótulo declarado depois dela) precisa bater exatamente
			// com o número de bytes que o Pass 2 realmente emite agora.
			// Uma divergência aqui corrompe silenciosamente a tabela de
			// símbolos do módulo inteiro (foi exatamente a causa de dois
			// bugs reais encontrados em 2026-09-10: LD A,(rótulo) e DB
			// com rótulo, ambos subestimados em 1 byte no Pass 1).
			estSize, estErr := a.estimateSize(upperMnem, line.operands, line.rawTokens)
			beforeLen := len(a.codeBytes)
			err := a.encodeInstruction(upperMnem, line.label, line.operands, line.rawTokens, line.lineNum)
			if err != nil {
				return nil, fmt.Errorf("linha %d: erro ao codificar '%s': %w", line.lineNum, line.mnemonic, err)
			}
			actualSize := uint16(len(a.codeBytes) - beforeLen)
			if estErr == nil && actualSize != estSize {
				return nil, fmt.Errorf("linha %d: inconsistência interna do assembler em '%s': Pass 1 estimou %d byte(s), Pass 2 emitiu %d byte(s) -- isso desalinharia os rótulos seguintes no módulo (bug no KAJI80, não no código-fonte)",
					line.lineNum, line.mnemonic, estSize, actualSize)
			}
		}
	}

	// Construir o ObjectFile (.MOB)
	obj := mob.NewObjectFile()

	// Segmento CODE (ou dados brutos)
	segIdx := obj.AddSegment(mob.SegmentCode, a.bank, a.codeBytes, 0)

	// Mapear símbolos no .MOB
	symIndexMap := make(map[string]uint16)

	// Símbolos públicos (definidos aqui). Ordenados por nome antes de
	// adicionar ao ObjectFile: a.symbols é um map, cuja ordem de iteração o
	// Go embaralha a cada execução -- sem isso, montar o MESMO fonte duas
	// vezes gera arquivos .MOB byte-a-byte diferentes (tabela de símbolos
	// em ordem diferente), o que é inofensivo para o linker (resolve por
	// nome) mas produz diffs espúrios permanentes em builds versionados no
	// Git a cada recompilação.
	publicNames := make([]string, 0, len(a.publics))
	for name := range a.publics {
		if _, ok := a.symbols[name]; ok {
			publicNames = append(publicNames, name)
		}
	}
	sort.Strings(publicNames)
	for _, name := range publicNames {
		idx := obj.AddSymbol(name, mob.SymbolPublic, mob.SymbolProc, segIdx, a.symbols[name])
		symIndexMap[name] = idx
	}

	// Símbolos externos (importados), mesma correção de determinismo.
	externNames := make([]string, 0, len(a.externs))
	for name := range a.externs {
		externNames = append(externNames, name)
	}
	sort.Strings(externNames)
	for _, name := range externNames {
		idx := obj.AddSymbol(name, mob.SymbolExtern, mob.SymbolProc, 0, 0)
		symIndexMap[name] = idx
	}

	// Se um símbolo local for alvo de relocation e não for PUBLIC nem EXTERN,
	// adicionamos como local ou resolvemos estaticamente.
	for _, r := range a.relocs {
		symIdx, exists := symIndexMap[r.symbolName]
		if !exists {
			// Se o símbolo é um label local do mesmo segmento
			if offset, ok := a.symbols[r.symbolName]; ok {
				// Adiciona como símbolo exportado ou público local para o linker resolver
				symIdx = obj.AddSymbol(r.symbolName, mob.SymbolPublic, mob.SymbolProc, segIdx, offset)
				symIndexMap[r.symbolName] = symIdx
			} else {
				// Símbolo não declarado localmente: assumir EXTERN
				symIdx = obj.AddSymbol(r.symbolName, mob.SymbolExtern, mob.SymbolProc, 0, 0)
				symIndexMap[r.symbolName] = symIdx
			}
		}

		obj.AddRelocation(segIdx, r.offset, symIdx, r.relocType)
	}

	return obj, nil
}

// tokenizeLines agrupa os tokens em linhas lógicas de código
func (a *Assembler) tokenizeLines(source string) ([]parsedLine, error) {
	lexer := NewLexer(source)
	var lineTokens [][]Token
	var currentTokens []Token

	for {
		tok, err := lexer.NextToken()
		if err != nil {
			return nil, err
		}

		if tok.Type == TokenNewline || tok.Type == TokenEOF {
			if len(currentTokens) > 0 {
				lineTokens = append(lineTokens, currentTokens)
				currentTokens = nil
			}
			if tok.Type == TokenEOF {
				break
			}
			continue
		}

		currentTokens = append(currentTokens, tok)
	}

	// Ordem importa: IF/ELSE/ENDIF primeiro (remove ramos mortos antes de
	// qualquer outra coisa enxergar essas linhas), REPT depois (duplica
	// blocos -- se uma macro é chamada dentro de um REPT, REPT precisa
	// duplicar a CHAMADA primeiro, pra cada cópia ganhar seu próprio ID de
	// expansão quando a macro for expandida em cima do resultado), MACRO
	// depois (compartilha o mesmo contador de expansão com REPT, pra toda
	// combinação continuar com ID único), rótulos locais por último --
	// ver preprocessor.go.
	lineTokens, err := a.filterConditionals(lineTokens)
	if err != nil {
		return nil, err
	}
	expCounter := 0
	lineTokens, err = expandRept(lineTokens, &expCounter)
	if err != nil {
		return nil, err
	}
	lineTokens, err = a.expandMacros(lineTokens, make(map[string]*macroDef), &expCounter, 0)
	if err != nil {
		return nil, err
	}
	if err := resolveLocalLabels(lineTokens); err != nil {
		return nil, err
	}

	var lines []parsedLine
	for _, lt := range lineTokens {
		parsed, err := a.parseLine(lt)
		if err != nil {
			return nil, err
		}
		lines = append(lines, parsed)
	}

	return lines, nil
}

func (a *Assembler) parseLine(tokens []Token) (parsedLine, error) {
	pl := parsedLine{
		lineNum:   tokens[0].Line,
		rawTokens: tokens,
	}

	idx := 0
	// Verificar se começa com label (IDENT seguido de COLON, EQU, ou '=' --
	// esta última é a nova forma "Nome = expressão", variável reatribuível,
	// distinta de EQU, que só permite definir uma vez).
	if idx < len(tokens) && tokens[idx].Type == TokenIdentifier {
		if idx+1 < len(tokens) && tokens[idx+1].Type == TokenColon {
			pl.label = tokens[idx].Value
			idx += 2
		} else if idx+1 < len(tokens) && strings.EqualFold(tokens[idx+1].Value, "EQU") {
			pl.label = tokens[idx].Value
			pl.mnemonic = "EQU"
			idx += 2
		} else if idx+1 < len(tokens) && tokens[idx+1].Type == TokenAssign {
			pl.label = tokens[idx].Value
			pl.mnemonic = "ASSIGN"
			idx += 2
		}
	}

	if idx >= len(tokens) {
		return pl, nil
	}

	// Mnemônico ou Diretiva (se ainda não preenchido por EQU)
	if pl.mnemonic == "" {
		if tokens[idx].Type == TokenIdentifier {
			pl.mnemonic = tokens[idx].Value
			idx++
		} else {
			return pl, fmt.Errorf("esperado mnemônico ou diretiva na linha %d, coluna %d", tokens[idx].Line, tokens[idx].Col)
		}
	}

	// Reconstruir operandos separados por vírgula -- só no nível zero de
	// parênteses, pra não quebrar uma chamada de função de múltiplos
	// argumentos dentro de uma expressão (ex.: "POW(2,3)" como operando
	// único de DB, não dois operandos "POW(2" e "3)").
	var currentOp strings.Builder
	parenDepth := 0
	for idx < len(tokens) {
		if tokens[idx].Type == TokenComma && parenDepth == 0 {
			pl.operands = append(pl.operands, strings.TrimSpace(currentOp.String()))
			currentOp.Reset()
			idx++
			continue
		}
		if tokens[idx].Type == TokenLParen {
			parenDepth++
		} else if tokens[idx].Type == TokenRParen {
			parenDepth--
		}
		if tokens[idx].Type == TokenString {
			// O lexer já devolve o conteúdo sem as aspas (correto para DB,
			// que lê os tokens crus diretamente, não os operandos
			// reconstruídos aqui) -- mas para instruções como LD/CP que
			// dependem do texto do operando (parseImm8), perder as aspas
			// torna "'$'" indistinguível de um identificador solto ou de
			// um prefixo hex malformado, e parseImm8 silenciosamente
			// devolve 0. Devolve as aspas simples aqui para que um literal
			// de caractere continue reconhecível como tal.
			currentOp.WriteByte('\'')
			currentOp.WriteString(tokens[idx].Value)
			currentOp.WriteByte('\'')
		} else {
			currentOp.WriteString(tokens[idx].Value)
		}
		idx++
	}
	if currentOp.Len() > 0 {
		pl.operands = append(pl.operands, strings.TrimSpace(currentOp.String()))
	}

	return pl, nil
}

func (a *Assembler) estimateSize(mnem string, ops []string, tokens []Token) (uint16, error) {
	switch mnem {
	case "NOP", "HALT", "DI", "EI", "RET", "EXX", "RLCA", "RRCA", "RLA", "RRA", "CPL", "SCF", "CCF":
		if len(ops) > 0 {
			// RET cc
			return 1, nil
		}
		return 1, nil
	case "NEG":
		return 2, nil
	case "RLC", "RRC", "RL", "RR", "SLA", "SRA", "SRL", "BIT", "RES", "SET":
		return 2, nil
	case "EX":
		return 1, nil
	case "PUSH", "POP":
		if len(ops) > 0 {
			op := strings.ToUpper(ops[0])
			if op == "IX" || op == "IY" {
				return 2, nil
			}
		}
		return 1, nil
	case "CALL", "JP":
		// CALL cc, nn ou CALL nn
		return 3, nil
	case "JR", "DJNZ":
		return 2, nil
	case "IN", "OUT":
		return 2, nil
	case "INC", "DEC":
		if len(ops) > 0 {
			op := strings.ToUpper(ops[0])
			if op == "IX" || op == "IY" {
				return 2, nil
			}
		}
		return 1, nil
	case "ADD":
		if len(ops) == 2 {
			dst := strings.ToUpper(ops[0])
			if dst == "HL" {
				return 1, nil
			}
			if dst == "IX" || dst == "IY" {
				return 2, nil
			}
			src := strings.ToUpper(ops[1])
			if _, ok := reg8Map[src]; ok {
				return 1, nil
			}
			if src == "(HL)" {
				return 1, nil
			}
			if isIndexedOperand(src) {
				return 3, nil // ADD A,(IX+d) / ADD A,(IY+d): DD/FD 86 d
			}
			return 2, nil
		}
		if len(ops) == 1 {
			op := strings.ToUpper(ops[0])
			if _, ok := reg8Map[op]; ok {
				return 1, nil
			}
			if op == "(HL)" {
				return 1, nil
			}
			if isIndexedOperand(op) {
				return 3, nil
			}
			return 2, nil
		}
		return 1, nil
	case "ADC", "SBC":
		if len(ops) == 2 {
			dst := strings.ToUpper(ops[0])
			if dst == "HL" {
				return 2, nil // ED 4A/5A/6A/7A ou ED 42/52/62/72 (2 bytes)
			}
			src := strings.ToUpper(ops[1])
			if _, ok := reg8Map[src]; ok {
				return 1, nil
			}
			if src == "(HL)" {
				return 1, nil
			}
			if isIndexedOperand(src) {
				return 3, nil // ADC/SBC A,(IX+d) / (IY+d): DD/FD 8E/9E d
			}
			return 2, nil
		}
		if len(ops) == 1 {
			op := strings.ToUpper(ops[0])
			if _, ok := reg8Map[op]; ok {
				return 1, nil
			}
			if op == "(HL)" {
				return 1, nil
			}
			if isIndexedOperand(op) {
				return 3, nil
			}
			return 2, nil
		}
		return 1, nil
	case "SUB", "AND", "XOR", "OR", "CP":
		if len(ops) == 2 {
			src := strings.ToUpper(ops[1])
			if _, ok := reg8Map[src]; ok {
				return 1, nil
			}
			if src == "(HL)" {
				return 1, nil
			}
			if isIndexedOperand(src) {
				return 3, nil // SUB/AND/XOR/OR/CP (IX+d) / (IY+d): DD/FD <op> d
			}
			return 2, nil
		}
		if len(ops) == 1 {
			op := strings.ToUpper(ops[0])
			if _, ok := reg8Map[op]; ok {
				return 1, nil
			}
			if op == "(HL)" {
				return 1, nil
			}
			if isIndexedOperand(op) {
				return 3, nil
			}
			return 2, nil
		}
		return 1, nil
	case "LD":
		return a.estimateLdSize(ops)
	case "DB", "DEFB", "BYTE", "DT", "DEFT":
		// Conta por OPERANDO (já separado por vírgula respeitando
		// profundidade de parênteses em parseLine), não por token bruto --
		// um operando de expressão como "X*Y" é 3 tokens (IDENT, STAR,
		// IDENT) mas só 1 byte de dado. parseLine já devolve um literal de
		// string/caractere entre aspas simples (convenção estabelecida
		// pra LD/CP não perderem a distinção entre string e símbolo), então
		// um operando assim conta o número de caracteres reais dentro das
		// aspas; qualquer outro operando é uma expressão de 1 byte.
		var total uint16
		for _, op := range ops {
			if len(op) >= 2 && op[0] == '\'' && op[len(op)-1] == '\'' {
				total += uint16(len(op) - 2)
			} else {
				total++
			}
		}
		return total, nil
	case "DW", "DEFW", "WORD":
		return uint16(len(ops) * 2), nil
	case "DS", "DEFS", "BLKB":
		if len(ops) > 0 {
			var count int
			_, _ = fmt.Sscanf(ops[0], "%d", &count)
			return uint16(count), nil
		}
		return 0, nil
	case "EQU", "ASSIGN":
		return 0, nil
	case "CALLBIOS":
		// LD IX,nn (DD 21 nn nn = 4 bytes) + CALL BIOS_Call (CD nn nn = 3
		// bytes) = 7.
		return 7, nil
	case "CALLDOS":
		// LD C,n (0E nn = 2 bytes) + CALL 0005h (CD 05 00 = 3 bytes) = 5.
		return 5, nil
	case "INCBIN":
		data, err := a.resolveIncbinBytes(ops)
		if err != nil {
			return 0, err
		}
		return uint16(len(data)), nil
	default:
		return 1, nil
	}
}

func (a *Assembler) estimateLdSize(ops []string) (uint16, error) {
	if len(ops) < 2 {
		return 0, fmt.Errorf("LD requer 2 operandos")
	}
	dst := strings.ToUpper(ops[0])
	src := strings.ToUpper(ops[1])

	// LD SP, HL / LD SP, IX / LD SP, IY
	if dst == "SP" {
		if src == "HL" {
			return 1, nil
		}
		if src == "IX" || src == "IY" {
			return 2, nil
		}
	}

	// LD r, (IX+d) / LD (IX+d), r / LD (IX+d), n
	if isIX, isIY, _, ok := parseIndexed(src); ok && (isIX || isIY) {
		if _, okD := reg8Map[dst]; okD {
			return 3, nil
		}
	}
	if isIX, isIY, _, ok := parseIndexed(dst); ok && (isIX || isIY) {
		if _, okS := reg8Map[src]; okS {
			return 3, nil
		}
		return 4, nil // LD (IX+d), n
	}

	// LD r, r'
	if _, okD := reg8Map[dst]; okD {
		if _, okS := reg8Map[src]; okS {
			return 1, nil
		}
		// LD r, (HL)
		if src == "(HL)" {
			return 1, nil
		}
		// LD A, (BC) / LD A, (DE)
		if dst == "A" && (src == "(BC)" || src == "(DE)") {
			return 1, nil
		}
		// LD A, (nn) -- precisa ser checado antes do fallback "LD r, n"
		// abaixo, senão um endereco/rotulo entre parenteses e confundido
		// com um imediato de 8 bits, subestimando o tamanho real (3 bytes)
		// em 1 byte e desalinhando todos os rotulos seguintes no modulo.
		if dst == "A" && strings.HasPrefix(src, "(") {
			return 3, nil
		}
		// LD r, n
		return 2, nil
	}
	// LD (HL), r
	if dst == "(HL)" {
		if _, okS := reg8Map[src]; okS {
			return 1, nil
		}
		// LD (HL), n
		return 2, nil
	}
	// LD rr, nn
	if _, okD := reg16Map[dst]; okD {
		return 3, nil
	}
	// LD IX, nn / LD IY, nn
	if dst == "IX" || dst == "IY" {
		return 4, nil
	}
	// LD (BC), A / LD (DE), A
	if (dst == "(BC)" || dst == "(DE)") && src == "A" {
		return 1, nil
	}
	// LD A, (BC) / LD A, (DE)
	if dst == "A" && (src == "(BC)" || src == "(DE)") {
		return 1, nil
	}
	// LD A, (nn) / LD (nn), A
	if (dst == "A" && strings.HasPrefix(src, "(")) || (strings.HasPrefix(dst, "(") && src == "A") {
		return 3, nil
	}
	// LD HL, (nn) / LD (nn), HL
	if (dst == "HL" && strings.HasPrefix(src, "(")) || (strings.HasPrefix(dst, "(") && src == "HL") {
		return 3, nil
	}

	return 2, nil
}

func (a *Assembler) encodeInstruction(mnem string, label string, ops []string, tokens []Token, lineNum int) error {
	switch mnem {
	case "ASSIGN":
		// "Nome = expressão" -- reavaliada aqui (não só no Pass 1) porque o
		// valor pode ter mudado desde a última vez que esta variável foi
		// referenciada; ver comentário em Assemble() sobre por que
		// a.variables é reconstruída do zero antes do Pass 2.
		if label == "" || len(ops) == 0 {
			return fmt.Errorf("linha %d: atribuição de variável malformada", lineNum)
		}
		val, ok, err := a.EvalExpr(ops[0])
		if err != nil {
			return fmt.Errorf("linha %d: %s = ...: %w", lineNum, label, err)
		}
		if !ok {
			return fmt.Errorf("linha %d: %s = ...: '%s' não é uma expressão numérica válida", lineNum, label, ops[0])
		}
		a.variables[label] = val
		return nil
	case "CALLBIOS":
		// LD IX,rotina / CALL BIOS_Call -- reaproveita a rotina já
		// hardware-testada em lib/src/bios.asm (inter-slot call de
		// verdade: lê o slot em EXPTBL-1 e faz CALSLT) em vez de inlinar
		// a sequência completa toda vez (~15 bytes por uso e zero
		// dependência da MSXLIB, como o asMSX faz) -- escolha registrada
		// no plano aprovado: menor código gerado em troca de uma
		// dependência implícita com um símbolo específico da MSXLIB,
		// "nossa própria sintaxe e recursos" preferindo reaproveitar o
		// que o KIZUNA já construiu e já provou funcionar.
		if len(ops) != 1 {
			return fmt.Errorf("linha %d: CALLBIOS requer 1 operando (a rotina da BIOS a chamar)", lineNum)
		}
		a.emit(0xDD, 0x21) // LD IX, nn
		if err := a.emitAddressOrReloc(ops[0]); err != nil {
			return fmt.Errorf("linha %d: CALLBIOS %s: %w", lineNum, ops[0], err)
		}
		a.externs["BIOS_Call"] = true
		a.emit(0xCD) // CALL nn
		return a.emitAddressOrReloc("BIOS_Call")
	case "CALLDOS":
		// LD C,código / CALL 0005h -- sempre inlined, igual ao asMSX
		// (pequeno demais pra valer a pena uma rotina compartilhada, e
		// BDOS_Call em lib/src/bdos.asm já é só "CALL BDOS_ENTRY / RET",
		// então inlinar direto poupa até a chamada extra).
		if len(ops) != 1 {
			return fmt.Errorf("linha %d: CALLDOS requer 1 operando (o código de função do MSX-DOS)", lineNum)
		}
		a.emit(0x0E, a.parseImm8(ops[0])) // LD C, n
		a.emit(0xCD, 0x05, 0x00)          // CALL 0005h
		return nil
	case "INCBIN":
		data, err := a.resolveIncbinBytes(ops)
		if err != nil {
			return err
		}
		a.emit(data...)
		return nil
	case "NOP":
		a.emit(0x00)
	case "HALT":
		a.emit(0x76)
	case "DI":
		a.emit(0xF3)
	case "EI":
		a.emit(0xFB)
	case "EXX":
		a.emit(0xD9)
	case "RLCA":
		a.emit(0x07)
	case "RRCA":
		a.emit(0x0F)
	case "RLA":
		a.emit(0x17)
	case "RRA":
		a.emit(0x1F)
	case "CPL":
		a.emit(0x2F)
	case "SCF":
		a.emit(0x37)
	case "CCF":
		a.emit(0x3F)
	case "NEG":
		a.emit(0xED, 0x44)
	case "RLC", "RRC", "RL", "RR", "SLA", "SRA", "SRL":
		if len(ops) != 1 {
			return fmt.Errorf("%s requer 1 operando", mnem)
		}
		rStr := strings.ToUpper(ops[0])
		r, ok := reg8Map[rStr]
		if !ok && rStr == "(HL)" {
			r = 6
			ok = true
		}
		if !ok {
			return fmt.Errorf("registrador inválido para %s: %s", mnem, ops[0])
		}
		cbMap := map[string]uint8{
			"RLC": 0x00, "RRC": 0x08, "RL": 0x10, "RR": 0x18,
			"SLA": 0x20, "SRA": 0x28, "SRL": 0x38,
		}
		a.emit(0xCB, cbMap[mnem]|r)
	case "BIT", "RES", "SET":
		if len(ops) != 2 {
			return fmt.Errorf("%s requer 2 operandos (bit, reg)", mnem)
		}
		bitVal := a.parseImm8(ops[0])
		if bitVal > 7 {
			return fmt.Errorf("bit deve estar entre 0 e 7: %d", bitVal)
		}
		rStr := strings.ToUpper(ops[1])
		r, ok := reg8Map[rStr]
		if !ok && rStr == "(HL)" {
			r = 6
			ok = true
		}
		if !ok {
			return fmt.Errorf("registrador inválido para %s: %s", mnem, ops[1])
		}
		baseMap := map[string]uint8{
			"BIT": 0x40, "RES": 0x80, "SET": 0xC0,
		}
		a.emit(0xCB, baseMap[mnem]|(bitVal<<3)|r)
	case "EX":
		if len(ops) == 2 && strings.EqualFold(ops[0], "DE") && strings.EqualFold(ops[1], "HL") {
			a.emit(0xEB)
		} else if len(ops) == 2 && strings.EqualFold(ops[0], "AF") && strings.EqualFold(ops[1], "AF'") {
			a.emit(0x08)
		} else {
			return fmt.Errorf("combinação de EX não suportada: %v", ops)
		}
	case "RET":
		if len(ops) == 0 {
			a.emit(0xC9)
		} else {
			cc, ok := condMap[strings.ToUpper(ops[0])]
			if !ok {
				return fmt.Errorf("condição inválida para RET: %s", ops[0])
			}
			a.emit(0xC0 | (cc << 3))
		}
	case "CALL":
		if len(ops) == 1 {
			a.emit(0xCD)
			return a.emitAddressOrReloc(ops[0])
		} else if len(ops) == 2 {
			cc, ok := condMap[strings.ToUpper(ops[0])]
			if !ok {
				return fmt.Errorf("condição inválida para CALL: %s", ops[0])
			}
			a.emit(0xC4 | (cc << 3))
			return a.emitAddressOrReloc(ops[1])
		}
	case "JP":
		if len(ops) == 1 {
			op := strings.ToUpper(ops[0])
			if op == "(HL)" {
				a.emit(0xE9)
			} else if op == "(IX)" {
				a.emit(0xDD, 0xE9)
			} else if op == "(IY)" {
				a.emit(0xFD, 0xE9)
			} else {
				a.emit(0xC3)
				return a.emitAddressOrReloc(ops[0])
			}
		} else if len(ops) == 2 {
			cc, ok := condMap[strings.ToUpper(ops[0])]
			if !ok {
				return fmt.Errorf("condição inválida para JP: %s", ops[0])
			}
			a.emit(0xC2 | (cc << 3))
			return a.emitAddressOrReloc(ops[1])
		}
	case "JR":
		if len(ops) == 1 {
			a.emit(0x18)
			return a.emitRelativeOrReloc(ops[0])
		} else if len(ops) == 2 {
			cc, ok := condMap[strings.ToUpper(ops[0])]
			if !ok || cc > 3 { // Apenas NZ(0), Z(1), NC(2), C(3)
				return fmt.Errorf("condição inválida para JR: %s", ops[0])
			}
			a.emit(0x20 | (cc << 3))
			return a.emitRelativeOrReloc(ops[1])
		}
	case "DJNZ":
		if len(ops) == 1 {
			a.emit(0x10)
			return a.emitRelativeOrReloc(ops[0])
		}
	case "PUSH":
		if len(ops) != 1 {
			return fmt.Errorf("PUSH requer 1 registrador")
		}
		reg := strings.ToUpper(ops[0])
		if reg == "IX" {
			a.emit(0xDD, 0xE5)
		} else if reg == "IY" {
			a.emit(0xFD, 0xE5)
		} else if p, ok := reg16PushPopMap[reg]; ok {
			a.emit(0xC5 | (p << 4))
		} else {
			return fmt.Errorf("registrador inválido para PUSH: %s", reg)
		}
	case "POP":
		if len(ops) != 1 {
			return fmt.Errorf("POP requer 1 registrador")
		}
		reg := strings.ToUpper(ops[0])
		if reg == "IX" {
			a.emit(0xDD, 0xE1)
		} else if reg == "IY" {
			a.emit(0xFD, 0xE1)
		} else if p, ok := reg16PushPopMap[reg]; ok {
			a.emit(0xC1 | (p << 4))
		} else {
			return fmt.Errorf("registrador inválido para POP: %s", reg)
		}
	case "IN":
		if len(ops) == 2 && strings.EqualFold(ops[0], "A") {
			// IN A, (n)
			port := strings.Trim(ops[1], "()")
			val := a.parseImm8(port)
			a.emit(0xDB, val)
		} else {
			return fmt.Errorf("forma de IN não suportada: %v", ops)
		}
	case "OUT":
		if len(ops) == 2 && strings.EqualFold(ops[1], "A") {
			// OUT (n), A
			port := strings.Trim(ops[0], "()")
			val := a.parseImm8(port)
			a.emit(0xD3, val)
		} else {
			return fmt.Errorf("forma de OUT não suportada: %v", ops)
		}
	case "INC":
		if len(ops) != 1 {
			return fmt.Errorf("INC requer 1 operando")
		}
		op := strings.ToUpper(ops[0])
		if r, ok := reg8Map[op]; ok {
			a.emit(0x04 | (r << 3))
		} else if op == "(HL)" {
			a.emit(0x34)
		} else if p, ok := reg16Map[op]; ok {
			a.emit(0x03 | (p << 4))
		} else if op == "IX" {
			a.emit(0xDD, 0x23)
		} else if op == "IY" {
			a.emit(0xFD, 0x23)
		} else {
			return fmt.Errorf("operando inválido para INC: %s", op)
		}
	case "DEC":
		if len(ops) != 1 {
			return fmt.Errorf("DEC requer 1 operando")
		}
		op := strings.ToUpper(ops[0])
		if r, ok := reg8Map[op]; ok {
			a.emit(0x05 | (r << 3))
		} else if op == "(HL)" {
			a.emit(0x35)
		} else if p, ok := reg16Map[op]; ok {
			a.emit(0x0B | (p << 4))
		} else if op == "IX" {
			a.emit(0xDD, 0x2B)
		} else if op == "IY" {
			a.emit(0xFD, 0x2B)
		} else {
			return fmt.Errorf("operando inválido para DEC: %s", op)
		}
	case "ADD":
		if len(ops) == 2 {
			dst := strings.ToUpper(ops[0])
			if dst == "HL" {
				p, ok := reg16Map[strings.ToUpper(ops[1])]
				if !ok {
					return fmt.Errorf("registrador inválido para ADD HL: %s", ops[1])
				}
				a.emit(0x09 | (p << 4))
				return nil
			}
			if dst == "IX" {
				p, ok := reg16Map[strings.ToUpper(ops[1])]
				if !ok {
					return fmt.Errorf("registrador inválido para ADD IX: %s", ops[1])
				}
				a.emit(0xDD, 0x09|(p<<4))
				return nil
			}
			if dst == "IY" {
				p, ok := reg16Map[strings.ToUpper(ops[1])]
				if !ok {
					return fmt.Errorf("registrador inválido para ADD IY: %s", ops[1])
				}
				a.emit(0xFD, 0x09|(p<<4))
				return nil
			}
		}
		return a.encodeAlu8(mnem, ops)
	case "ADC":
		if len(ops) == 2 && strings.EqualFold(ops[0], "HL") {
			p, ok := reg16Map[strings.ToUpper(ops[1])]
			if !ok {
				return fmt.Errorf("registrador inválido para ADC HL: %s", ops[1])
			}
			a.emit(0xED, 0x4A|(p<<4))
			return nil
		}
		return a.encodeAlu8(mnem, ops)
	case "SBC":
		if len(ops) == 2 && strings.EqualFold(ops[0], "HL") {
			p, ok := reg16Map[strings.ToUpper(ops[1])]
			if !ok {
				return fmt.Errorf("registrador inválido para SBC HL: %s", ops[1])
			}
			a.emit(0xED, 0x42|(p<<4))
			return nil
		}
		return a.encodeAlu8(mnem, ops)
	case "SUB", "AND", "XOR", "OR", "CP":
		return a.encodeAlu8(mnem, ops)
	case "LD":
		return a.encodeLd(ops)
	case "DB", "DEFB", "BYTE", "DT", "DEFT":
		// Espelha o Pass 1 (estimateSize): itera por OPERANDO (já separado
		// por vírgula respeitando parênteses), não por token bruto -- uma
		// expressão como "X*Y" precisa ser avaliada como UM valor, não
		// emitida token a token.
		for _, op := range ops {
			if len(op) >= 2 && op[0] == '\'' && op[len(op)-1] == '\'' {
				content := op[1 : len(op)-1]
				for i := 0; i < len(content); i++ {
					a.emit(content[i])
				}
				continue
			}
			val, ok, err := a.EvalExpr(op)
			if err != nil {
				return fmt.Errorf("linha %d: %s %s: %w", lineNum, mnem, op, err)
			}
			if ok {
				a.emit(uint8(int64(val)))
				continue
			}
			a.emit(a.parseImm8(op))
		}
	case "DW", "DEFW", "WORD":
		for _, op := range ops {
			val, ok, err := a.EvalExpr(op)
			if err != nil {
				return fmt.Errorf("linha %d: %s %s: %w", lineNum, mnem, op, err)
			}
			if ok {
				iv := int64(val)
				a.emit(uint8(iv&0xFF), uint8((iv>>8)&0xFF))
				continue
			}
			if err := a.emitAddressOrReloc(op); err != nil {
				return err
			}
		}
	case "DS", "DEFS", "BLKB":
		var count int
		if len(ops) > 0 {
			_, _ = fmt.Sscanf(ops[0], "%d", &count)
			for i := 0; i < count; i++ {
				a.emit(0x00)
			}
		}
	default:
		return fmt.Errorf("instrução desconhecida '%s'", mnem)
	}
	return nil
}

// isIndexedOperand reporta se o operando é uma forma indexada (IX+d)/(IY+d).
// Usado pelos estimadores de tamanho e codificadores das operações ALU de 8
// bits (ADD, ADC, SUB, SBC, AND, XOR, OR, CP), que antes desta correção
// caíam silenciosamente no ramo de "imediato de 8 bits" para esse operando:
// parseImm8("(IX+8)") não reconhece a sintaxe e devolve 0, então "CP (IX+8)"
// virava "CP 0" sem erro nenhum -- e o Pass 1 estimava o mesmo tamanho (2
// bytes) que o Pass 2 emitia, então a verificação de consistência interna
// do Assemble() não detectava a divergência semântica (só detecta diferença
// de TAMANHO, não de significado). Foi a causa raiz real do bug gráfico de
// SCREEN 2 em VDP_Line/VDP_BoxFill (lib/src/vdp.asm), que são os únicos
// usos dessa forma no projeto inteiro.
func isIndexedOperand(op string) bool {
	isIX, isIY, _, ok := parseIndexed(op)
	return ok && (isIX || isIY)
}

func (a *Assembler) encodeAlu8(mnem string, ops []string) error {
	opCode := aluMap[mnem]
	target := ops[0]
	if len(ops) == 2 {
		target = ops[1]
	}
	upperTarget := strings.ToUpper(target)
	if r, ok := reg8Map[upperTarget]; ok {
		a.emit(0x80 | (opCode << 3) | r)
	} else if upperTarget == "(HL)" {
		a.emit(0x86 | (opCode << 3))
	} else if isIX, isIY, disp, ok := parseIndexed(upperTarget); ok && (isIX || isIY) {
		// ADD/ADC/SUB/SBC/AND/XOR/OR/CP A,(IX+d) ou (IY+d): DD/FD <opcode> d
		prefix := uint8(0xDD)
		if isIY {
			prefix = 0xFD
		}
		a.emit(prefix, 0x86|(opCode<<3), uint8(disp))
	} else if strings.HasPrefix(target, "(") && strings.HasSuffix(target, ")") {
		// Forma de memória não reconhecida (nem (HL) nem (IX+d)/(IY+d)) --
		// sem este check, caía no fallback de imediato abaixo, que não tem
		// como reportar erro e silenciosamente virava "ADD/CP/etc A, 0"
		// (mesma classe de bug já achada 2x nesta sessão: Label+N e
		// "LD B,(nn)"). Ex.: "CP (Algum_Endereco)" -- não existe forma ALU
		// indireta pra endereço absoluto no Z80, só via (HL)/(IX+d)/(IY+d).
		return fmt.Errorf("forma de %s não suportada: %s -- ALU indireto só existe via (HL) ou (IX+d)/(IY+d), nunca endereço absoluto", mnem, target)
	} else {
		// Imediato de 8 bits
		val := a.parseImm8(target)
		a.emit(0xC6|(opCode<<3), val)
	}
	return nil
}

func (a *Assembler) encodeLd(ops []string) error {
	if len(ops) < 2 {
		return fmt.Errorf("LD requer 2 operandos")
	}
	dst := strings.ToUpper(ops[0])
	src := strings.ToUpper(ops[1])

	// LD SP, HL / LD SP, IX / LD SP, IY
	if dst == "SP" {
		if src == "HL" {
			a.emit(0xF9)
			return nil
		}
		if src == "IX" {
			a.emit(0xDD, 0xF9)
			return nil
		}
		if src == "IY" {
			a.emit(0xFD, 0xF9)
			return nil
		}
	}

	// LD r, (IX+d) / LD r, (IY+d)
	if isIX, isIY, disp, ok := parseIndexed(src); ok && (isIX || isIY) {
		if d, okD := reg8Map[dst]; okD {
			prefix := uint8(0xDD)
			if isIY {
				prefix = 0xFD
			}
			a.emit(prefix, 0x46|(d<<3), uint8(disp))
			return nil
		}
	}

	// LD (IX+d), r / LD (IY+d), r / LD (IX+d), n
	if isIX, isIY, disp, ok := parseIndexed(dst); ok && (isIX || isIY) {
		prefix := uint8(0xDD)
		if isIY {
			prefix = 0xFD
		}
		if s, okS := reg8Map[src]; okS {
			a.emit(prefix, 0x70|s, uint8(disp))
			return nil
		}
		// LD (IX+d), n
		val := a.parseImm8(ops[1])
		a.emit(prefix, 0x36, uint8(disp), val)
		return nil
	}

	// LD r, r'
	if d, okD := reg8Map[dst]; okD {
		if s, okS := reg8Map[src]; okS {
			a.emit(0x40 | (d << 3) | s)
			return nil
		}
		if src == "(HL)" {
			a.emit(0x46 | (d << 3))
			return nil
		}
		if dst == "A" && src == "(BC)" {
			a.emit(0x0A)
			return nil
		}
		if dst == "A" && src == "(DE)" {
			a.emit(0x1A)
			return nil
		}
		if dst == "A" && strings.HasPrefix(src, "(") && strings.HasSuffix(src, ")") {
			addr := strings.Trim(ops[1], "()")
			a.emit(0x3A)
			return a.emitAddressOrReloc(addr)
		}
		// O Z80 de verdade só tem endereçamento absoluto de 16 bits pra UM
		// registrador de 8 bits através de A ("LD A,(nn)", opcode 0x3A) --
		// "LD B,(nn)"/"LD C,(nn)"/etc. simplesmente não existem. Sem este
		// check, cai no fallback de imediato abaixo (parseImm8, que não
		// tem como reportar erro) e monta em silêncio como "LD r, 0" --
		// bug real encontrado em lib/src/float.asm (Float_Cmp32): 3 linhas
		// "LD B,(Float_MantHi)" etc viravam "LD B,0", quebrando comparações
		// sempre que o fluxo alcançava esses trechos.
		if strings.HasPrefix(src, "(") && strings.HasSuffix(src, ")") {
			return fmt.Errorf("forma de LD não suportada: LD %s, %s -- só LD A,(nn) tem endereçamento absoluto de 16 bits pra um registrador de 8 bits; carregue em A e mova com LD %s,A", ops[0], ops[1], ops[0])
		}
		// LD r, n
		val := a.parseImm8(ops[1])
		a.emit(0x06|(d<<3), val)
		return nil
	}

	// LD (HL), r / n
	if dst == "(HL)" {
		if s, okS := reg8Map[src]; okS {
			a.emit(0x70 | s)
			return nil
		}
		val := a.parseImm8(ops[1])
		a.emit(0x36, val)
		return nil
	}

	// LD (BC), A / LD (DE), A
	if dst == "(BC)" && src == "A" {
		a.emit(0x02)
		return nil
	}
	if dst == "(DE)" && src == "A" {
		a.emit(0x12)
		return nil
	}

	// LD (nn), A
	if strings.HasPrefix(dst, "(") && strings.HasSuffix(dst, ")") && src == "A" {
		addr := strings.Trim(ops[0], "()")
		a.emit(0x32)
		return a.emitAddressOrReloc(addr)
	}

	// LD HL, (nn)
	if dst == "HL" && strings.HasPrefix(src, "(") && strings.HasSuffix(src, ")") {
		addr := strings.Trim(ops[1], "()")
		a.emit(0x2A)
		return a.emitAddressOrReloc(addr)
	}

	// LD (nn), HL
	if strings.HasPrefix(dst, "(") && strings.HasSuffix(dst, ")") && src == "HL" {
		addr := strings.Trim(ops[0], "()")
		a.emit(0x22)
		return a.emitAddressOrReloc(addr)
	}

	// LD rr, nn
	if p, okP := reg16Map[dst]; okP {
		a.emit(0x01 | (p << 4))
		return a.emitAddressOrReloc(ops[1])
	}

	// LD IX, nn / LD IY, nn
	if dst == "IX" {
		a.emit(0xDD, 0x21)
		return a.emitAddressOrReloc(ops[1])
	}
	if dst == "IY" {
		a.emit(0xFD, 0x21)
		return a.emitAddressOrReloc(ops[1])
	}

	return fmt.Errorf("forma de LD não suportada: LD %s, %s", ops[0], ops[1])
}

func parseIndexed(op string) (isIX bool, isIY bool, disp int8, ok bool) {
	clean := strings.ToUpper(strings.ReplaceAll(op, " ", ""))
	if !strings.HasPrefix(clean, "(") || !strings.HasSuffix(clean, ")") {
		return false, false, 0, false
	}
	inner := clean[1 : len(clean)-1]
	var prefix string
	if strings.HasPrefix(inner, "IX") {
		isIX = true
		prefix = "IX"
	} else if strings.HasPrefix(inner, "IY") {
		isIY = true
		prefix = "IY"
	} else {
		return false, false, 0, false
	}

	rest := inner[len(prefix):]
	if rest == "" {
		return isIX, isIY, 0, true
	}

	sign := 1
	if rest[0] == '+' {
		rest = rest[1:]
	} else if rest[0] == '-' {
		sign = -1
		rest = rest[1:]
	} else {
		return false, false, 0, false
	}

	var val int64
	if strings.HasSuffix(rest, "H") {
		v, err := parseHex(rest[:len(rest)-1])
		if err != nil {
			return false, false, 0, false
		}
		val = v
	} else if strings.HasPrefix(rest, "$") || strings.HasPrefix(rest, "#") {
		v, err := parseHex(rest[1:])
		if err != nil {
			return false, false, 0, false
		}
		val = v
	} else {
		v, err := strconv.ParseInt(rest, 10, 32)
		if err != nil {
			return false, false, 0, false
		}
		val = v
	}

	return isIX, isIY, int8(int64(sign) * val), true
}

func (a *Assembler) emit(bytes ...uint8) {
	a.codeBytes = append(a.codeBytes, bytes...)
}

func (a *Assembler) parseImm8(s string) uint8 {
	s = strings.TrimSpace(s)
	if val, ok := a.constants[s]; ok {
		return uint8(val & 0xFF)
	}
	// Rótulo pré-definido (BIOS/BDOS/BIOSVARS) -- ex.: "LD C, F_OPEN" --
	// só como último recurso, código do usuário (rótulo do arquivo ou
	// EXTERN explícito) sempre tem prioridade. Trunca pro byte baixo, já
	// que a maioria dos usos de 8 bits aqui é código de função BDOS
	// (cabe num byte de verdade); um endereço BIOS de 16 bits usado aqui
	// por engano só trunca, mesmo comportamento silencioso que qualquer
	// outra constante grande demais já tinha antes desta leva.
	if _, isLocal := a.symbols[s]; !isLocal {
		if _, isExtern := a.externs[s]; !isExtern {
			if v, ok := lookupPredefined(s); ok {
				return uint8(v & 0xFF)
			}
		}
	}
	// Literal de caractere entre aspas simples: 'A', '$', etc. -- sem isso,
	// caía no Sscanf genérico abaixo, que falha silenciosamente pra essa
	// sintaxe e devolve 0 (mesma classe de bug já vista com operandos
	// indexados não reconhecidos, ver estimateSize/encodeAlu8).
	if len(s) == 3 && s[0] == '\'' && s[2] == '\'' {
		return s[1]
	}
	// Tratar hex como 0x10, 10h, $10, #10
	if strings.HasPrefix(s, "$") || strings.HasPrefix(s, "#") {
		var val uint8
		_, _ = fmt.Sscanf(s[1:], "%x", &val)
		return val
	}
	if strings.HasSuffix(strings.ToLower(s), "h") {
		var val uint8
		_, _ = fmt.Sscanf(s[:len(s)-1], "%x", &val)
		return val
	}
	var val int
	_, _ = fmt.Sscanf(s, "%v", &val)
	return uint8(val)
}

// isValidSymbolName reporta se s tem a forma de um identificador de verdade
// (letra ou '_' seguido de letras/dígitos/'_') -- usado como último filtro
// antes de emitAddressOrReloc/emitRelativeOrReloc tratarem algo como um
// símbolo a resolver por relocation. Sem isso, QUALQUER string que não
// batesse com constante EQU nem número literal virava um "símbolo" válido
// em silêncio, mesmo coisas como "(Float_DestAddr)" (operando de memória
// mal-formado) ou "Float_UnpA+1" (aritmética de label, que o KAJI80 não
// suporta) -- gerando relocations fantasma que só falhavam (ou pior,
// resolviam por acidente) na hora da linkagem, bem longe da linha real do
// bug. Dois bugs reais desta classe já foram achados e corrigidos em
// pontos individuais (Label+N em lib/src/float.asm, "LD B,(nn)" em
// Float_Cmp32) antes desta validação central existir.
func isValidSymbolName(s string) bool {
	if s == "" {
		return false
	}
	for i, r := range s {
		isLetter := (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || r == '_'
		isDigit := r >= '0' && r <= '9'
		if i == 0 {
			if !isLetter {
				return false
			}
		} else if !isLetter && !isDigit {
			return false
		}
	}
	return true
}

func (a *Assembler) emitAddressOrReloc(symbolOrAddr string) error {
	currOffset := uint16(len(a.codeBytes))
	symbolOrAddr = strings.TrimSpace(symbolOrAddr)

	// Se for uma constante definida por EQU
	if cVal, ok := a.constants[symbolOrAddr]; ok {
		lo := uint8(cVal & 0xFF)
		hi := uint8((cVal >> 8) & 0xFF)
		a.emit(lo, hi)
		return nil
	}

	// Se for número literal (ex: 0x1234, 100, $C000)
	var numVal int64
	var err error
	if strings.HasPrefix(symbolOrAddr, "$") || strings.HasPrefix(symbolOrAddr, "#") {
		numVal, err = parseHex(symbolOrAddr[1:])
	} else if strings.HasSuffix(strings.ToLower(symbolOrAddr), "h") {
		numVal, err = parseHex(symbolOrAddr[:len(symbolOrAddr)-1])
	} else if strings.HasPrefix(strings.ToLower(symbolOrAddr), "0x") {
		numVal, err = parseHex(symbolOrAddr[2:])
	} else {
		var n int
		n, err = fmt.Sscanf(symbolOrAddr, "%d", &numVal)
		if n == 0 {
			err = fmt.Errorf("não é número")
		}
	}

	if err == nil {
		// É número literal
		lo := uint8(numVal & 0xFF)
		hi := uint8((numVal >> 8) & 0xFF)
		a.emit(lo, hi)
		return nil
	}

	if !isValidSymbolName(symbolOrAddr) {
		return fmt.Errorf("operando de endereço inválido: '%s' não é um número, constante EQU nem nome de símbolo válido", symbolOrAddr)
	}

	// Rótulo pré-definido (BIOS/BDOS/BIOSVARS) -- só como ÚLTIMO recurso,
	// depois de checar que o nome não é um rótulo de verdade definido
	// neste arquivo (a.symbols, já populado pelo Pass 1 nesse ponto) nem
	// um EXTERN explícito -- código do usuário sempre tem prioridade sobre
	// o valor pré-definido, nunca o contrário.
	if _, isLocal := a.symbols[symbolOrAddr]; !isLocal {
		if _, isExtern := a.externs[symbolOrAddr]; !isExtern {
			if v, ok := lookupPredefined(symbolOrAddr); ok {
				a.emit(uint8(v&0xFF), uint8((v>>8)&0xFF))
				return nil
			}
		}
	}

	// É um símbolo/label (precisa de relocation ABS16 no .MOB)
	a.relocs = append(a.relocs, tempReloc{
		offset:     currOffset,
		symbolName: symbolOrAddr,
		relocType:  mob.RelocAbs16,
	})

	// Espaço reservado para o endereço de 16 bits
	a.emit(0x00, 0x00)
	return nil
}

func (a *Assembler) emitRelativeOrReloc(symbolOrTarget string) error {
	currOffset := uint16(len(a.codeBytes))
	symbolOrTarget = strings.TrimSpace(symbolOrTarget)

	// Se o label já for conhecido localmente
	if targetOffset, ok := a.symbols[symbolOrTarget]; ok {
		// O salto relativo no Z80 é medido a partir de PC + 2
		nextPC := int(currOffset) + 1 // +1 para o byte do deslocamento
		disp := int(targetOffset) - nextPC
		if disp < -128 || disp > 127 {
			// Salto muito longo
			a.relocs = append(a.relocs, tempReloc{
				offset:     currOffset,
				symbolName: symbolOrTarget,
				relocType:  mob.RelocRel8,
			})
			a.emit(0x00)
			return nil
		}
		a.emit(uint8(int8(disp)))
		return nil
	}

	if !isValidSymbolName(symbolOrTarget) {
		return fmt.Errorf("alvo de salto inválido: '%s' não é um nome de símbolo válido", symbolOrTarget)
	}

	// Símbolo não resolvido no mesmo escopo local -> Relocation REL8
	a.relocs = append(a.relocs, tempReloc{
		offset:     currOffset,
		symbolName: symbolOrTarget,
		relocType:  mob.RelocRel8,
	})
	a.emit(0x00)
	return nil
}

func parseHex(s string) (int64, error) {
	var val int64
	n, err := fmt.Sscanf(s, "%x", &val)
	if n == 0 || err != nil {
		return 0, fmt.Errorf("hex inválido")
	}
	return val, nil
}

// =============================================================================
// INCBIN "arquivo", SKIP=x, SIZE=y
//
// Sintaxe: caminho do arquivo entre aspas (mesma convenção de string do
// resto do assembler), seguido opcionalmente de "SKIP=n" e/ou "SIZE=n"
// separados por vírgula, em qualquer ordem. Usa "=" (já suportado desde a
// Fase 1, pra "Nome = expressão") em vez de "SKIP X"/"SIZE Y" com espaço
// -- parseLine concatena tokens sem espaço entre eles ao reconstruir um
// operando, então "SKIP 10" viraria o texto "SKIP10" (ambíguo de separar
// de volta); "SKIP=10" tokeniza como IDENT+ASSIGN+NUMBER e reconstrói sem
// ambiguidade nenhuma.
// =============================================================================

// parseIncbinOperands extrai o caminho do arquivo e os parâmetros
// opcionais SKIP=/SIZE= dos operandos já separados por vírgula.
func parseIncbinOperands(ops []string) (path string, skip int, size int, hasSize bool, err error) {
	if len(ops) == 0 {
		return "", 0, 0, false, fmt.Errorf("INCBIN requer o caminho de um arquivo entre aspas, ex.: INCBIN \"sprites.bin\"")
	}
	raw := ops[0]
	if len(raw) < 2 || raw[0] != '\'' || raw[len(raw)-1] != '\'' {
		return "", 0, 0, false, fmt.Errorf("INCBIN requer o caminho do arquivo entre aspas, ex.: INCBIN \"sprites.bin\"")
	}
	path = raw[1 : len(raw)-1]

	for _, op := range ops[1:] {
		upper := strings.ToUpper(op)
		switch {
		case strings.HasPrefix(upper, "SKIP="):
			v, convErr := strconv.Atoi(op[len("SKIP="):])
			if convErr != nil {
				return "", 0, 0, false, fmt.Errorf("INCBIN: SKIP inválido: %q", op)
			}
			skip = v
		case strings.HasPrefix(upper, "SIZE="):
			v, convErr := strconv.Atoi(op[len("SIZE="):])
			if convErr != nil {
				return "", 0, 0, false, fmt.Errorf("INCBIN: SIZE inválido: %q", op)
			}
			size = v
			hasSize = true
		default:
			return "", 0, 0, false, fmt.Errorf("INCBIN: operando desconhecido %q (esperado SKIP=n ou SIZE=n)", op)
		}
	}
	return path, skip, size, hasSize, nil
}

// readIncbinFile lê (e põe em cache, chaveado pelo caminho já resolvido)
// o conteúdo bruto de um arquivo pra INCBIN -- o mesmo arquivo pode ser
// lido até 3 vezes pela mesma linha (Pass 1, checagem de consistência do
// Pass 2, e o Pass 2 de verdade), o cache garante que todas enxergam
// exatamente os mesmos bytes, mesmo que o arquivo mude no disco no meio
// do processo (situação de borda, mas o cache resolve de graça).
func (a *Assembler) readIncbinFile(path string) ([]byte, error) {
	resolved := path
	if !filepath.IsAbs(resolved) && a.baseDir != "" {
		resolved = filepath.Join(a.baseDir, path)
	}
	if a.incbinCache == nil {
		a.incbinCache = make(map[string][]byte)
	}
	if data, ok := a.incbinCache[resolved]; ok {
		return data, nil
	}
	data, err := os.ReadFile(resolved)
	if err != nil {
		return nil, fmt.Errorf("INCBIN: erro ao ler arquivo '%s': %w", path, err)
	}
	a.incbinCache[resolved] = data
	return data, nil
}

// resolveIncbinBytes devolve o slice de bytes final (já com SKIP/SIZE
// aplicados) pra uma linha INCBIN.
func (a *Assembler) resolveIncbinBytes(ops []string) ([]byte, error) {
	path, skip, size, hasSize, err := parseIncbinOperands(ops)
	if err != nil {
		return nil, err
	}
	data, err := a.readIncbinFile(path)
	if err != nil {
		return nil, err
	}
	if skip < 0 || skip > len(data) {
		return nil, fmt.Errorf("INCBIN \"%s\": SKIP=%d além do tamanho do arquivo (%d bytes)", path, skip, len(data))
	}
	data = data[skip:]
	if hasSize {
		if size < 0 || size > len(data) {
			return nil, fmt.Errorf("INCBIN \"%s\": SIZE=%d além do que resta do arquivo depois do SKIP (%d bytes disponíveis)", path, size, len(data))
		}
		data = data[:size]
	}
	return data, nil
}
