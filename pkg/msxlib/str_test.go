package msxlib_test

import (
	"strings"
	"testing"

	"github.com/wilsonpilon/kizuna/pkg/libtest"
)

// putS escreve s no formato tamanho+dados (s tem no máximo 255 bytes).
func putS(r *libtest.Runner, addr int, s string) uint16 {
	r.M.Mem[addr] = byte(len(s))
	copy(r.M.Mem[addr+1:], s)
	return uint16(addr)
}

func getS(r *libtest.Runner, addr uint16) string {
	n := int(r.M.Mem[addr])
	return string(r.M.Mem[int(addr)+1 : int(addr)+1+n])
}

func strSamples() []string {
	s := append([]string{}, cstrSamples...)
	s = append(s, strings.Repeat("x", 200), strings.Repeat("ab", 127), strings.Repeat("y", 255), "needle in a haystack", "in a", "hay")
	return s
}

func TestStrLenCopyConvert(t *testing.T) {
	r := libtest.NewRunner(t, "STR_Len", "STR_Copy", "STR_FromC", "STR_ToC")
	for _, s := range strSamples() {
		putS(r, cA, s)
		in := canary().WithHL(cA)
		out := r.Call("STR_Len", in)
		if out.A != byte(len(s)) {
			t.Fatalf("Len(%q) = %d", s, out.A)
		}
		keep(t, "Len", in, out, "BCDEHL")

		clear(r.M.Mem[cB : cB+300])
		in = canary().WithHL(cA).WithDE(cB)
		r.Call("STR_Copy", in)
		if getS(r, cB) != s || r.M.Mem[cB+1+len(s)] != 0 {
			t.Fatalf("Copy(%q) = %q", s, getS(r, cB))
		}

		// ToC / FromC
		clear(r.M.Mem[cC : cC+300])
		out = r.Call("STR_ToC", canary().WithHL(cA).WithDE(cC))
		if getC(r, cC) != s || out.DE() != uint16(cC+len(s)) {
			t.Fatalf("ToC(%q) = %q, DE=%04X", s, getC(r, cC), out.DE())
		}
		if strings.ContainsRune(s, 0) {
			continue
		}
		putC(r, cC, s)
		in = canary().WithHL(cC).WithDE(cB)
		out = r.Call("STR_FromC", in)
		if getS(r, cB) != s {
			t.Fatalf("FromC(%q) = %q", s, getS(r, cB))
		}
		keep(t, "FromC", in, out, "DE")
	}
	// FromC trunca em 255
	long := strings.Repeat("z", 400)
	putC(r, cC, long)
	r.Call("STR_FromC", canary().WithHL(cC).WithDE(cB))
	if got := getS(r, cB); got != long[:255] {
		t.Errorf("FromC(400 chars) devolveu %d caracteres, quer 255", len(got))
	}
}

func TestStrCatCompare(t *testing.T) {
	r := libtest.NewRunner(t, "STR_Cat", "STR_Compare")
	samples := strSamples()
	for _, a := range samples {
		for _, b := range samples {
			putS(r, cA, a)
			putS(r, cB, b)
			in := canary().WithHL(cA).WithDE(cB)
			out := r.Call("STR_Compare", in)
			want := sign(strings.Compare(a, b))
			if out.A != want || out.Zero() != (want == 0) {
				t.Fatalf("Compare(%q,%q) = %d Z=%v, quer %d", a, b, out.A, out.Zero(), want)
			}

			// Cat: anexa a ao final de b, truncando em 255
			putS(r, cA, a)
			putS(r, cB, b)
			clear(r.M.Mem[cB+1+len(b) : cB+300]) // o que sobra depois nao pode ser lido
			r.Call("STR_Cat", canary().WithHL(cA).WithDE(cB))
			want2 := b + a
			if len(want2) > 255 {
				want2 = want2[:255]
			}
			if got := getS(r, cB); got != want2 {
				t.Fatalf("Cat(%q, %q) = %q (%d), quer %d caracteres", b, a, got, len(got), len(want2))
			}
		}
	}
}

