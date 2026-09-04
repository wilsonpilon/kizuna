package kaji80

import (
	"fmt"
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
	symbols    map[string]uint16 // label -> offset no segmento
	codeBytes  []byte
	relocs     []tempReloc
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
				val := a.parseConstant(line.operands[0])
				a.constants[line.label] = val
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

	// Passo 2: Codificar instruções e gerar relocações
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
			err := a.encodeInstruction(upperMnem, line.operands, line.rawTokens, line.lineNum)
			if err != nil {
				return nil, fmt.Errorf("linha %d: erro ao codificar '%s': %w", line.lineNum, line.mnemonic, err)
			}
		}
	}

	// Construir o ObjectFile (.MOB)
	obj := mob.NewObjectFile()

	// Segmento CODE (ou dados brutos)
	segIdx := obj.AddSegment(mob.SegmentCode, a.bank, a.codeBytes, 0)

	// Mapear símbolos no .MOB
	symIndexMap := make(map[string]uint16)

	// Símbolos públicos (definidos aqui)
	for name, offset := range a.symbols {
		if a.publics[name] {
			idx := obj.AddSymbol(name, mob.SymbolPublic, mob.SymbolProc, segIdx, offset)
			symIndexMap[name] = idx
		}
	}

	// Símbolos externos (importados)
	for name := range a.externs {
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
	var lines []parsedLine
	var currentTokens []Token

	for {
		tok, err := lexer.NextToken()
		if err != nil {
			return nil, err
		}

		if tok.Type == TokenNewline || tok.Type == TokenEOF {
			if len(currentTokens) > 0 {
				parsed, err := a.parseLine(currentTokens)
				if err != nil {
					return nil, err
				}
				lines = append(lines, parsed)
				currentTokens = nil
			}
			if tok.Type == TokenEOF {
				break
			}
			continue
		}

		currentTokens = append(currentTokens, tok)
	}

	return lines, nil
}

