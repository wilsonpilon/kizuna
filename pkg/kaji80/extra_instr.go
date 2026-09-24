package kaji80

import (
	"fmt"
	"strings"
)

// Instruções Z80 que o montador ainda não cobria: operações de bloco (LDIR,
// CPIR, ...), RLD/RRD, RETI/RETN, DAA, IM, RST, LD I,A / LD R,A, LD rr,(nn) com
// BC/DE/SP e IX/IY, JP (HL)/(IX)/(IY), IN/OUT com outros registradores, e as
// formas indexadas (IX+d)/(IY+d) de rotação/BIT/RES/SET/INC/DEC.
//
// extraEncode devolve os bytes fixos da instrução e, quando há um endereço
// absoluto de 16 bits depois deles, o operando de endereço (resolvido por
// emitAddressOrReloc, então rótulos e EXTERN geram relocation normalmente).
// Quando a instrução NÃO é uma destas formas, devolve handled=false e o
// caminho normal do montador cuida dela -- assim nada que já funcionava muda.
// O mesmo cálculo serve ao Pass 1 (tamanho) e ao Pass 2 (bytes), o que
// mantém os dois passes sempre de acordo.

var extraFixed = map[string][]uint8{
	"LDI": {0xED, 0xA0}, "LDIR": {0xED, 0xB0}, "LDD": {0xED, 0xA8}, "LDDR": {0xED, 0xB8},
	"CPI": {0xED, 0xA1}, "CPIR": {0xED, 0xB1}, "CPD": {0xED, 0xA9}, "CPDR": {0xED, 0xB9},
	"INI": {0xED, 0xA2}, "INIR": {0xED, 0xB2}, "IND": {0xED, 0xAA}, "INDR": {0xED, 0xBA},
	"OUTI": {0xED, 0xA3}, "OTIR": {0xED, 0xB3}, "OUTD": {0xED, 0xAB}, "OTDR": {0xED, 0xBB},
	"RLD": {0xED, 0x6F}, "RRD": {0xED, 0x67},
	"RETI": {0xED, 0x4D}, "RETN": {0xED, 0x45},
	"DAA": {0x27},
}

// Rotações/deslocamentos: código do byte final das formas (HL) e (IX+d)/(IY+d).
var extraRotate = map[string]uint8{
	"RLC": 0x06, "RRC": 0x0E, "RL": 0x16, "RR": 0x1E, "SLA": 0x26, "SRA": 0x2E, "SRL": 0x3E,
}

func isPlainMemOperand(op string) bool {
	up := strings.ToUpper(strings.ReplaceAll(op, " ", ""))
	switch up {
	case "(HL)", "(BC)", "(DE)", "(SP)", "(C)":
		return false
	}
	if isIndexedOperand(op) {
		return false
	}
	return strings.HasPrefix(up, "(") && strings.HasSuffix(up, ")")
}

func stripParens(op string) string {
	op = strings.TrimSpace(op)
	return strings.TrimSpace(op[1 : len(op)-1])
}

func indexPrefix(op string) (uint8, int8, bool) {
	isIX, isIY, d, ok := parseIndexed(op)
	if !ok {
		return 0, 0, false
	}
	if isIX {
		return 0xDD, d, true
	}
	if isIY {
		return 0xFD, d, true
	}
	return 0, 0, false
}

