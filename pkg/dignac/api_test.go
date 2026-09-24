package dignac

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wilsonpilon/kizuna/pkg/api"
	"github.com/wilsonpilon/kizuna/pkg/libtest"
	"github.com/wilsonpilon/kizuna/pkg/mob"
	"github.com/wilsonpilon/kizuna/pkg/z80sim"
)

func compileBasic(t *testing.T, src string, set *api.Set) (*mob.ObjectFile, string, error) {
	t.Helper()
	parser, err := NewParser(NewLexer(src))
	if err != nil {
		t.Fatalf("parser: %v", err)
	}
	mod, err := parser.ParseModule()
	if err != nil {
		t.Fatalf("sintaxe: %v", err)
	}
	cg := NewCodeGenerator(mod)
	cg.SetAPI(set)
	return cg.Compile()
}

func portTrace(m *z80sim.Machine) string {
	var sb strings.Builder
	for _, w := range m.Ports {
		sb.WriteString(strings.ToUpper(string("0123456789abcdef"[w.Port>>4])) + strings.ToUpper(string("0123456789abcdef"[w.Port&15])) + ":")
		sb.WriteString(strings.ToUpper(string("0123456789abcdef"[w.Value>>4])) + strings.ToUpper(string("0123456789abcdef"[w.Value&15])) + " ")
	}
	return strings.TrimSpace(sb.String())
}

// TestAPICallsFromBasic compila um programa BASIC que chama rotinas da MSXLIB
// pelos descritores reais de lib/api, liga contra a MSXLIB recém-montada e
// roda no simulador Z80.
func TestAPICallsFromBasic(t *testing.T) {
	src := `MODULE ApiDemo
BANK 0
PUBLIC Main

PROCEDURE Main()
    LOCAL t%
    t% = 5
    PSG_Write(7, 62)
    VDP_SetColor(15, 4)
    PrintDec16(Mul16(6, 7))
    BDOS_PrintChar(33)
    PrintDec16(t%)
END PROCEDURE
END MODULE
`
	obj, asm, err := compileBasic(t, src, libtest.RepoAPI(t))
	if err != nil {
		t.Fatalf("compilação: %v\n%s", err, asm)
	}
	for _, want := range []string{"EXTERN", "PSG_Write", "VDP_SetColor", "Mul16", "PrintDec16", "BDOS_PrintChar"} {
		if !strings.Contains(asm, want) {
			t.Errorf("assembly gerado não menciona %s:\n%s", want, asm)
		}
	}

	res := libtest.LinkObjects(t, obj)
	m := libtest.Machine(res)
	if err := m.Run(2_000_000); err != nil {
		t.Fatal(err)
	}
	// 6*7 = 42, '!' e depois t% = 5 -- este último prova que o frame pointer
	// (IX) sobreviveu às chamadas.
	if got := string(m.Console); got != "42!5" {
		t.Errorf("console = %q, quer %q", got, "42!5")
	}
	// PSG_Write(7,62): OUT (A0),7 ; OUT (A1),62 -- VDP_SetColor(15,4): R#7=F4
	trace := portTrace(m)
	if !strings.Contains(trace, "A0:07 A1:3E") {
		t.Errorf("traço de portas sem a escrita do PSG (R#7=3E): %s", trace)
	}
	if !strings.Contains(trace, "99:F4 99:87") {
		t.Errorf("traço de portas sem VDP R#7=F4: %s", trace)
	}
}

// TestAPIPreservesFrameIX prova, com uma rotina que DESTRÓI IX de propósito,
// que o compilador salva/restaura o frame pointer em volta da chamada.
func TestAPIPreservesFrameIX(t *testing.T) {
	evil := libtest.Assemble(t, `MODULE Evil
BANK 0
PUBLIC Evil, EvilAdd
Evil:
    LD IX, 1234h
    LD IY, 4321h
    RET
EvilAdd:
    ADD HL, DE
    LD IX, 0BEEFh
    RET
ENDMOD
`, "")
	src := `MODULE Ix
BANK 0
PUBLIC Main

PROCEDURE Main()
    LOCAL a%, b%
    a% = 11
    b% = 22
    Evil(1)
    b% = EvilAdd(a%, b%)
    PrintDec16(a%)
    BDOS_PrintChar(44)
    PrintDec16(b%)
END PROCEDURE
END MODULE
`
	full := libtest.RepoAPI(t)
	if err := full.Merge("proc Evil(x: byte in A)\nfunc EvilAdd(a: word in HL, b: word in DE): word out HL", "evil.api"); err != nil {
		t.Fatal(err)
	}
	obj, asm, err := compileBasic(t, src, full)
	if err != nil {
		t.Fatalf("compilação: %v\n%s", err, asm)
	}
	res := libtest.LinkObjects(t, obj, evil)
	m := libtest.Machine(res)
	if err := m.Run(2_000_000); err != nil {
		t.Fatal(err)
	}
	if got := string(m.Console); got != "11,33" {
		t.Errorf("console = %q, quer %q (IX destruído pela rotina não foi restaurado?)\n%s", got, "11,33", asm)
	}
}

