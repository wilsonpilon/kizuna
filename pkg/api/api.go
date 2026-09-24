// Package api lê os descritores de API da MSXLIB (lib/api/*.api) e gera o
// código Z80 que chama uma rotina de convenção de REGISTRADORES a partir de
// uma linguagem de alto nível (DIGNAC/WIRTH80).
//
// As rotinas da MSXLIB recebem parâmetros em registradores (A, BC, HL...),
// não na pilha como as procedures escritas em BASIC/Pascal. Sem um descritor,
// cada rotina exigiria um comando dedicado no compilador. Com ele, o
// compilador só precisa saber, por rotina: quais parâmetros existem, em que
// registrador cada um entra e em que registrador o valor volta.
//
// # Formato (uma declaração por linha; ';' inicia comentário)
//
//	proc VDP_SetColor(fg: byte in H, bg: byte in L)
//	func StrLen(s: word in HL): word out HL
//	func MATH_Random8(): byte out A
//	alias basic  Random = MATH_RandomRange
//	alias pascal Random = MATH_RandomRange
//	alias all    Rnd    = MATH_RandomRange
//
// Tipos: byte (1 registrador de 8 bits: A B C D E H L), word e ptr (par de
// 16 bits: BC DE HL). O valor de retorno de uma func sai em A, BC, DE ou HL.
//
// # Contrato das rotinas descritas
//
//   - preservam SP (e a pilha em geral);
//   - podem destruir QUALQUER outro registrador -- inclusive IX, que o DIGNAC e
//     o WIRTH80 usam como frame pointer: EmitCall salva e restaura IX em volta
//     da chamada, então quem escreve a rotina não precisa se preocupar com isso
//     (rotinas que usam BIOS/CALSLT destroem IX de qualquer forma).
//
// # Geração de código
//
// O compilador avalia cada argumento (em HL) e o empilha, da esquerda para a
// direita. LoadRegisters devolve a sequência que desempilha tudo nos
// registradores certos; ReturnToHL converte o retorno para HL, que é onde os
// compiladores esperam o valor de uma expressão. Parâmetros de 8 bits são
// desempilhados num par de registradores que ainda não contenha nada já
// carregado (ver Routine.plan); Validate rejeita assinaturas para as quais
// não existe tal ordem.
package api

import (
	"fmt"
	"sort"
	"strings"
)

// Lang identifica a linguagem de um alias.
type Lang string

const (
	LangBasic  Lang = "basic"
	LangPascal Lang = "pascal"
	LangAll    Lang = "all"
)

// Param é um parâmetro de uma rotina.
type Param struct {
	Name string
	Type string // "byte", "word" ou "ptr"
	Reg  string // "A".."L" (byte) ou "BC"/"DE"/"HL" (word/ptr)
}

// Routine é uma rotina da MSXLIB descrita num .api.
type Routine struct {
	Name   string
	Params []Param
	Func   bool   // true = func (devolve valor)
	Ret    string // "" (proc) ou A/BC/DE/HL
	Source string // "arquivo:linha", para mensagens de erro
}

// Set é um conjunto de rotinas carregadas de um ou mais arquivos .api.
type Set struct {
	routines map[string]*Routine // chave: nome em maiúsculas
	aliases  map[Lang]map[string]string
}

// NewSet cria um conjunto vazio.
func NewSet() *Set {
	return &Set{
		routines: make(map[string]*Routine),
		aliases: map[Lang]map[string]string{
			LangBasic: {}, LangPascal: {}, LangAll: {},
		},
	}
}

// Lookup procura uma rotina pelo nome (sem diferenciar maiúsculas) ou por um
// alias válido para a linguagem. Devolve a rotina com o nome CANÔNICO, que é o
// que o compilador deve emitir no CALL/EXTERN.
func (s *Set) Lookup(lang Lang, name string) (*Routine, bool) {
	if s == nil {
		return nil, false
	}
	key := strings.ToUpper(name)
	if target, ok := s.aliases[lang][key]; ok {
		key = target
	} else if target, ok := s.aliases[LangAll][key]; ok {
		key = target
	}
	r, ok := s.routines[key]
	return r, ok
}

// Names devolve os nomes canônicos, em ordem alfabética.
func (s *Set) Names() []string {
	var out []string
	for _, r := range s.routines {
		out = append(out, r.Name)
	}
	sort.Strings(out)
	return out
}