func (a *Assembler) extraEncode(mnem string, ops []string) (b []byte, addr string, handled bool, err error) {
	up := make([]string, len(ops))
	for i, o := range ops {
		up[i] = strings.ToUpper(strings.ReplaceAll(o, " ", ""))
	}

	if fixed, ok := extraFixed[mnem]; ok && len(ops) == 0 {
		return fixed, "", true, nil
	}

	switch mnem {
	case "IM":
		if len(ops) != 1 {
			return nil, "", true, fmt.Errorf("IM requer 1 operando (0, 1 ou 2)")
		}
		v, ok, e := a.EvalExpr(ops[0])
		if e != nil || !ok {
			return nil, "", true, fmt.Errorf("IM: modo inválido %q (use 0, 1 ou 2)", ops[0])
		}
		switch v {
		case 0:
			return []byte{0xED, 0x46}, "", true, nil
		case 1:
			return []byte{0xED, 0x56}, "", true, nil
		case 2:
			return []byte{0xED, 0x5E}, "", true, nil
		}
		return nil, "", true, fmt.Errorf("IM: modo %v inválido (use 0, 1 ou 2)", v)

	case "RST":
		if len(ops) != 1 {
			return nil, "", true, fmt.Errorf("RST requer 1 operando (0, 8, 10h, ... 38h)")
		}
		v, ok, e := a.EvalExpr(ops[0])
		if e != nil || !ok || v < 0 || v > 0x38 || v != float64(int(v)) || int(v)%8 != 0 {
			return nil, "", true, fmt.Errorf("RST: alvo inválido %q (só 00h, 08h, 10h, 18h, 20h, 28h, 30h ou 38h)", ops[0])
		}
		return []byte{0xC7 | uint8(int(v))}, "", true, nil

	case "LD":
		if len(ops) != 2 {
			return nil, "", false, nil
		}
		switch {
		case up[0] == "I" && up[1] == "A":
			return []byte{0xED, 0x47}, "", true, nil
		case up[0] == "R" && up[1] == "A":
			return []byte{0xED, 0x4F}, "", true, nil
		}
		loads := map[string]uint8{"BC": 0x4B, "DE": 0x5B, "SP": 0x7B}
		stores := map[string]uint8{"BC": 0x43, "DE": 0x53, "SP": 0x73}
		if op, ok := loads[up[0]]; ok && isPlainMemOperand(ops[1]) {
			return []byte{0xED, op}, stripParens(ops[1]), true, nil
		}
		if op, ok := stores[up[1]]; ok && isPlainMemOperand(ops[0]) {
			return []byte{0xED, op}, stripParens(ops[0]), true, nil
		}
		for _, ix := range []struct {
			name   string
			prefix uint8
		}{{"IX", 0xDD}, {"IY", 0xFD}} {
			if up[0] == ix.name && isPlainMemOperand(ops[1]) {
				return []byte{ix.prefix, 0x2A}, stripParens(ops[1]), true, nil
			}
			if up[1] == ix.name && isPlainMemOperand(ops[0]) {
				return []byte{ix.prefix, 0x22}, stripParens(ops[0]), true, nil
			}
		}
		return nil, "", false, nil

	case "JP":
		if len(ops) == 1 {
			switch up[0] {
			case "(HL)":
				return []byte{0xE9}, "", true, nil
			case "(IX)":
				return []byte{0xDD, 0xE9}, "", true, nil
			case "(IY)":
				return []byte{0xFD, 0xE9}, "", true, nil
			}
		}
		return nil, "", false, nil

	case "ADD":
		if len(ops) == 2 && up[0] == "IX" && up[1] == "IX" {
			return []byte{0xDD, 0x29}, "", true, nil
		}
		if len(ops) == 2 && up[0] == "IY" && up[1] == "IY" {
			return []byte{0xFD, 0x29}, "", true, nil
		}
		return nil, "", false, nil

	case "IN":
		// IN r,(C) para qualquer registrador de 8 bits. Inclui A: o caminho normal
		// montava "IN A,(C)" como "IN A,(0)" (DB 00) em silêncio -- (C) não é uma
		// porta imediata.
		if len(ops) == 2 && up[1] == "(C)" {
			if r, ok := reg8Map[up[0]]; ok {
				return []byte{0xED, 0x40 | r<<3}, "", true, nil
			}
		}
		return nil, "", false, nil

	case "OUT":
		if len(ops) == 2 && up[0] == "(C)" {
			if r, ok := reg8Map[up[1]]; ok {
				return []byte{0xED, 0x41 | r<<3}, "", true, nil
			}
		}
		return nil, "", false, nil

	case "INC", "DEC":
		if len(ops) == 1 {
			if p, d, ok := indexPrefix(ops[0]); ok {
				op := uint8(0x34)
				if mnem == "DEC" {
					op = 0x35
				}
				return []byte{p, op, uint8(d)}, "", true, nil
			}
		}
		return nil, "", false, nil

	case "BIT", "RES", "SET":
		if len(ops) == 2 {
			if p, d, ok := indexPrefix(ops[1]); ok {
				n, okN, e := a.EvalExpr(ops[0])
				if e != nil || !okN || n < 0 || n > 7 || n != float64(int(n)) {
					return nil, "", true, fmt.Errorf("%s: número de bit inválido %q (0..7)", mnem, ops[0])
				}
				base := map[string]uint8{"BIT": 0x40, "RES": 0x80, "SET": 0xC0}[mnem]
				return []byte{p, 0xCB, uint8(d), base | uint8(int(n))<<3 | 0x06}, "", true, nil
			}
		}
		return nil, "", false, nil
	}

	if code, ok := extraRotate[mnem]; ok && len(ops) == 1 {
		if p, d, okI := indexPrefix(ops[0]); okI {
			return []byte{p, 0xCB, uint8(d), code}, "", true, nil
		}
	}
	return nil, "", false, nil
}

// extraSize é o tamanho da instrução tratada por extraEncode.
func (a *Assembler) extraSize(mnem string, ops []string) (uint16, bool, error) {
	b, addr, handled, err := a.extraEncode(mnem, ops)
	if !handled {
		return 0, false, nil
	}
	if err != nil {
		return 0, true, err
	}
	n := uint16(len(b))
	if addr != "" {
		n += 2
	}
	return n, true, nil
}

// encodeExtra emite a instrução tratada por extraEncode.
func (a *Assembler) encodeExtra(mnem string, ops []string) (bool, error) {
	b, addr, handled, err := a.extraEncode(mnem, ops)
	if !handled {
		return false, nil
	}
	if err != nil {
		return true, err
	}
	a.emit(b...)
	if addr != "" {
		if err := a.emitAddressOrReloc(addr); err != nil {
			return true, err
		}
	}
	return true, nil
}
