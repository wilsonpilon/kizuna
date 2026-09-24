package msxlib_test

import (
	"testing"

	"github.com/wilsonpilon/kizuna/pkg/libtest"
)

func TestConOutput(t *testing.T) {
	r := libtest.NewRunner(t, "CON_PrintCStr", "CON_NewLine", "CON_PrintLine", "CON_PrintI16", "CON_Cls", "CON_Locate")
	call := func(fn string, in libtest.Regs) string {
		t.Helper()
		r.M.Console = nil
		out := r.Call(fn, in)
		keep(t, fn, in, out, "ABCDEHL")
		return string(r.M.Console)
	}
	putC(r, cA, "Ola, MSX!")
	if got := call("CON_PrintCStr", canary().WithHL(cA)); got != "Ola, MSX!" {
		t.Errorf("PrintCStr = %q", got)
	}
	putC(r, cA, "")
	if got := call("CON_PrintCStr", canary().WithHL(cA)); got != "" {
		t.Errorf("PrintCStr vazia = %q", got)
	}
	if got := call("CON_NewLine", canary()); got != "\r\n" {
		t.Errorf("NewLine = %q", got)
	}
	putC(r, cA, "linha")
	if got := call("CON_PrintLine", canary().WithHL(cA)); got != "linha\r\n" {
		t.Errorf("PrintLine = %q", got)
	}
	for _, c := range []struct {
		v    uint16
		want string
	}{{0, "0"}, {7, "7"}, {12345, "12345"}, {32767, "32767"}, {0xFFFF, "-1"}, {0x8000, "-32768"}, {0xFF38, "-200"}} {
		if got := call("CON_PrintI16", canary().WithHL(c.v)); got != c.want {
			t.Errorf("PrintI16(%#04x) = %q, quer %q", c.v, got, c.want)
		}
	}
	if got := call("CON_Cls", canary()); got != "\f" {
		t.Errorf("Cls = %q", got)
	}
	in := canary()
	in.A, in.B = 5, 3 // coluna 5, linha 3
	if got := call("CON_Locate", in); got != "\x1bY\x23\x25" {
		t.Errorf("Locate(5,3) = %q", got)
	}
}

func TestConInput(t *testing.T) {
	r := libtest.NewRunner(t, "CON_ReadLine", "CON_ReadKey", "CON_KeyPressed")

	// tecla / teste de tecla
	if out := r.Call("CON_KeyPressed", canary()); out.A != 0 || !out.Zero() {
		t.Errorf("KeyPressed sem tecla: A=%d Z=%v", out.A, out.Zero())
	}
	r.M.Input = []byte("xy")
	in := canary()
	out := r.Call("CON_KeyPressed", in)
	if out.A != 1 || out.Zero() {
		t.Errorf("KeyPressed com tecla: A=%d Z=%v", out.A, out.Zero())
	}
	keep(t, "KeyPressed", in, out, "BCDEHL")
	out = r.Call("CON_ReadKey", in)
	if out.A != 'x' {
		t.Errorf("ReadKey = %q", out.A)
	}
	keep(t, "ReadKey", in, out, "BCDEHL")
	r.Call("CON_ReadKey", in)
	if out := r.Call("CON_KeyPressed", in); out.A != 0 {
		t.Errorf("KeyPressed depois de consumir tudo: A=%d", out.A)
	}

	// linha
	for _, c := range []struct {
		input string
		max   byte
		want  string
	}{{"hello\r\n", 20, "hello"}, {"hello\r", 20, "hello"}, {"hello\n", 20, "hello"}, {"\r\n", 20, ""}, {"abcdefghij\r\n", 4, "abcd"}, {"tail\r\nrest", 10, "tail"}} {
		clear(r.M.Mem[cA : cA+300])
		r.M.Input = []byte(c.input)
		in := canary().WithDE(cA)
		in.A = c.max
		out := r.Call("CON_ReadLine", in)
		if out.HL() != cA+1 || getS(r, out.HL()) != c.want {
			t.Errorf("ReadLine(%q, max %d): HL=%04X %q, quer %q", c.input, c.max, out.HL(), getS(r, out.HL()), c.want)
		}
		if r.M.Mem[cA] != c.max {
			t.Errorf("byte de tamanho maximo do buffer = %d", r.M.Mem[cA])
		}
		keep(t, "ReadLine", in, out, "BCDE")
	}
	// a fila continua com o que sobrou depois do ENTER
	r.M.Input = []byte("tail\r\nrest")
	in = canary().WithDE(cA)
	in.A = 10
	r.Call("CON_ReadLine", in)
	if string(r.M.Input) != "rest" {
		t.Errorf("fila depois da linha = %q, quer %q", r.M.Input, "rest")
	}
}