func TestStrLeftRightMid(t *testing.T) {
	r := libtest.NewRunner(t, "STR_Left", "STR_Right", "STR_Mid")
	min := func(a, b int) int {
		if a < b {
			return a
		}
		return b
	}
	for _, s := range strSamples() {
		for _, n := range []int{0, 1, 2, 3, 5, 100, 254, 255} {
			putS(r, cA, s)
			clear(r.M.Mem[cB : cB+300])
			in := canary().WithHL(cA).WithDE(cB)
			in.A = byte(n)
			r.Call("STR_Left", in)
			if got, want := getS(r, cB), s[:min(n, len(s))]; got != want {
				t.Fatalf("Left(%q, %d) = %q, quer %q", s, n, got, want)
			}
			r.Call("STR_Right", in)
			if got, want := getS(r, cB), s[len(s)-min(n, len(s)):]; got != want {
				t.Fatalf("Right(%q, %d) = %q, quer %q", s, n, got, want)
			}
		}
		for _, start := range []int{0, 1, 2, 3, 5, 10, 200, 255} {
			for _, count := range []int{0, 1, 2, 5, 50, 255} {
				putS(r, cA, s)
				clear(r.M.Mem[cB : cB+300])
				in := canary().WithHL(cA).WithDE(cB)
				in.A, in.B = byte(start), byte(count)
				r.Call("STR_Mid", in)
				st := start
				if st == 0 {
					st = 1
				}
				want := ""
				if st <= len(s) {
					want = s[st-1 : st-1+min(count, len(s)-st+1)]
				}
				if got := getS(r, cB); got != want {
					t.Fatalf("Mid(%q, %d, %d) = %q, quer %q", s, start, count, got, want)
				}
			}
		}
	}
}

func TestStrInStr(t *testing.T) {
	r := libtest.NewRunner(t, "STR_InStr")
	samples := append(strSamples(), "b", "ab", "aab", "aaaab", "ba", "abab", "ababab")
	for _, h := range samples {
		for _, n := range append(samples, "haystack", "a haystack", "hay", "x") {
			putS(r, cA, h)
			putS(r, cB, n)
			in := canary().WithHL(cA).WithDE(cB)
			out := r.Call("STR_InStr", in)
			want := byte(strings.Index(h, n) + 1)
			if n == "" && len(h) == 0 {
				want = 0
			}
			if out.A != want {
				t.Fatalf("InStr(%q, %q) = %d, quer %d", h, n, out.A, want)
			}
			keep(t, "InStr", in, out, "BCDEHL")
		}
	}
	// o frame com IX não pode vazar pilha nem trocar IX do chamador
	r.M.SetIX(0xBEEF)
	putS(r, cA, "hello world")
	putS(r, cB, "o w")
	if out := r.Call("STR_InStr", canary().WithHL(cA).WithDE(cB)); out.A != 5 || r.M.IX() != 0xBEEF {
		t.Errorf("InStr: A=%d IX=%04X", out.A, r.M.IX())
	}
}

func TestStrCaseChrRepeat(t *testing.T) {
	r := libtest.NewRunner(t, "STR_ToUpper", "STR_ToLower", "STR_Chr", "STR_Repeat")
	for _, s := range strSamples() {
		up, low := []byte(s), []byte(s)
		for i, c := range up {
			if c >= 'a' && c <= 'z' {
				up[i] = c - 32
			}
			if c >= 'A' && c <= 'Z' {
				low[i] = c + 32
			}
		}
		for _, x := range []struct {
			fn, want string
		}{{"STR_ToUpper", string(up)}, {"STR_ToLower", string(low)}} {
			putS(r, cA, s)
			in := canary().WithHL(cA)
			out := r.Call(x.fn, in)
			if getS(r, cA) != x.want {
				t.Fatalf("%s(%q) = %q", x.fn, s, getS(r, cA))
			}
			keep(t, x.fn, in, out, "ABCDEHL")
		}
	}
	for _, c := range []byte{'A', 0, 0xFF, ' '} {
		in := canary().WithDE(cB)
		in.A = c
		out := r.Call("STR_Chr", in)
		if getS(r, cB) != string([]byte{c}) {
			t.Fatalf("Chr(%02X) = %q", c, getS(r, cB))
		}
		keep(t, "Chr", in, out, "ABCDEHL")
	}
	for _, n := range []int{0, 1, 5, 200, 255} {
		in := canary().WithDE(cB)
		in.A, in.B = '*', byte(n)
		out := r.Call("STR_Repeat", in)
		if getS(r, cB) != strings.Repeat("*", n) {
			t.Fatalf("Repeat(%d) = %q", n, getS(r, cB))
		}
		keep(t, "Repeat", in, out, "DEHL")
	}
}
