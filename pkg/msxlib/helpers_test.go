package msxlib_test

import (
	"math/rand"
	"strings"
	"testing"

	"github.com/wilsonpilon/kizuna/pkg/libtest"
)

// edge16 são valores de borda de 16 bits usados em todos os testes de faixa.
var edge16 = []uint16{
	0, 1, 2, 3, 9, 10, 11, 99, 100, 127, 128, 255, 256, 257, 1000, 4095, 4096,
	0x7FFE, 0x7FFF, 0x8000, 0x8001, 0xFF00, 0xFFFE, 0xFFFF,
}

// pairs16 devolve todos os pares das bordas mais n pares pseudoaleatórios
// (semente fixa: o teste é determinístico).
func pairs16(n int) [][2]uint16 {
	var out [][2]uint16
	for _, a := range edge16 {
		for _, b := range edge16 {
			out = append(out, [2]uint16{a, b})
		}
	}
	r := rand.New(rand.NewSource(1))
	for i := 0; i < n; i++ {
		out = append(out, [2]uint16{uint16(r.Intn(65536)), uint16(r.Intn(65536))})
	}
	return out
}

// keep confere que os registradores listados (letras A B C D E H L) saíram
// iguais aos de entrada.
func keep(t *testing.T, what string, in, out libtest.Regs, regs string) {
	t.Helper()
	get := func(r libtest.Regs, c rune) byte {
		switch c {
		case 'A':
			return r.A
		case 'B':
			return r.B
		case 'C':
			return r.C
		case 'D':
			return r.D
		case 'E':
			return r.E
		case 'H':
			return r.H
		case 'L':
			return r.L
		}
		t.Fatalf("registrador %q desconhecido", c)
		return 0
	}
	for _, c := range regs {
		if get(in, c) != get(out, c) {
			t.Fatalf("%s: registrador %c mudou (%02X -> %02X), devia ser preservado", what, c, get(in, c), get(out, c))
		}
	}
}

// canary devolve registradores de entrada com valores "marcadores" distintos,
// para a checagem de preservação ser significativa.
func canary() libtest.Regs {
	return libtest.Regs{A: 0xA5, B: 0xB1, C: 0xC2, D: 0xD3, E: 0xE4, H: 0x11, L: 0x22}
}

func names(list ...string) []string { return list }

func join(list []string) string { return strings.Join(list, ", ") }