func (a *Assembler) parseLine(tokens []Token) (parsedLine, error) {
	pl := parsedLine{
		lineNum:   tokens[0].Line,
		rawTokens: tokens,
	}

	idx := 0
	// Verificar se começa com label (IDENT seguido de COLON ou se seguido de EQU)
	if idx < len(tokens) && tokens[idx].Type == TokenIdentifier {
		if idx+1 < len(tokens) && tokens[idx+1].Type == TokenColon {
			pl.label = tokens[idx].Value
			idx += 2
		} else if idx+1 < len(tokens) && strings.EqualFold(tokens[idx+1].Value, "EQU") {
			pl.label = tokens[idx].Value
			pl.mnemonic = "EQU"
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

	// Reconstruir operandos separados por vírgula
	var currentOp strings.Builder
	for idx < len(tokens) {
		if tokens[idx].Type == TokenComma {
			pl.operands = append(pl.operands, strings.TrimSpace(currentOp.String()))
			currentOp.Reset()
			idx++
			continue
		}
		currentOp.WriteString(tokens[idx].Value)
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
			return 2, nil
		}
		return 1, nil
	case "LD":
		return a.estimateLdSize(ops)
	case "DB", "DEFB", "BYTE":
		var total uint16
		for _, tok := range tokens[1:] {
			if tok.Type == TokenString {
				total += uint16(len(tok.Value))
			} else if tok.Type == TokenNumber || tok.Type == TokenIdentifier {
				total++
			}
		}
		if total == 0 {
			total = uint16(len(ops))
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
	case "EQU":
		return 0, nil
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

	// LD SP, HL
	if dst == "SP" && src == "HL" {
		return 1, nil
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

func (a *Assembler) encodeInstruction(mnem string, ops []string, tokens []Token, lineNum int) error {
	switch mnem {
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
			a.emitAddressOrReloc(ops[0])
		} else if len(ops) == 2 {
			cc, ok := condMap[strings.ToUpper(ops[0])]
			if !ok {
				return fmt.Errorf("condição inválida para CALL: %s", ops[0])
			}
			a.emit(0xC4 | (cc << 3))
			a.emitAddressOrReloc(ops[1])
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
				a.emitAddressOrReloc(ops[0])
			}
		} else if len(ops) == 2 {
			cc, ok := condMap[strings.ToUpper(ops[0])]
			if !ok {
				return fmt.Errorf("condição inválida para JP: %s", ops[0])
			}
			a.emit(0xC2 | (cc << 3))
			a.emitAddressOrReloc(ops[1])
		}
	case "JR":
		if len(ops) == 1 {
			a.emit(0x18)
			a.emitRelativeOrReloc(ops[0])
		} else if len(ops) == 2 {
			cc, ok := condMap[strings.ToUpper(ops[0])]
			if !ok || cc > 3 { // Apenas NZ(0), Z(1), NC(2), C(3)
				return fmt.Errorf("condição inválida para JR: %s", ops[0])
			}
			a.emit(0x20 | (cc << 3))
			a.emitRelativeOrReloc(ops[1])
		}
	case "DJNZ":
		if len(ops) == 1 {
			a.emit(0x10)
			a.emitRelativeOrReloc(ops[0])
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
	case "DB", "DEFB", "BYTE":
		for _, tok := range tokens[1:] {
			if tok.Type == TokenString {
				for i := 0; i < len(tok.Value); i++ {
					a.emit(tok.Value[i])
				}
			} else if tok.Type == TokenNumber {
				a.emit(uint8(tok.Number))
			} else if tok.Type == TokenIdentifier {
				a.emit(a.parseImm8(tok.Value))
			}
		}
	case "DW", "DEFW", "WORD":
		for _, op := range ops {
			a.emitAddressOrReloc(op)
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

	// LD SP, HL
	if dst == "SP" && src == "HL" {
		a.emit(0xF9)
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
			a.emitAddressOrReloc(addr)
			return nil
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
		a.emitAddressOrReloc(addr)
		return nil
	}

	// LD HL, (nn)
	if dst == "HL" && strings.HasPrefix(src, "(") && strings.HasSuffix(src, ")") {
		addr := strings.Trim(ops[1], "()")
		a.emit(0x2A)
		a.emitAddressOrReloc(addr)
		return nil
	}

	// LD (nn), HL
	if strings.HasPrefix(dst, "(") && strings.HasSuffix(dst, ")") && src == "HL" {
		addr := strings.Trim(ops[0], "()")
		a.emit(0x22)
		a.emitAddressOrReloc(addr)
		return nil
	}

	// LD rr, nn
	if p, okP := reg16Map[dst]; okP {
		a.emit(0x01 | (p << 4))
		a.emitAddressOrReloc(ops[1])
		return nil
	}

	// LD IX, nn / LD IY, nn
	if dst == "IX" {
		a.emit(0xDD, 0x21)
		a.emitAddressOrReloc(ops[1])
		return nil
	}
	if dst == "IY" {
		a.emit(0xFD, 0x21)
		a.emitAddressOrReloc(ops[1])
		return nil
	}

	return fmt.Errorf("forma de LD não suportada: LD %s, %s", ops[0], ops[1])
}

func (a *Assembler) emit(bytes ...uint8) {
	a.codeBytes = append(a.codeBytes, bytes...)
}

func (a *Assembler) parseConstant(s string) int64 {
	s = strings.TrimSpace(s)
	if val, ok := a.constants[s]; ok {
		return val
	}
	if strings.HasPrefix(s, "$") || strings.HasPrefix(s, "#") {
		v, _ := parseHex(s[1:])
		return v
	}
	if strings.HasSuffix(strings.ToLower(s), "h") {
		v, _ := parseHex(s[:len(s)-1])
		return v
	}
	if strings.HasPrefix(strings.ToLower(s), "0x") {
		v, _ := parseHex(s[2:])
		return v
	}
	var val int64
	_, _ = fmt.Sscanf(s, "%v", &val)
	return val
}

func (a *Assembler) parseImm8(s string) uint8 {
	s = strings.TrimSpace(s)
	if val, ok := a.constants[s]; ok {
		return uint8(val & 0xFF)
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

func (a *Assembler) emitAddressOrReloc(symbolOrAddr string) {
	currOffset := uint16(len(a.codeBytes))
	symbolOrAddr = strings.TrimSpace(symbolOrAddr)

	// Se for uma constante definida por EQU
	if cVal, ok := a.constants[symbolOrAddr]; ok {
		lo := uint8(cVal & 0xFF)
		hi := uint8((cVal >> 8) & 0xFF)
		a.emit(lo, hi)
		return
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
		return
	}

	// É um símbolo/label (precisa de relocation ABS16 no .MOB)
	a.relocs = append(a.relocs, tempReloc{
		offset:     currOffset,
		symbolName: symbolOrAddr,
		relocType:  mob.RelocAbs16,
	})

	// Espaço reservado para o endereço de 16 bits
	a.emit(0x00, 0x00)
}

func (a *Assembler) emitRelativeOrReloc(symbolOrTarget string) {
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
			return
		}
		a.emit(uint8(int8(disp)))
		return
	}

	// Símbolo não resolvido no mesmo escopo local -> Relocation REL8
	a.relocs = append(a.relocs, tempReloc{
		offset:     currOffset,
		symbolName: symbolOrTarget,
		relocType:  mob.RelocRel8,
	})
	a.emit(0x00)
}

func parseHex(s string) (int64, error) {
	var val int64
	n, err := fmt.Sscanf(s, "%x", &val)
	if n == 0 || err != nil {
		return 0, fmt.Errorf("hex inválido")
	}
	return val, nil
}
