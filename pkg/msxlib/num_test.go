package msxlib_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/wilsonpilon/kizuna/pkg/libtest"
)

func TestNumToText(t *testing.T) {
	r := libtest.NewRunner(t, "NUM_U16ToCStr", "NUM_I16ToCStr", "NUM_U8ToHex", "NUM_U16ToHex", "NUM_U8ToBin", "NUM_U16ToBin")
	for v := 0; v < 65536; v++ {
		clear(r.M.Mem[cA : cA+24])
		in := canary().WithHL(uint16(v)).WithDE(cA)
		out := r.Call("NUM_U16ToCStr", in)
		if want := strconv.Itoa(v); getC(r, cA) != want || out.DE() != uint16(cA+len(want)) {
			t.Fatalf("U16ToCStr(%d) = %q DE=%04X", v, getC(r, cA), out.DE())
		}
		keep(t, "U16ToCStr", in, out, "BC")

		out = r.Call("NUM_I16ToCStr", in)
		if want := strconv.Itoa(int(int16(v))); getC(r, cA) != want || out.DE() != uint16(cA+len(want)) {
			t.Fatalf("I16ToCStr(%d) = %q", int16(v), getC(r, cA))
		}
		keep(t, "I16ToCStr", in, out, "BC")

		out = r.Call("NUM_U16ToHex", in)
		if want := strings.ToUpper(strconv.FormatUint(uint64(v)|0x10000, 16)[1:]); getC(r, cA) != want || out.DE() != cA+4 {
			t.Fatalf("U16ToHex(%d) = %q", v, getC(r, cA))
		}
		keep(t, "U16ToHex", in, out, "BCHL")

		if v%251 == 0 || v < 300 || v > 65300 {
			out = r.Call("NUM_U16ToBin", in)
			if want := strconv.FormatUint(uint64(v)|0x10000, 2)[1:]; getC(r, cA) != want || out.DE() != cA+16 {
				t.Fatalf("U16ToBin(%d) = %q", v, getC(r, cA))
			}
			keep(t, "U16ToBin", in, out, "BCHL")
		}
	}
	for v := 0; v < 256; v++ {
		in := canary().WithDE(cA)
		in.A = byte(v)
		out := r.Call("NUM_U8ToHex", in)
		if want := strings.ToUpper(strconv.FormatUint(uint64(v)|0x100, 16)[1:]); getC(r, cA) != want || out.DE() != cA+2 {
			t.Fatalf("U8ToHex(%d) = %q", v, getC(r, cA))
		}
		keep(t, "U8ToHex", in, out, "BCHL")
		out = r.Call("NUM_U8ToBin", in)
		if want := strconv.FormatUint(uint64(v)|0x100, 2)[1:]; getC(r, cA) != want || out.DE() != cA+8 {
			t.Fatalf("U8ToBin(%d) = %q", v, getC(r, cA))
		}
		keep(t, "U8ToBin", in, out, "BCHL")
	}
}

// scanRef é a referência em Go do leitor decimal: pula espaços/TABs, sinal
// opcional, dígitos. Devolve valor, quantos caracteres consumiu e se deu erro.
func scanRef(s string, signed bool) (int, int, bool) {
	i := 0
	for i < len(s) && (s[i] == ' ' || s[i] == '\t') {
		i++
	}
	neg := false
	if i < len(s) && s[i] == '+' {
		i++
	} else if signed && i < len(s) && s[i] == '-' {
		neg = true
		i++
	}
	start := i
	v := 0
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		v = v*10 + int(s[i]-'0')
		if v > 65535 {
			return 0, i, true
		}
		i++
	}
	if i == start {
		return 0, i, true
	}
	if signed {
		if neg && v > 32768 || !neg && v > 32767 {
			return 0, i, true
		}
		if neg {
			v = -v
		}
	}
	return v, i, false
}

// hexRef é a referência do leitor hexadecimal (prefixos $, &H e 0x opcionais).
func hexRef(s string) (int, int, bool) {
	i := 0
	for i < len(s) && (s[i] == ' ' || s[i] == '\t') {
		i++
	}
	switch {
	case i < len(s) && s[i] == '$':
		i++
	case i+1 < len(s) && s[i] == '&' && (s[i+1] == 'H' || s[i+1] == 'h'):
		i += 2
	case i+1 < len(s) && s[i] == '0' && (s[i+1] == 'x' || s[i+1] == 'X'):
		i += 2
	}
	start := i
	v := 0
	for i < len(s) {
		var d int
		c := s[i]
		switch {
		case c >= '0' && c <= '9':
			d = int(c - '0')
		case c|0x20 >= 'a' && c|0x20 <= 'f':
			d = int(c|0x20-'a') + 10
		default:
			d = -1
		}
		if d < 0 {
			break
		}
		if v > 0xFFF {
			return 0, i, true
		}
		v = v<<4 | d
		i++
	}
	if i == start {
		return 0, i, true
	}
	return v, i, false
}

var numTexts = []string{
	"0", "7", "42", "65535", "65536", "99999", "100000", "32767", "32768", "-32768", "-32769", "-1", "-0", "+5", "+-5",
	"  12", "\t\t34", "12abc", "12 34", "abc", "", "   ", "-", "+", "- 5", "00000000000012", "0000000000000000000099999",
	"1,2", "9", "007", "65535x", "6553", "12345678901234567890",
	"FF", "ff", "0xFF", "0XfF", "&H1F", "&h1f", "$c0DE", "12345", "0x", "0", "$", "&H", "&", "G", "FFFF", "10000", "1G", "0x 1",
}

