package msxlib_test

import (
	"strings"
	"testing"

	"github.com/wilsonpilon/kizuna/pkg/libtest"
)

// putC escreve s (mais o zero) em addr e devolve addr.
func putC(r *libtest.Runner, addr int, s string) uint16 {
	copy(r.M.Mem[addr:], s)
	r.M.Mem[addr+len(s)] = 0
	return uint16(addr)
}

// getC lê uma string terminada em zero de addr.
func getC(r *libtest.Runner, addr uint16) string {
	end := int(addr)
	for r.M.Mem[end] != 0 {
		end++
	}
	return string(r.M.Mem[addr:end])
}

var cstrSamples = []string{
	"", "a", "A", "abc", "ABC", "abd", "abcd", "Hello, World!", "hello, world!", "  leading", "trailing  ",
	"  both  ", "\t\ttabs\t\n", "   ", "x y z", "aaaa", "0123456789", "MiXeD cAsE 123", "\x7f\x80\xff", "zzz", "Zzz",
}

const (
	cA = 0x9000
	cB = 0x9200
	cC = 0x9400
)

func sign(a int) byte {
	switch {
	case a < 0:
		return 0xFF
	case a > 0:
		return 1
	}
	return 0
}

func TestCStrLenCopyCat(t *testing.T) {
	r := libtest.NewRunner(t, "CSTR_Len", "CSTR_Copy", "CSTR_CopyN", "CSTR_Cat", "CSTR_CatN")
	for _, s := range cstrSamples {
		putC(r, cA, s)
		in := canary().WithHL(cA)
		out := r.Call("CSTR_Len", in)
		if out.HL() != uint16(len(s)) {
			t.Fatalf("Len(%q) = %d", s, out.HL())
		}
		keep(t, "Len", in, out, "BCDE")

		// Copy
		clear(r.M.Mem[cB : cB+300])
		in = canary().WithHL(cA).WithDE(cB)
		out = r.Call("CSTR_Copy", in)
		if getC(r, cB) != s || out.DE() != uint16(cB+len(s)) {
			t.Fatalf("Copy(%q) = %q, DE=%04X", s, getC(r, cB), out.DE())
		}
		keep(t, "Copy", in, out, "BC")

		// CopyN
		for _, n := range []int{0, 1, 2, len(s), len(s) + 5} {
			clear(r.M.Mem[cB : cB+300])
			for i := 0; i < 40; i++ {
				r.M.Mem[cB+i] = 0xEE // lixo: CopyN precisa terminar com zero
			}
			out := r.Call("CSTR_CopyN", canary().WithHL(cA).WithDE(cB).WithBC(uint16(n)))
			want := s
			if n < len(s) {
				want = s[:n]
			}
			if getC(r, cB) != want || out.DE() != uint16(cB+len(want)) {
				t.Fatalf("CopyN(%q, %d) = %q (quer %q), DE=%04X", s, n, getC(r, cB), want, out.DE())
			}
		}

		// Cat / CatN: anexa s ao final de "AB"
		for _, base := range []string{"", "AB", "long base string"} {
			putC(r, cB, base)
			putC(r, cA, s)
			in := canary().WithHL(cA).WithDE(cB)
			out := r.Call("CSTR_Cat", in)
			if getC(r, cB) != base+s || out.DE() != uint16(cB+len(base+s)) {
				t.Fatalf("Cat(%q,%q) = %q", base, s, getC(r, cB))
			}
			keep(t, "Cat", in, out, "BC")
			for _, n := range []int{0, 1, 3, 100} {
				putC(r, cB, base)
				r.Call("CSTR_CatN", canary().WithHL(cA).WithDE(cB).WithBC(uint16(n)))
				w := s
				if n < len(w) {
					w = w[:n]
				}
				if getC(r, cB) != base+w {
					t.Fatalf("CatN(%q,%q,%d) = %q, quer %q", base, s, n, getC(r, cB), base+w)
				}
			}
		}
	}
}

func TestCStrCompare(t *testing.T) {
	r := libtest.NewRunner(t, "CSTR_Compare", "CSTR_CompareN", "CSTR_CompareNoCase")
	for _, a := range cstrSamples {
		for _, b := range cstrSamples {
			putC(r, cA, a)
			putC(r, cB, b)
			in := canary().WithHL(cA).WithDE(cB)

			out := r.Call("CSTR_Compare", in)
			want := sign(strings.Compare(a, b))
			if out.A != want || out.Zero() != (want == 0) {
				t.Fatalf("Compare(%q,%q) = %d Z=%v, quer %d", a, b, out.A, out.Zero(), want)
			}
			keep(t, "Compare", in, out, "BC")

			for _, n := range []int{0, 1, 2, 3, 100} {
				la, lb := a, b
				if n < len(la) {
					la = la[:n]
				}
				if n < len(lb) {
					lb = lb[:n]
				}
				out := r.Call("CSTR_CompareN", canary().WithHL(cA).WithDE(cB).WithBC(uint16(n)))
				if want := sign(strings.Compare(la, lb)); out.A != want {
					t.Fatalf("CompareN(%q,%q,%d) = %d, quer %d", a, b, n, out.A, want)
				}
			}

			out = r.Call("CSTR_CompareNoCase", in)
			upper := func(s string) string { // só ASCII, como a rotina
				bs := []byte(s)
				for i, c := range bs {
					if c >= 'a' && c <= 'z' {
						bs[i] = c - 32
					}
				}
				return string(bs)
			}
			want = sign(strings.Compare(upper(a), upper(b)))
			if out.A != want || out.Zero() != (want == 0) {
				t.Fatalf("CompareNoCase(%q,%q) = %d, quer %d", a, b, out.A, want)
			}
			keep(t, "CompareNoCase", in, out, "BC")
		}
	}
}