// Len é a quantidade de rotinas.
func (s *Set) Len() int {
	if s == nil {
		return 0
	}
	return len(s.routines)
}

// Merge acrescenta as declarações de src.
func (s *Set) Merge(src string, filename string) error {
	type pendingAlias struct {
		lang        Lang
		alias, dest string
		where       string
	}
	var pending []pendingAlias

	for i, raw := range strings.Split(src, "\n") {
		where := fmt.Sprintf("%s:%d", filename, i+1)
		line := raw
		if k := strings.Index(line, ";"); k >= 0 {
			line = line[:k]
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		word, rest := splitWord(line)
		switch strings.ToLower(word) {
		case "proc", "func":
			r, err := parseRoutine(strings.EqualFold(word, "func"), rest, where)
			if err != nil {
				return err
			}
			key := strings.ToUpper(r.Name)
			if prev, dup := s.routines[key]; dup {
				return fmt.Errorf("%s: rotina %q já declarada em %s", where, r.Name, prev.Source)
			}
			s.routines[key] = r
		case "alias":
			lang, body := splitWord(rest)
			l := Lang(strings.ToLower(lang))
			if l != LangBasic && l != LangPascal && l != LangAll {
				return fmt.Errorf("%s: linguagem de alias inválida %q (use basic, pascal ou all)", where, lang)
			}
			left, right, ok := strings.Cut(body, "=")
			if !ok {
				return fmt.Errorf("%s: alias precisa da forma \"alias <lang> <apelido> = <rotina>\"", where)
			}
			a, d := strings.TrimSpace(left), strings.TrimSpace(right)
			if !isIdent(a) || !isIdent(d) {
				return fmt.Errorf("%s: alias com nome inválido: %q = %q", where, a, d)
			}
			pending = append(pending, pendingAlias{l, strings.ToUpper(a), strings.ToUpper(d), where})
		default:
			return fmt.Errorf("%s: declaração desconhecida %q (esperado proc, func ou alias)", where, word)
		}
	}
	// Aliases só depois de todas as rotinas do arquivo, para poderem apontar
	// para rotinas declaradas mais abaixo.
	for _, p := range pending {
		if _, ok := s.routines[p.dest]; !ok {
			return fmt.Errorf("%s: alias aponta para rotina inexistente %q", p.where, p.dest)
		}
		if prev, dup := s.aliases[p.lang][p.alias]; dup && prev != p.dest {
			return fmt.Errorf("%s: alias %q já usado para %s", p.where, p.alias, prev)
		}
		if _, clash := s.routines[p.alias]; clash && p.alias != p.dest {
			return fmt.Errorf("%s: alias %q colide com uma rotina de mesmo nome", p.where, p.alias)
		}
		s.aliases[p.lang][p.alias] = p.dest
	}
	return nil
}

// Parse é atalho para NewSet + Merge.
func Parse(src, filename string) (*Set, error) {
	s := NewSet()
	if err := s.Merge(src, filename); err != nil {
		return nil, err
	}
	return s, nil
}

func splitWord(s string) (string, string) {
	s = strings.TrimSpace(s)
	i := strings.IndexAny(s, " \t")
	if i < 0 {
		return s, ""
	}
	return s[:i], strings.TrimSpace(s[i+1:])
}

func isIdent(s string) bool {
	if s == "" {
		return false
	}
	for i, c := range s {
		switch {
		case c == '_' || (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z'):
		case i > 0 && c >= '0' && c <= '9':
		default:
			return false
		}
	}
	return true
}

var (
	regs8  = map[string]bool{"A": true, "B": true, "C": true, "D": true, "E": true, "H": true, "L": true}
	regs16 = map[string]bool{"BC": true, "DE": true, "HL": true}
)

func parseRoutine(isFunc bool, rest, where string) (*Routine, error) {
	open := strings.Index(rest, "(")
	closeIdx := strings.LastIndex(rest, ")")
	if open < 0 || closeIdx < open {
		return nil, fmt.Errorf("%s: esperado \"nome(parametros)\"", where)
	}
	r := &Routine{Name: strings.TrimSpace(rest[:open]), Func: isFunc, Source: where}
	if !isIdent(r.Name) {
		return nil, fmt.Errorf("%s: nome de rotina inválido %q", where, r.Name)
	}
	tail := strings.TrimSpace(rest[closeIdx+1:])
	if isFunc {
		if !strings.HasPrefix(tail, ":") {
			return nil, fmt.Errorf("%s: func %s precisa declarar o retorno: \": <tipo> out <REG>\"", where, r.Name)
		}
		fields := strings.Fields(strings.TrimPrefix(tail, ":"))
		if len(fields) != 3 || fields[1] != "out" {
			return nil, fmt.Errorf("%s: retorno de %s deve ser \": <tipo> out <REG>\"", where, r.Name)
		}
		typ, reg := strings.ToLower(fields[0]), strings.ToUpper(fields[2])
		if err := checkTypeReg(typ, reg); err != nil {
			return nil, fmt.Errorf("%s: retorno de %s: %v", where, r.Name, err)
		}
		if reg != "A" && !regs16[reg] {
			return nil, fmt.Errorf("%s: retorno de %s só pode sair em A, BC, DE ou HL (não %s)", where, r.Name, reg)
		}
		r.Ret = reg
	} else if tail != "" {
		return nil, fmt.Errorf("%s: proc %s não devolve valor (texto inesperado %q)", where, r.Name, tail)
	}

	inner := strings.TrimSpace(rest[open+1 : closeIdx])
	if inner != "" {
		for _, ptxt := range strings.Split(inner, ",") {
			name, spec, ok := strings.Cut(ptxt, ":")
			if !ok {
				return nil, fmt.Errorf("%s: parâmetro %q sem \": <tipo> in <REG>\"", where, strings.TrimSpace(ptxt))
			}
			name = strings.TrimSpace(name)
			fields := strings.Fields(spec)
			if !isIdent(name) || len(fields) != 3 || fields[1] != "in" {
				return nil, fmt.Errorf("%s: parâmetro inválido %q (esperado \"nome: <tipo> in <REG>\")", where, strings.TrimSpace(ptxt))
			}
			typ, reg := strings.ToLower(fields[0]), strings.ToUpper(fields[2])
			if err := checkTypeReg(typ, reg); err != nil {
				return nil, fmt.Errorf("%s: parâmetro %s de %s: %v", where, name, r.Name, err)
			}
			r.Params = append(r.Params, Param{Name: name, Type: typ, Reg: reg})
		}
	}
	if err := r.Validate(); err != nil {
		return nil, fmt.Errorf("%s: %v", where, err)
	}
	return r, nil
}

func checkTypeReg(typ, reg string) error {
	switch typ {
	case "byte":
		if !regs8[reg] {
			return fmt.Errorf("tipo byte pede registrador de 8 bits (A B C D E H L), não %s", reg)
		}
	case "word", "ptr":
		if !regs16[reg] {
			return fmt.Errorf("tipo %s pede par de 16 bits (BC DE HL), não %s", typ, reg)
		}
	default:
		return fmt.Errorf("tipo desconhecido %q (use byte, word ou ptr)", typ)
	}
	return nil
}

// pairOf devolve o par de 16 bits a que o registrador pertence.
func pairOf(reg string) string {
	switch reg {
	case "B", "C", "BC":
		return "BC"
	case "D", "E", "DE":
		return "DE"
	case "H", "L", "HL":
		return "HL"
	}
	return "" // A
}

// Validate confere: nenhum registrador de entrada usado duas vezes (contando
// que BC cobre B e C etc.) e que existe uma ordem de POPs capaz de carregar
// todos os parâmetros (ver plan).
func (r *Routine) Validate() error {
	used := map[string]string{}
	for _, p := range r.Params {
		regs := []string{p.Reg}
		if len(p.Reg) == 2 {
			regs = []string{string(p.Reg[0]), string(p.Reg[1])}
		}
		for _, rg := range regs {
			if other, dup := used[rg]; dup {
				return fmt.Errorf("parâmetros %s e %s de %s usam o mesmo registrador (%s)", other, p.Name, r.Name, rg)
			}
			used[rg] = p.Name
		}
	}
	if _, err := r.plan(); err != nil {
		return err
	}
	return nil
}

// lowReg devolve o registrador baixo de um par (C, E ou L).
func lowReg(pair string) string { return string(pair[1]) }

// plan decide, para cada parâmetro (do último para o primeiro, que é a ordem
// em que a pilha entrega), de qual par fazer POP e se depois é preciso um LD.
//
//   - parâmetro de 16 bits: POP direto no par de destino;
//   - parâmetro de 8 bits em X: POP num par P e, se X não for o registrador
//     baixo de P, "LD X, <baixo de P>". P não pode conter nenhum registrador
//     já carregado por um parâmetro anterior nesta sequência (senão o POP o
//     destruiria). Prefere o próprio par de X, depois os demais.
func (r *Routine) plan() ([]popStep, error) {
	loaded := map[string]bool{}
	var steps []popStep
	for i := len(r.Params) - 1; i >= 0; i-- {
		p := r.Params[i]
		if len(p.Reg) == 2 {
			steps = append(steps, popStep{pair: p.Reg})
			loaded[string(p.Reg[0])], loaded[string(p.Reg[1])] = true, true
			continue
		}
		cands := []string{"BC", "DE", "HL"}
		if own := pairOf(p.Reg); own != "" {
			cands = append([]string{own}, without(cands, own)...)
		}
		chosen := ""
		for _, c := range cands {
			if !loaded[string(c[0])] && !loaded[string(c[1])] {
				chosen = c
				break
			}
		}
		if chosen == "" {
			return nil, fmt.Errorf("%s: não há par de registradores livre para carregar o parâmetro %s (%s) -- "+
				"declare os parâmetros de 8 bits DEPOIS dos de 16 bits na lista, ou use outro registrador", r.Name, p.Name, p.Reg)
		}
		st := popStep{pair: chosen}
		if lowReg(chosen) != p.Reg {
			st.move = fmt.Sprintf("    LD %s, %s", p.Reg, lowReg(chosen))
		}
		steps = append(steps, st)
		loaded[p.Reg] = true
	}
	return steps, nil
}

type popStep struct {
	pair string
	move string // "" se não precisa de LD
}

func without(list []string, drop string) []string {
	var out []string
	for _, x := range list {
		if x != drop {
			out = append(out, x)
		}
	}
	return out
}

// LoadRegisters devolve as linhas de Assembly que desempilham os argumentos
// (já empilhados da esquerda para a direita, cada um com 16 bits) nos
// registradores da rotina. nargs deve ser igual a len(r.Params).
func (r *Routine) LoadRegisters(nargs int) ([]string, error) {
	if nargs != len(r.Params) {
		return nil, fmt.Errorf("%s espera %d argumento(s), recebeu %d", r.Name, len(r.Params), nargs)
	}
	steps, err := r.plan()
	if err != nil {
		return nil, err
	}
	var out []string
	for _, st := range steps {
		out = append(out, "    POP "+st.pair)
		if st.move != "" {
			out = append(out, st.move)
		}
	}
	return out, nil
}

// ReturnToHL devolve as linhas que levam o valor de retorno para HL (nada se
// já estiver em HL, ou se a rotina for proc).
func (r *Routine) ReturnToHL() []string {
	switch r.Ret {
	case "A":
		return []string{"    LD L, A", "    LD H, 0"}
	case "BC":
		return []string{"    LD H, B", "    LD L, C"}
	case "DE":
		return []string{"    EX DE, HL"}
	}
	return nil
}

// EmitCall escreve em sb a chamada completa da rotina: avalia e empilha cada
// argumento (pushArg(i) deve gerar o argumento i em HL e fazer PUSH HL,
// exatamente como os compiladores já fazem para procedures em pilha),
// desempilha nos registradores, CALL e, se asExpr, converte o retorno para HL.
//
// Não há limpeza de pilha: todos os argumentos são desempilhados na carga.
func (r *Routine) EmitCall(sb *strings.Builder, nargs int, pushArg func(i int) error, asExpr bool) error {
	if asExpr && !r.Func {
		return fmt.Errorf("%s é uma proc (não devolve valor) e não pode ser usada numa expressão", r.Name)
	}
	loads, err := r.LoadRegisters(nargs)
	if err != nil {
		return err
	}
	for i := 0; i < nargs; i++ {
		if err := pushArg(i); err != nil {
			return err
		}
	}
	for _, l := range loads {
		sb.WriteString(l)
		sb.WriteString("\n")
	}
	// A rotina pode destruir IX (o frame pointer dos compiladores): salva e
	// restaura em volta do CALL. PUSH/POP IX não afetam A/BC/DE/HL nem flags
	// de retorno que o compilador use depois.
	sb.WriteString("    PUSH IX\n")
	sb.WriteString("    CALL " + r.Name + "\n")
	sb.WriteString("    POP IX\n")
	if asExpr {
		for _, l := range r.ReturnToHL() {
			sb.WriteString(l)
			sb.WriteString("\n")
		}
	}
	return nil
}
