package api_test

import (
	"fmt"
	"strings"
	"testing"

	. "github.com/wilsonpilon/kizuna/pkg/api"
	"github.com/wilsonpilon/kizuna/pkg/libtest"
	"github.com/wilsonpilon/kizuna/pkg/musubi"
)

func TestParseBasics(t *testing.T) {
	src := `
; comentário de linha inteira
proc VDP_SetColor(fg: byte in H, bg: byte in L)   ; comentário no fim
func StrLen(s: ptr in HL): word out HL
func MATH_Random8(): byte out A
alias basic  Rnd = MATH_Random8
alias pascal Random = MATH_Random8
alias all    Len = StrLen
`
	s, err := Parse(src, "t.api")
	if err != nil {
		t.Fatal(err)
	}
	if s.Len() != 3 {
		t.Fatalf("Len = %d, quer 3", s.Len())
	}

	r, ok := s.Lookup(LangBasic, "vdp_setcolor") // sem diferenciar maiúsculas
	if !ok || r.Name != "VDP_SetColor" || len(r.Params) != 2 || r.Func {
		t.Fatalf("Lookup VDP_SetColor = %+v, %v", r, ok)
	}
	if r.Params[0] != (Param{"fg", "byte", "H"}) || r.Params[1] != (Param{"bg", "byte", "L"}) {
		t.Errorf("params = %+v", r.Params)
	}

	// alias por linguagem
	if r, ok := s.Lookup(LangBasic, "RND"); !ok || r.Name != "MATH_Random8" {
		t.Errorf("alias basic Rnd: %+v %v", r, ok)
	}
	if _, ok := s.Lookup(LangPascal, "Rnd"); ok {
		t.Error("alias basic não deveria valer em Pascal")
	}
	if r, ok := s.Lookup(LangPascal, "random"); !ok || r.Name != "MATH_Random8" {
		t.Errorf("alias pascal Random: %+v %v", r, ok)
	}
	// alias all vale nas duas
	for _, l := range []Lang{LangBasic, LangPascal} {
		if r, ok := s.Lookup(l, "Len"); !ok || r.Name != "StrLen" {
			t.Errorf("alias all Len em %s: %+v %v", l, r, ok)
		}
	}
	if r, _ := s.Lookup(LangBasic, "StrLen"); r.Ret != "HL" || !r.Func {
		t.Errorf("StrLen = %+v", r)
	}
	if _, ok := (*Set)(nil).Lookup(LangBasic, "x"); ok {
		t.Error("Set nil não deve achar nada")
	}
}

func TestParseErrors(t *testing.T) {
	cases := []struct{ name, src, want string }{
		{"sem parênteses", "proc Foo", "nome(parametros)"},
		{"nome inválido", "proc 1Foo()", "nome de rotina inválido"},
		{"func sem retorno", "func Foo()", "precisa declarar o retorno"},
		{"proc com retorno", "proc Foo(): byte out A", "não devolve valor"},
		{"retorno tipo errado", "func Foo(): byte out HL", "8 bits"},
		{"retorno reg 8 bits não-A", "func Foo(): byte out B", "só pode sair em A, BC, DE ou HL"},
		{"tipo desconhecido", "proc Foo(x: long in A)", "tipo desconhecido"},
		{"word em reg de 8 bits", "proc Foo(x: word in A)", "16 bits"},
		{"byte em par", "proc Foo(x: byte in HL)", "8 bits"},
		{"param sem tipo", "proc Foo(x)", "sem"},
		{"param sem in", "proc Foo(x: byte A)", "parâmetro inválido"},
		{"reg repetido", "proc Foo(a: byte in B, b: byte in B)", "mesmo registrador"},
		{"par cobre metade", "proc Foo(a: word in BC, b: byte in C)", "mesmo registrador"},
		{"sem par livre para o byte", "proc Foo(a: byte in A, b: word in BC, c: word in DE, d: word in HL)", "não há par de registradores livre"},
		{"decl desconhecida", "sub Foo()", "declaração desconhecida"},
		{"rotina duplicada", "proc Foo()\nproc FOO()", "já declarada"},
		{"alias lang inválida", "proc Foo()\nalias c X = Foo", "linguagem de alias inválida"},
		{"alias sem igual", "proc Foo()\nalias all X Foo", "alias precisa"},
		{"alias destino inexistente", "alias all X = Nope", "inexistente"},
		{"alias colide", "proc Foo()\nproc Bar()\nalias all Bar = Foo", "colide"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := Parse(c.src, "t.api")
			if err == nil {
				t.Fatalf("esperava erro contendo %q", c.want)
			}
			if !strings.Contains(err.Error(), c.want) {
				t.Errorf("erro %q não contém %q", err, c.want)
			}
			if !strings.Contains(err.Error(), "t.api:") {
				t.Errorf("erro sem arquivo:linha: %q", err)
			}
		})
	}
}

func TestLoadRegistersArgCount(t *testing.T) {
	s, _ := Parse("proc P(a: byte in A)", "t.api")
	r, _ := s.Lookup(LangAll, "P")
	if _, err := r.LoadRegisters(2); err == nil || !strings.Contains(err.Error(), "espera 1") {
		t.Errorf("erro = %v", err)
	}
}