func TestCStrFind(t *testing.T) {
	r := libtest.NewRunner(t, "CSTR_FindChar", "CSTR_FindLastChar", "CSTR_FindStr")
	for _, s := range cstrSamples {
		putC(r, cA, s)
		for _, ch := range []byte{'a', 'A', 'l', ' ', '0', 'z', 0, 0xFF, 'x'} {
			in := canary().WithHL(cA)
			in.A = ch

			out := r.Call("CSTR_FindChar", in)
			idx := strings.IndexByte(s+"\x00", ch)
			if ch != 0 && idx == len(s) {
				idx = -1
			}
			checkFind(t, "FindChar", s, ch, out, idx)
			keep(t, "FindChar", in, out, "BC")

			out = r.Call("CSTR_FindLastChar", in)
			idx = strings.LastIndexByte(s, ch)
			if ch == 0 {
				idx = len(s)
			}
			checkFind(t, "FindLastChar", s, ch, out, idx)
			keep(t, "FindLastChar", in, out, "BCDE")
		}
		for _, needle := range []string{"", "a", "abc", "cd", "o, W", "  ", "zzzz", "z", "\x80", s} {
			putC(r, cB, needle)
			in := canary().WithHL(cA).WithDE(cB)
			out := r.Call("CSTR_FindStr", in)
			idx := strings.Index(s, needle)
			if idx < 0 {
				if out.HL() != 0 || out.Zero() {
					t.Fatalf("FindStr(%q,%q): HL=%04X Z=%v, quer 0/NZ", s, needle, out.HL(), out.Zero())
				}
			} else if out.HL() != uint16(cA+idx) || !out.Zero() {
				t.Fatalf("FindStr(%q,%q): HL=%04X Z=%v, quer %04X", s, needle, out.HL(), out.Zero(), cA+idx)
			}
			keep(t, "FindStr", in, out, "BCDE")
		}
	}
}

func checkFind(t *testing.T, what, s string, ch byte, out libtest.Regs, idx int) {
	t.Helper()
	if idx < 0 {
		if out.HL() != 0 || out.Zero() {
			t.Fatalf("%s(%q,%#02x): HL=%04X Z=%v, quer 0/NZ", what, s, ch, out.HL(), out.Zero())
		}
		return
	}
	if out.HL() != uint16(cA+idx) || !out.Zero() {
		t.Fatalf("%s(%q,%#02x): HL=%04X Z=%v, quer %04X", what, s, ch, out.HL(), out.Zero(), cA+idx)
	}
}

func TestCStrInPlace(t *testing.T) {
	r := libtest.NewRunner(t, "CSTR_ToUpper", "CSTR_ToLower", "CSTR_Reverse", "CSTR_TrimLeft", "CSTR_TrimRight", "CSTR_ReplaceChar")
	isSpace := func(c byte) bool { return c == ' ' || (c >= 9 && c <= 13) }
	for _, s := range cstrSamples {
		in := canary().WithHL(cA)
		check := func(fn, want string, keepRegs string, in libtest.Regs) {
			t.Helper()
			putC(r, cA, s)
			out := r.Call(fn, in)
			if got := getC(r, cA); got != want {
				t.Fatalf("%s(%q) = %q, quer %q", fn, s, got, want)
			}
			keep(t, fn, in, out, keepRegs)
		}
		up, low := []byte(s), []byte(s)
		for i, c := range up {
			if c >= 'a' && c <= 'z' {
				up[i] = c - 32
			}
			if c >= 'A' && c <= 'Z' {
				low[i] = c + 32
			}
		}
		check("CSTR_ToUpper", string(up), "BCDEHL", in)
		check("CSTR_ToLower", string(low), "BCDEHL", in)

		rev := []byte(s)
		for i, j := 0, len(rev)-1; i < j; i, j = i+1, j-1 {
			rev[i], rev[j] = rev[j], rev[i]
		}
		check("CSTR_Reverse", string(rev), "ABCDEHL", in)

		left := s
		for len(left) > 0 && isSpace(left[0]) {
			left = left[1:]
		}
		check("CSTR_TrimLeft", left, "ABCDEHL", in)
		right := s
		for len(right) > 0 && isSpace(right[len(right)-1]) {
			right = right[:len(right)-1]
		}
		check("CSTR_TrimRight", right, "ABCDEHL", in)

		rin := canary().WithHL(cA).WithDE(uint16('a')<<8 | '#')
		check("CSTR_ReplaceChar", strings.ReplaceAll(s, "a", "#"), "ABCDEHL", rin)
	}
}
