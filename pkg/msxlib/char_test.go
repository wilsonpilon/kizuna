package msxlib_test

import (
	"testing"

	"github.com/wilsonpilon/kizuna/pkg/libtest"
)

func TestCharPredicates(t *testing.T) {
	isDigit := func(c int) bool { return c >= '0' && c <= '9' }
	isUpper := func(c int) bool { return c >= 'A' && c <= 'Z' }
	isLower := func(c int) bool { return c >= 'a' && c <= 'z' }
	isAlpha := func(c int) bool { return isUpper(c) || isLower(c) }
	preds := map[string]func(int) bool{
		"CHAR_IsDigit":    isDigit,
		"CHAR_IsUpper":    isUpper,
		"CHAR_IsLower":    isLower,
		"CHAR_IsAlpha":    isAlpha,
		"CHAR_IsAlNum":    func(c int) bool { return isAlpha(c) || isDigit(c) },
		"CHAR_IsHexDigit": func(c int) bool { return isDigit(c) || (c >= 'A' && c <= 'F') || (c >= 'a' && c <= 'f') },
		"CHAR_IsSpace":    func(c int) bool { return c == ' ' || (c >= 9 && c <= 13) },
		"CHAR_IsControl":  func(c int) bool { return c < 32 || c == 127 },
		"CHAR_IsAscii":    func(c int) bool { return c < 128 },
		"CHAR_IsPrint":    func(c int) bool { return c >= 32 && c <= 126 },
		"CHAR_IsGraph":    func(c int) bool { return c >= 33 && c <= 126 },
		"CHAR_IsPunct": func(c int) bool {
			return c >= 33 && c <= 126 && !isAlpha(c) && !isDigit(c)
		},
	}
	var list []string
	for n := range preds {
		list = append(list, n)
	}
	r := libtest.NewRunner(t, list...)
	for name, want := range preds {
		for c := 0; c < 256; c++ {
			in := canary()
			in.A = byte(c)
			out := r.Call(name, in)
			w := byte(0)
			if want(c) {
				w = 1
			}
			if out.A != w || out.Zero() != (w == 0) {
				t.Fatalf("%s(%#02x) = A %d Z %v, quer A %d Z %v", name, c, out.A, out.Zero(), w, w == 0)
			}
			keep(t, name, in, out, "BCDEHL")
		}
	}
}

func TestCharConversions(t *testing.T) {
	r := libtest.NewRunner(t, "CHAR_ToUpper", "CHAR_ToLower", "CHAR_DigitValue", "CHAR_HexChar")
	for c := 0; c < 256; c++ {
		in := canary()
		in.A = byte(c)

		wantUp, wantLow := byte(c), byte(c)
		if c >= 'a' && c <= 'z' {
			wantUp = byte(c - 32)
		}
		if c >= 'A' && c <= 'Z' {
			wantLow = byte(c + 32)
		}
		if out := r.Call("CHAR_ToUpper", in); out.A != wantUp {
			t.Fatalf("ToUpper(%#02x) = %#02x", c, out.A)
		} else {
			keep(t, "ToUpper", in, out, "BCDEHL")
		}
		if out := r.Call("CHAR_ToLower", in); out.A != wantLow {
			t.Fatalf("ToLower(%#02x) = %#02x", c, out.A)
		} else {
			keep(t, "ToLower", in, out, "BCDEHL")
		}

		wantDig := byte(0xFF)
		switch {
		case c >= '0' && c <= '9':
			wantDig = byte(c - '0')
		case c >= 'A' && c <= 'F':
			wantDig = byte(c - 'A' + 10)
		case c >= 'a' && c <= 'f':
			wantDig = byte(c - 'a' + 10)
		}
		if out := r.Call("CHAR_DigitValue", in); out.A != wantDig {
			t.Fatalf("DigitValue(%#02x) = %#02x, quer %#02x", c, out.A, wantDig)
		} else {
			keep(t, "DigitValue", in, out, "BCDEHL")
		}

		if out := r.Call("CHAR_HexChar", in); out.A != "0123456789ABCDEF"[c&15] {
			t.Fatalf("HexChar(%#02x) = %q", c, out.A)
		} else {
			keep(t, "HexChar", in, out, "BCDEHL")
		}
	}
}
