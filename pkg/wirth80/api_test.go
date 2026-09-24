package wirth80

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wilsonpilon/kizuna/pkg/api"
	"github.com/wilsonpilon/kizuna/pkg/libtest"
	"github.com/wilsonpilon/kizuna/pkg/mob"
)

func compilePascal(t *testing.T, src string, set *api.Set) (*mob.ObjectFile, string, error) {
	t.Helper()
	parser, err := NewParser(NewLexer(src))
	if err != nil {
		t.Fatalf("parser: %v", err)
	}
	prog, err := parser.ParseProgram()
	if err != nil {
		t.Fatalf("sintaxe: %v", err)
	}
	cg := NewCodeGenerator(prog)
	cg.SetAPI(set)
	return cg.Compile()
}

// TestAPICallsFromPascal: o mesmo programa do teste do DIGNAC, em Pascal,
// contra os descritores reais de lib/api, no simulador Z80.
func TestAPICallsFromPascal(t *testing.T) {
	src := `program ApiDemo;
var
    t: Integer;
begin
    t := 5;
    PSG_Write(7, 62);
    VDP_SetColor(15, 4);
    PrintDec16(Mul16(6, 7));
    BDOS_PrintChar(33);
    PrintDec16(t);
end.
`
	obj, asm, err := compilePascal(t, src, libtest.RepoAPI(t))
	if err != nil {
		t.Fatalf("compilação: %v\n%s", err, asm)
	}
	m := libtest.Machine(libtest.LinkObjects(t, obj))
	if err := m.Run(2_000_000); err != nil {
		t.Fatal(err)
	}
	if got := string(m.Console); got != "42!5" {
		t.Errorf("console = %q, quer %q\n%s", got, "42!5", asm)
	}
	found := func(port, val byte) bool {
		for _, w := range m.Ports {
			if w.Port == port && w.Value == val {
				return true
			}
		}
		return false
	}
	if !found(0xA0, 0x07) || !found(0xA1, 0x3E) {
		t.Errorf("sem a escrita R#7=3E no PSG: %v", m.Ports)
	}
	if !found(0x99, 0xF4) || !found(0x99, 0x87) {
		t.Errorf("sem a escrita R#7=F4 no VDP: %v", m.Ports)
	}
}

// A rotina destrói IX/IY de propósito; as variáveis locais de uma procedure
// (IX-relativas) têm que sobreviver.
func TestAPIPreservesFrameIXPascal(t *testing.T) {
	evil := libtest.Assemble(t, `MODULE Evil
BANK 0
PUBLIC EvilAdd
EvilAdd:
    ADD HL, DE
    LD IX, 0BEEFh
    RET
ENDMOD
`, "")
	set := libtest.RepoAPI(t)
	if err := set.Merge("func EvilAdd(a: word in HL, b: word in DE): word out HL", "evil.api"); err != nil {
		t.Fatal(err)
	}
	src := `program Ix;
procedure Show(x: Integer);
var
    a, b: Integer;
begin
    a := 11;
    b := EvilAdd(x, 22);
    PrintDec16(a);
    BDOS_PrintChar(44);
    PrintDec16(b);
end;
begin
    Show(9);
end.
`
	obj, asm, err := compilePascal(t, src, set)
	if err != nil {
		t.Fatalf("compilação: %v\n%s", err, asm)
	}
	m := libtest.Machine(libtest.LinkObjects(t, obj, evil))
	if err := m.Run(2_000_000); err != nil {
		t.Fatal(err)
	}
	if got := string(m.Console); got != "11,31" {
		t.Errorf("console = %q, quer %q\n%s", got, "11,31", asm)
	}
}

func TestAPIPascalAliasAndErrors(t *testing.T) {
	set := libtest.RepoAPI(t)
	if err := set.Merge("alias pascal Poke = PSG_Write\nalias basic Doke = PSG_Write", "alias.api"); err != nil {
		t.Fatal(err)
	}
	// alias pascal vale; alias basic não
	if _, asm, err := compilePascal(t, "program A;\nbegin\n    Poke(7, 1);\nend.\n", set); err != nil {
		t.Fatalf("alias pascal: %v\n%s", err, asm)
	}
	cases := []struct{ name, body, want string }{
		{"argumentos a menos", "PSG_Write(7);", "espera 2"},
		{"proc em expressão", "t := PSG_MuteAll();", "não pode ser usada numa expressão"},
	}
	for _, c := range cases {
		src := "program E;\nvar t: Integer;\nbegin\n    " + c.body + "\nend.\n"
		_, _, err := compilePascal(t, src, set)
		if err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: erro = %v, quer conter %q", c.name, err, c.want)
		}
	}
}

func TestAPIUserProcedureWinsPascal(t *testing.T) {
	src := `program U;
procedure Mul16(a: Integer; b: Integer);
begin
    WriteLn('meu');
end;
begin
    Mul16(1, 2);
end.
`
	obj, asm, err := compilePascal(t, src, libtest.RepoAPI(t))
	if err != nil {
		t.Fatalf("%v\n%s", err, asm)
	}
	for _, s := range obj.Symbols {
		if s.Name == "Mul16" && s.Class == mob.SymbolExtern {
			t.Errorf("Mul16 virou EXTERN -- a procedure do usuário deveria vencer:\n%s", asm)
		}
	}
}

// O sample/api/api_demo.pas é o exemplo Pascal que o Wilson testa no hardware.
func TestSampleAPIDemoPascal(t *testing.T) {
	root := libtest.RepoRoot(t)
	data, err := os.ReadFile(filepath.Join(root, "sample", "api", "api_demo.pas"))
	if err != nil {
		t.Fatal(err)
	}
	obj, asm, err := compilePascal(t, string(data), libtest.RepoAPI(t))
	if err != nil {
		t.Fatalf("%v\n%s", err, asm)
	}
	m := libtest.Machine(libtest.LinkObjects(t, obj))
	if err := m.Run(5_000_000); err != nil {
		t.Fatal(err)
	}
	out := string(m.Console)
	for _, want := range []string{"6 * 7 = 42\r\n", "100 / 7 = 14\r\n", "255 em hexa: 0xFF\r\n", "Fim."} {
		if !strings.Contains(out, want) {
			t.Errorf("saída sem %q:\n%q", want, out)
		}
	}
	tone := func(reg, val byte) bool {
		for i := 0; i+1 < len(m.Ports); i++ {
			if m.Ports[i].Port == 0xA0 && m.Ports[i].Value == reg && m.Ports[i+1].Port == 0xA1 && m.Ports[i+1].Value == val {
				return true
			}
		}
		return false
	}
	if !tone(0, 0xFE) || !tone(1, 0x00) || !tone(8, 0x0C) {
		t.Errorf("traço do PSG sem o tom de 440 Hz (R0=FE R1=00 R8=0C): %v", m.Ports)
	}
	if len(m.CALSLTCalls) == 0 {
		t.Error("BIOS_CLS deveria ter feito CALSLT (CHGMOD/INITXT)")
	}
}