func TestNumTextToNum(t *testing.T) {
	r := libtest.NewRunner(t, "NUM_CStrToU16", "NUM_CStrToI16", "NUM_CStrToHex16", "STR_Val", "STR_HexVal")
	for _, s := range numTexts {
		for _, fn := range []string{"NUM_CStrToU16", "NUM_CStrToI16", "NUM_CStrToHex16"} {
			var v, used int
			var bad bool
			switch fn {
			case "NUM_CStrToU16":
				v, used, bad = scanRef(s, false)
			case "NUM_CStrToI16":
				v, used, bad = scanRef(s, true)
			default:
				v, used, bad = hexRef(s)
			}
			putC(r, cA, s)
			in := canary().WithHL(cA)
			out := r.Call(fn, in)
			if out.Carry() != bad {
				t.Fatalf("%s(%q): carry=%v, quer %v", fn, s, out.Carry(), bad)
			}
			if bad {
				if out.HL() != 0 {
					t.Fatalf("%s(%q): erro mas HL=%d", fn, s, out.HL())
				}
			} else {
				if out.HL() != uint16(v) {
					t.Fatalf("%s(%q) = %d, quer %d", fn, s, int16(out.HL()), v)
				}
				if out.DE() != uint16(cA+used) {
					t.Fatalf("%s(%q): DE=%04X, quer %04X (consumiu %d)", fn, s, out.DE(), cA+used, used)
				}
			}
			keep(t, fn, in, out, "BC")
		}

		// STR_Val / STR_HexVal: a string tamanho+dados NAO tem terminador; o que vem
		// depois dela na memoria (aqui, digitos) nao pode ser lido.
		if len(s) > 255 {
			continue
		}
		for _, fn := range []string{"STR_Val", "STR_HexVal"} {
			var v int
			var bad bool
			if fn == "STR_Val" {
				v, _, bad = scanRef(s, true)
			} else {
				v, _, bad = hexRef(s)
			}
			// o lixo depois da string tenta enganar o leitor: digitos, sinais e prefixos
			for _, garbage := range []string{"1234567890abcdef", "+99", "-99", "$FF", "0x7F", "&H7F", "5"} {
				putS(r, cA, s)
				clear(r.M.Mem[cA+1+len(s) : cA+300]) // sem restos de casos anteriores mascarando o lixo
				copy(r.M.Mem[cA+1+len(s):], garbage)
				in := canary().WithHL(cA)
				out := r.Call(fn, in)
				if out.Carry() != bad || (!bad && out.HL() != uint16(v)) || (bad && out.HL() != 0) {
					t.Fatalf("%s(%q) [lixo depois: %q] = %d carry=%v, quer %d carry=%v", fn, s, garbage, int16(out.HL()), out.Carry(), v, bad)
				}
				keep(t, fn, in, out, "BCDE")
			}
		}
	}
}

func TestStrFromNum(t *testing.T) {
	r := libtest.NewRunner(t, "STR_FromU16", "STR_FromI16", "STR_Hex8", "STR_Hex16")
	for v := 0; v < 65536; v += 1 + v/40 {
		in := canary().WithHL(uint16(v)).WithDE(cB)
		out := r.Call("STR_FromU16", in)
		if getS(r, cB) != strconv.Itoa(v) {
			t.Fatalf("FromU16(%d) = %q", v, getS(r, cB))
		}
		keep(t, "FromU16", in, out, "BC")
		r.Call("STR_FromI16", in)
		if getS(r, cB) != strconv.Itoa(int(int16(v))) {
			t.Fatalf("FromI16(%d) = %q", int16(v), getS(r, cB))
		}
		r.Call("STR_Hex16", in)
		if want := strings.ToUpper(strconv.FormatUint(uint64(v)|0x10000, 16)[1:]); getS(r, cB) != want {
			t.Fatalf("Hex16(%d) = %q", v, getS(r, cB))
		}
	}
	for v := 0; v < 256; v++ {
		in := canary().WithDE(cB)
		in.A = byte(v)
		r.Call("STR_Hex8", in)
		if want := strings.ToUpper(strconv.FormatUint(uint64(v)|0x100, 16)[1:]); getS(r, cB) != want {
			t.Fatalf("Hex8(%d) = %q", v, getS(r, cB))
		}
	}
	for _, v := range []uint16{0, 1, 9, 10, 99, 100, 32767, 32768, 65535} {
		r.Call("STR_FromI16", canary().WithHL(v).WithDE(cB))
		if getS(r, cB) != strconv.Itoa(int(int16(v))) {
			t.Fatalf("FromI16(%#x) = %q", v, getS(r, cB))
		}
	}
}

func TestNumDecW(t *testing.T) {
	r := libtest.NewRunner(t, "NUM_U16ToDecW")
	for _, v := range []int{0, 1, 5, 9, 10, 42, 99, 100, 999, 1234, 9999, 10000, 54321, 65535} {
		for w := 0; w <= 8; w++ {
			for _, pad := range []byte{' ', '0', '*'} {
				clear(r.M.Mem[cA : cA+24])
				in := canary().WithHL(uint16(v)).WithDE(cA)
				in.B, in.A = byte(w), pad
				out := r.Call("NUM_U16ToDecW", in)
				d := strconv.Itoa(v)
				want := d
				if len(d) < w {
					want = strings.Repeat(string(pad), w-len(d)) + d
				} else {
					want = d[len(d)-w:]
				}
				if got := getC(r, cA); got != want || out.DE() != uint16(cA+len(want)) {
					t.Fatalf("U16ToDecW(%d, largura %d, pad %q) = %q (DE=%04X), quer %q", v, w, pad, getC(r, cA), out.DE(), want)
				}
				keep(t, "U16ToDecW", in, out, "")
			}
		}
	}
}