func TestReturnToHL(t *testing.T) {
	cases := map[string][]string{
		"A":  {"    LD L, A", "    LD H, 0"},
		"BC": {"    LD H, B", "    LD L, C"},
		"DE": {"    EX DE, HL"},
		"HL": nil,
	}
	for reg, want := range cases {
		typ := "word"
		if reg == "A" {
			typ = "byte"
		}
		s, err := Parse(fmt.Sprintf("func F(): %s out %s", typ, reg), "t.api")
		if err != nil {
			t.Fatal(err)
		}
		r, _ := s.Lookup(LangAll, "F")
		got := r.ReturnToHL()
		if strings.Join(got, "|") != strings.Join(want, "|") {
			t.Errorf("ret %s: %v, quer %v", reg, got, want)
		}
	}
}

// regState são os registradores que o teste de execução confere.
type regState struct{ A, B, C, D, E, H, L byte }

// TestGeneratedLoadCodeOnZ80 é o teste que importa: gera de verdade o código
// de chamada (avalia argumentos, empilha, LoadRegisters), monta com o KAJI80,
// roda no simulador Z80 e confere que cada parâmetro caiu no registrador
// declarado E que a pilha voltou balanceada.
func TestGeneratedLoadCodeOnZ80(t *testing.T) {
	cases := []struct {
		name string
		sig  string
		args []int
		want regState
	}{
		{"dois bytes em H e L", "proc F(fg: byte in H, bg: byte in L)", []int{7, 4}, regState{H: 7, L: 4}},
		{"dois bytes em H e L, ordem inversa", "proc F(bg: byte in L, fg: byte in H)", []int{7, 4}, regState{H: 4, L: 7}},
		{"VDP_WriteReg: reg em C, valor em B", "proc F(reg: byte in C, val: byte in B)", []int{7, 0xE1}, regState{C: 7, B: 0xE1}},
		{"sprite: A,H,L,D,E", "proc F(i: byte in A, y: byte in H, x: byte in L, p: byte in D, c: byte in E)",
			[]int{3, 100, 50, 8, 15}, regState{A: 3, H: 100, L: 50, D: 8, E: 15}},
		{"três words", "proc F(d: word in HL, s: word in DE, n: word in BC)", []int{0x1234, 0x5678, 0x9ABC},
			regState{H: 0x12, L: 0x34, D: 0x56, E: 0x78, B: 0x9A, C: 0xBC}},
		{"handle B, buf DE, n HL (BDOS_FileRead)", "proc F(h: byte in B, buf: ptr in DE, n: word in HL)",
			[]int{5, 0x4000, 0x0100}, regState{B: 5, D: 0x40, E: 0, H: 1, L: 0}},
		{"byte trunca para o byte baixo", "proc F(x: byte in A)", []int{0x1FF}, regState{A: 0xFF}},
		{"A com par ocupado em HL", "proc F(m: byte in A, p: word in HL)", []int{9, 0xBEEF}, regState{A: 9, H: 0xBE, L: 0xEF}},
		{"A e byte em E", "proc F(a: byte in A, e: byte in E)", []int{1, 2}, regState{A: 1, E: 2}},
		{"oito bits em todos os pares", "proc F(b: byte in B, d: byte in D, h: byte in H)", []int{1, 2, 3}, regState{B: 1, D: 2, H: 3}},
		{"byte em A depois de três words", "proc F(b: word in BC, d: word in DE, h: word in HL, a: byte in A)",
			[]int{0x0102, 0x0304, 0x0506, 7}, regState{B: 1, C: 2, D: 3, E: 4, H: 5, L: 6, A: 7}},
		{"sem parâmetros", "proc F()", nil, regState{}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s, err := Parse(c.sig, "t.api")
			if err != nil {
				t.Fatal(err)
			}
			r, _ := s.Lookup(LangAll, "F")
			loads, err := r.LoadRegisters(len(c.args))
			if err != nil {
				t.Fatal(err)
			}

			var sb strings.Builder
			sb.WriteString("MODULE T\nBANK 0\nPUBLIC Start\nStart:\n")
			for _, a := range c.args {
				// como os compiladores: avalia em HL e empilha
				fmt.Fprintf(&sb, "    LD HL, %d\n    PUSH HL\n", a&0xFFFF)
			}
			sb.WriteString(strings.Join(loads, "\n"))
			sb.WriteString("\n    HALT\n")

			obj := libtest.Assemble(t, sb.String(), "")
			res, err := musubi.NewLinker(musubi.DefaultConfig()).Link(obj)
			if err != nil {
				t.Fatal(err)
			}
			m := libtest.Machine(res)
			if err := m.Run(10000); err != nil {
				t.Fatal(err)
			}
			cpu := m.CPU
			got := regState{cpu.A, cpu.B, cpu.C, cpu.D, cpu.E, cpu.H, cpu.L}

			// Só compara os registradores que são parâmetros; os demais são
			// rascunho e podem ter qualquer valor.
			isParam := func(reg string) bool {
				for _, p := range r.Params {
					if p.Reg == reg || (len(p.Reg) == 2 && strings.Contains(p.Reg, reg)) {
						return true
					}
				}
				return false
			}
			check := func(name string, g, w byte) {
				if isParam(name) && g != w {
					t.Errorf("registrador %s = %02X, quer %02X\nlinhas: %s", name, g, w, strings.Join(loads, " / "))
				}
			}
			check("A", got.A, c.want.A)
			check("B", got.B, c.want.B)
			check("C", got.C, c.want.C)
			check("D", got.D, c.want.D)
			check("E", got.E, c.want.E)
			check("H", got.H, c.want.H)
			check("L", got.L, c.want.L)
			if sp := cpu.SP(); sp != 0xFF00 {
				t.Errorf("SP = %04X, quer FF00 (pilha desbalanceada)", sp)
			}
		})
	}
}
