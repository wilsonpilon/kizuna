package kaji80

import (
	"strings"
	"testing"
)

func asmBody(t *testing.T, body string) ([]byte, error) {
	t.Helper()
	src := "MODULE X\nBANK 0\nPUBLIC S\nFOO EQU 6\nS:\n" + body + "\n"
	obj, err := NewAssembler().Assemble(src)
	if err != nil {
		return nil, err
	}
	return obj.Segments[0].Data, nil
}

// Regressão: imediatos de 8 bits com expressão (CP FOO+1, OUT (FOO+1),A...)
// eram lidos por Sscanf só até o primeiro caractere inválido e viravam 0 em
// silêncio (CP FOO+1 -> CP 0).
func TestImm8Expressions(t *testing.T) {
	cases := []struct {
		src  string
		want []byte
	}{
		{"CP FOO + 1", []byte{0xFE, 0x07}},
		{"LD A, FOO*2", []byte{0x3E, 0x0C}},
		{"LD A, 2*3", []byte{0x3E, 0x06}},
		{"CP 5+1", []byte{0xFE, 0x06}},
		{"BIT 1+1, A", []byte{0xCB, 0x57}},
		{"OUT (FOO+1), A", []byte{0xD3, 0x07}},
		{"IN A, (FOO+1)", []byte{0xDB, 0x07}},
		{"LD B, 1<<4", []byte{0x06, 0x10}},
		{"LD (HL), FOO-1", []byte{0x36, 0x05}},
		{"LD (IX+2), FOO+1", []byte{0xDD, 0x36, 0x02, 0x07}},
		{"DB FOO+1, 2*2", []byte{0x07, 0x04}},
		{"LD A, 'A'", []byte{0x3E, 0x41}},
		{"LD A, 'A'+1", []byte{0x3E, 0x42}},
		{"LD A, F0h", []byte{0x3E, 0xF0}},
		{"LD A, 0F0h", []byte{0x3E, 0xF0}},
		{"LD A, $F0", []byte{0x3E, 0xF0}},
		{"LD A, -1", []byte{0x3E, 0xFF}},
		{"LD HL, FOO+1", []byte{0x21, 0x07, 0x00}},
		{"LD DE, 2*3", []byte{0x11, 0x06, 0x00}},
		{"LD A, (FOO+1)", []byte{0x3A, 0x07, 0x00}},
		{"JP FOO+1", []byte{0xC3, 0x07, 0x00}},
		{"LD C, F_OPEN", []byte{0x0E, 0x43}},
	}
	for _, c := range cases {
		got, err := asmBody(t, c.src)
		if err != nil {
			t.Errorf("%q: %v", c.src, err)
			continue
		}
		if string(got) != string(c.want) {
			t.Errorf("%q: % X, quer % X", c.src, got, c.want)
		}
	}
}

// Nada disso pode virar 0 em silêncio: tem que ser erro.
func TestImm8Errors(t *testing.T) {
	bad := []string{
		"LD A, UnknownName",
		"CP UnknownName+1",
		"LD A, 300",
		"LD A, 2*",
		"OUT (UnknownName), A",
		"LD A, FOO+",
		"BIT 8, A",
	}
	for _, src := range bad {
		if _, err := asmBody(t, src); err == nil {
			t.Errorf("%q deveria dar erro", src)
		}
	}
	if _, err := asmBody(t, "LD A, Lbl+1\nLbl:"); err == nil || !strings.Contains(err.Error(), "Lbl") {
		t.Errorf("rótulo dentro de imediato: err = %v", err)
	}
}

// DS lia a contagem com Sscanf("%d"): "DS 0FFh" e "DS NOME" reservavam 0 bytes.
func TestDSCountExpressions(t *testing.T) {
	cases := []struct {
		src  string
		n    int
		fill byte
	}{
		{"DS 4", 4, 0}, {"DS 0Ah", 10, 0}, {"DS FOO", 6, 0}, {"DS FOO*2+1", 13, 0},
		{"DS 3, 0FFh", 3, 0xFF}, {"DS 0", 0, 0},
	}
	for _, c := range cases {
		got, err := asmBody(t, c.src)
		if err != nil {
			t.Errorf("%q: %v", c.src, err)
			continue
		}
		if len(got) != c.n {
			t.Errorf("%q: %d bytes, quer %d", c.src, len(got), c.n)
		}
		for _, b := range got {
			if b != c.fill {
				t.Errorf("%q: byte %02X, quer %02X", c.src, b, c.fill)
				break
			}
		}
	}
	for _, src := range []string{"DS Unknown", "DS", "DS -1"} {
		if _, err := asmBody(t, src); err == nil {
			t.Errorf("%q deveria dar erro", src)
		}
	}
}