func TestAPICompileErrors(t *testing.T) {
	set := libtest.RepoAPI(t)
	cases := []struct{ name, body, want string }{
		{"argumentos a menos", "    PSG_Write(7)", "espera 2"},
		{"argumentos a mais", "    BIOS_CLS(1)", "espera 0"},
		{"proc usada como expressão", "    LOCAL x%\n    x% = PSG_MuteAll()", "não pode ser usada numa expressão"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			src := "MODULE E\nBANK 0\nPUBLIC Main\nPROCEDURE Main()\n" + c.body + "\nEND PROCEDURE\nEND MODULE\n"
			_, _, err := compileBasic(t, src, set)
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Errorf("erro = %v, quer conter %q", err, c.want)
			}
		})
	}
}

// Um PROCEDURE do próprio programa com o mesmo nome de uma rotina descrita
// VENCE (chamada em pilha, sem EXTERN da rotina da MSXLIB).
func TestAPIUserProcedureWins(t *testing.T) {
	src := `MODULE U
BANK 0
PUBLIC Main

PROCEDURE Mul16(a%, b%)
    PRINT "meu"
END PROCEDURE

PROCEDURE Main()
    Mul16(1, 2)
END PROCEDURE
END MODULE
`
	obj, asm, err := compileBasic(t, src, libtest.RepoAPI(t))
	if err != nil {
		t.Fatalf("%v\n%s", err, asm)
	}
	for _, s := range obj.Symbols {
		if s.Name == "Mul16" && s.Class == mob.SymbolExtern {
			t.Errorf("Mul16 virou EXTERN -- a procedure do usuário deveria vencer:\n%s", asm)
		}
	}
}

// Sem descritores carregados, o comportamento é exatamente o de antes.
func TestAPINilSetIsHarmless(t *testing.T) {
	src := "MODULE N\nBANK 0\nPUBLIC Main\nPROCEDURE Foo(a%)\nEND PROCEDURE\nPROCEDURE Main()\n    Foo(1)\nEND PROCEDURE\nEND MODULE\n"
	if _, asm, err := compileBasic(t, src, nil); err != nil {
		t.Fatalf("%v\n%s", err, asm)
	}
}

// O sample/api/api_demo.bas é o exemplo que o Wilson testa no hardware: este
// teste garante que ele compila, liga e imprime o que o manual promete.
func TestSampleAPIDemoBasic(t *testing.T) {
	root := libtest.RepoRoot(t)
	data, err := os.ReadFile(filepath.Join(root, "sample", "api", "api_demo.bas"))
	if err != nil {
		t.Fatal(err)
	}
	obj, asm, err := compileBasic(t, string(data), libtest.RepoAPI(t))
	if err != nil {
		t.Fatalf("%v\n%s", err, asm)
	}
	m := libtest.Machine(libtest.LinkObjects(t, obj))
	if err := m.Run(5_000_000); err != nil {
		t.Fatal(err)
	}
	out := string(m.Console)
	for _, want := range []string{"6 * 7 =\r\n42", "100 / 7 =\r\n14", "255 em hexa:\r\nFF"} {
		if !strings.Contains(out, want) {
			t.Errorf("saída sem %q:\n%q", want, out)
		}
	}
	// tom de 440 Hz: período 254 = 00FEh nos registradores 0/1 do PSG, volume 12 no R#8
	trace := portTrace(m)
	for _, want := range []string{"A0:00 A1:FE", "A0:01 A1:00", "A0:08 A1:0C"} {
		if !strings.Contains(trace, want) {
			t.Errorf("traço do PSG sem %q:\n%s", want, trace)
		}
	}
}

// Sem descritores, chamar uma rotina da MSXLIB (convencao de registradores) e
// erro claro -- nunca uma chamada de pilha que liga e calcula errado.
func TestAPIMissingDescriptorIsAnError(t *testing.T) {
	src := "MODULE E\nBANK 0\nPUBLIC Main\nPROCEDURE Main()\n    PrintDec16(Mul16(6, 7))\nEND PROCEDURE\nEND MODULE\n"
	for _, set := range []*api.Set{nil, api.NewSet()} {
		_, _, err := compileBasic(t, src, set)
		if err == nil || !strings.Contains(err.Error(), "PROCEDURE") || !strings.Contains(err.Error(), "-api") {
			t.Errorf("erro = %v, quer mensagem sobre chamada nao declarada com dica de -api", err)
		}
	}
}
