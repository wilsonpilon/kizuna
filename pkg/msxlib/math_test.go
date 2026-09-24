package msxlib_test

import (
	"math"
	"math/rand"
	"testing"

	"github.com/wilsonpilon/kizuna/pkg/libtest"
)

// --- unários de 16 bits -------------------------------------------------------

func TestMathNegAbsSign16(t *testing.T) {
	r := libtest.NewRunner(t, "MATH_Neg16", "MATH_Abs16", "MATH_Sign16")
	for v := 0; v < 65536; v++ {
		in := canary().WithHL(uint16(v))
		s := int16(v)

		out := r.Call("MATH_Neg16", in)
		if out.HL() != uint16(-s) {
			t.Fatalf("Neg16(%d) = %d", s, int16(out.HL()))
		}
		keep(t, "Neg16", in, out, "ABCDE")

		out = r.Call("MATH_Abs16", in)
		wantAbs := s
		if s < 0 {
			wantAbs = -s // -32768 continua -32768
		}
		if out.HL() != uint16(wantAbs) {
			t.Fatalf("Abs16(%d) = %d", s, int16(out.HL()))
		}
		keep(t, "Abs16", in, out, "ABCDE")

		out = r.Call("MATH_Sign16", in)
		want := byte(0)
		if s > 0 {
			want = 1
		} else if s < 0 {
			want = 0xFF
		}
		if out.A != want {
			t.Fatalf("Sign16(%d) = %02X, quer %02X", s, out.A, want)
		}
		keep(t, "Sign16", in, out, "BCDEHL")
	}
}

// --- comparação, mínimo, máximo, clamp ---------------------------------------------

func TestMathCmpMinMax(t *testing.T) {
	r := libtest.NewRunner(t, "MATH_Cmp16U", "MATH_Cmp16S", "MATH_Min16U", "MATH_Max16U", "MATH_Min16S", "MATH_Max16S")
	for _, p := range pairs16(4000) {
		a, b := p[0], p[1]
		in := canary().WithHL(a).WithDE(b)

		out := r.Call("MATH_Cmp16U", in)
		if out.Carry() != (a < b) || out.Zero() != (a == b) {
			t.Fatalf("Cmp16U(%d,%d): C=%v Z=%v", a, b, out.Carry(), out.Zero())
		}
		keep(t, "Cmp16U", in, out, "ABCDEHL")

		out = r.Call("MATH_Cmp16S", in)
		sa, sb := int16(a), int16(b)
		if out.Carry() != (sa < sb) || out.Zero() != (sa == sb) {
			t.Fatalf("Cmp16S(%d,%d): C=%v Z=%v", sa, sb, out.Carry(), out.Zero())
		}
		keep(t, "Cmp16S", in, out, "ABCDEHL")

		type c struct {
			name string
			want uint16
		}
		wantMinU, wantMaxU := a, b
		if b < a {
			wantMinU, wantMaxU = b, a
		}
		wantMinS, wantMaxS := a, b
		if sb < sa {
			wantMinS, wantMaxS = b, a
		}
		for _, x := range []c{{"MATH_Min16U", wantMinU}, {"MATH_Max16U", wantMaxU}, {"MATH_Min16S", wantMinS}, {"MATH_Max16S", wantMaxS}} {
			out = r.Call(x.name, in)
			if out.HL() != x.want {
				t.Fatalf("%s(%d,%d) = %d, quer %d", x.name, a, b, out.HL(), x.want)
			}
			keep(t, x.name, in, out, "ABCDE")
		}
	}
}

func TestMathClamp16S(t *testing.T) {
	r := libtest.NewRunner(t, "MATH_Clamp16S")
	rng := rand.New(rand.NewSource(2))
	vals := append([]uint16{}, edge16...)
	for i := 0; i < 60; i++ {
		vals = append(vals, uint16(rng.Intn(65536)))
	}
	for _, v := range vals {
		for _, lo := range vals[:30] {
			for _, hi := range vals[:30] {
				in := canary().WithHL(v).WithDE(lo).WithBC(hi)
				out := r.Call("MATH_Clamp16S", in)
				sv, slo, shi := int16(v), int16(lo), int16(hi)
				want := sv
				if want < slo {
					want = slo
				}
				if want > shi {
					want = shi
				}
				if out.HL() != uint16(want) {
					t.Fatalf("Clamp16S(%d, min %d, max %d) = %d, quer %d", sv, slo, shi, int16(out.HL()), want)
				}
				keep(t, "Clamp16S", in, out, "ABCDE")
			}
		}
	}
}

// --- multiplicação -------------------------------------------------------------------

func TestMathMul8(t *testing.T) {
	r := libtest.NewRunner(t, "MATH_Mul8")
	for a := 0; a < 256; a++ {
		for e := 0; e < 256; e++ {
			in := canary()
			in.A, in.E = byte(a), byte(e)
			out := r.Call("MATH_Mul8", in)
			if out.HL() != uint16(a*e) {
				t.Fatalf("Mul8(%d,%d) = %d", a, e, out.HL())
			}
			keep(t, "Mul8", in, out, "BCDE")
		}
	}
}

func TestMathMul32(t *testing.T) {
	r := libtest.NewRunner(t, "MATH_MulU16x16", "MATH_MulS16x16", "MATH_Neg32")
	for _, p := range pairs16(3000) {
		a, b := p[0], p[1]
		in := canary().WithHL(a).WithDE(b)

		out := r.Call("MATH_MulU16x16", in)
		got := uint32(out.DE())<<16 | uint32(out.HL())
		if want := uint32(a) * uint32(b); got != want {
			t.Fatalf("MulU16x16(%d,%d) = %d, quer %d", a, b, got, want)
		}
		keep(t, "MulU16x16", in, out, "BC")

		out = r.Call("MATH_MulS16x16", in)
		got = uint32(out.DE())<<16 | uint32(out.HL())
		if want := uint32(int32(int16(a)) * int32(int16(b))); got != want {
			t.Fatalf("MulS16x16(%d,%d) = %d, quer %d", int16(a), int16(b), int32(got), int32(want))
		}
		keep(t, "MulS16x16", in, out, "BC")
	}

	rng := rand.New(rand.NewSource(3))
	vals := []uint32{0, 1, 0xFFFF, 0x10000, 0x7FFFFFFF, 0x80000000, 0xFFFFFFFF}
	for i := 0; i < 500; i++ {
		vals = append(vals, rng.Uint32())
	}
	for _, v := range vals {
		in := canary().WithDE(uint16(v >> 16)).WithHL(uint16(v))
		out := r.Call("MATH_Neg32", in)
		got := uint32(out.DE())<<16 | uint32(out.HL())
		if got != -v {
			t.Fatalf("Neg32(%#x) = %#x, quer %#x", v, got, -v)
		}
		keep(t, "Neg32", in, out, "BC")
	}
}

// --- divisão e resto -----------------------------------------------------------------

func TestMathDivMod(t *testing.T) {
	r := libtest.NewRunner(t, "MATH_Mod16", "MATH_DivS16", "MATH_ModS16")
	for _, p := range pairs16(4000) {
		a, b := p[0], p[1]
		in := canary().WithHL(a).WithDE(b)

		out := r.Call("MATH_Mod16", in)
		wantMod := uint16(0)
		if b != 0 {
			wantMod = a % b
		}
		if out.HL() != wantMod {
			t.Fatalf("Mod16(%d,%d) = %d, quer %d", a, b, out.HL(), wantMod)
		}
		keep(t, "Mod16", in, out, "BC")

		sa, sb := int16(a), int16(b)
		out = r.Call("MATH_DivS16", in)
		var wantQ, wantR int16
		switch {
		case sb == 0 && sa >= 0:
			wantQ, wantR = 0x7FFF, 0
		case sb == 0:
			wantQ, wantR = -0x8000, 0
		case sa == math.MinInt16 && sb == -1:
			wantQ, wantR = math.MinInt16, 0 // estouro documentado
		default:
			wantQ, wantR = sa/sb, sa%sb
		}
		if int16(out.HL()) != wantQ || int16(out.DE()) != wantR {
			t.Fatalf("DivS16(%d,%d) = q %d r %d, quer q %d r %d", sa, sb, int16(out.HL()), int16(out.DE()), wantQ, wantR)
		}
		keep(t, "DivS16", in, out, "BC")

		out = r.Call("MATH_ModS16", in)
		if int16(out.HL()) != wantR {
			t.Fatalf("ModS16(%d,%d) = %d, quer %d", sa, sb, int16(out.HL()), wantR)
		}
		keep(t, "ModS16", in, out, "BC")
	}
}

func TestMathDivMod10(t *testing.T) {
	r := libtest.NewRunner(t, "MATH_DivMod10")
	for v := 0; v < 65536; v++ {
		in := canary().WithHL(uint16(v))
		out := r.Call("MATH_DivMod10", in)
		if out.HL() != uint16(v/10) || out.A != byte(v%10) {
			t.Fatalf("DivMod10(%d) = %d r %d", v, out.HL(), out.A)
		}
		keep(t, "DivMod10", in, out, "BCDE")
	}
}

func TestMathDivU32By16(t *testing.T) {
	r := libtest.NewRunner(t, "MATH_DivU32By16")
	rng := rand.New(rand.NewSource(4))
	type tc struct {
		n uint32
		d uint16
	}
	cases := []tc{
		{0, 1}, {1, 1}, {0xFFFFFFFF, 1}, {0xFFFFFFFF, 0xFFFF}, {0xFFFF0000, 0xFFFF}, {0x80000000, 3},
		{0xFFFFFFFF, 0x8000}, {0x12345678, 0x0100}, {100, 7}, {65536, 65535}, {0xFFFE0001, 0xFFFF},
	}
	for i := 0; i < 3000; i++ {
		d := uint16(rng.Intn(65535) + 1)
		if i%3 == 0 {
			d = uint16(rng.Intn(300) + 1)
		}
		cases = append(cases, tc{rng.Uint32(), d})
	}
	for _, c := range cases {
		in := canary().WithDE(uint16(c.n >> 16)).WithHL(uint16(c.n)).WithBC(c.d)
		out := r.Call("MATH_DivU32By16", in)
		q := uint32(out.DE())<<16 | uint32(out.HL())
		if wq, wr := c.n/uint32(c.d), uint16(c.n%uint32(c.d)); q != wq || out.BC() != wr {
			t.Fatalf("DivU32By16(%d / %d) = q %d r %d, quer q %d r %d", c.n, c.d, q, out.BC(), wq, wr)
		}
	}
	// divisor zero
	out := r.Call("MATH_DivU32By16", canary().WithDE(1).WithHL(2).WithBC(0))
	if out.DE() != 0xFFFF || out.HL() != 0xFFFF || out.BC() != 0 {
		t.Errorf("divisor zero: DE:HL=%04X:%04X BC=%04X", out.DE(), out.HL(), out.BC())
	}
}

// --- deslocamentos e raiz ---------------------------------------------------------------

func TestMathShifts(t *testing.T) {
	r := libtest.NewRunner(t, "MATH_Shl16", "MATH_Shr16", "MATH_Sar16")
	for _, v := range edge16 {
		for _, b := range []byte{0, 1, 2, 7, 8, 9, 15, 16, 17, 100, 255} {
			in := canary().WithHL(v)
			in.B = b
			wantShl, wantShr, wantSar := uint16(0), uint16(0), uint16(0)
			if b < 16 {
				wantShl = v << b
				wantShr = v >> b
				wantSar = uint16(int16(v) >> b)
			} else if int16(v) < 0 {
				wantSar = 0xFFFF
			}
			for _, x := range []struct {
				name string
				want uint16
			}{{"MATH_Shl16", wantShl}, {"MATH_Shr16", wantShr}, {"MATH_Sar16", wantSar}} {
				out := r.Call(x.name, in)
				if out.HL() != x.want {
					t.Fatalf("%s(%#x, %d) = %#x, quer %#x", x.name, v, b, out.HL(), x.want)
				}
				keep(t, x.name, in, out, "ABCDE")
			}
		}
	}
}

func TestMathSqrt16(t *testing.T) {
	r := libtest.NewRunner(t, "MATH_Sqrt16")
	check := func(v int) {
		t.Helper()
		in := canary().WithHL(uint16(v))
		out := r.Call("MATH_Sqrt16", in)
		want := byte(math.Floor(math.Sqrt(float64(v))))
		if out.A != want {
			t.Fatalf("Sqrt16(%d) = %d, quer %d", v, out.A, want)
		}
		keep(t, "Sqrt16", in, out, "BCDEHL")
	}
	for k := 0; k <= 255; k++ { // quadrados perfeitos e vizinhos
		for _, d := range []int{-1, 0, 1} {
			if v := k*k + d; v >= 0 && v <= 65535 {
				check(v)
			}
		}
	}
	rng := rand.New(rand.NewSource(5))
	for i := 0; i < 300; i++ {
		check(rng.Intn(65536))
	}
	check(65535)
}

// --- aleatório ----------------------------------------------------------------------------

// xorshift16 é a referência do gerador: x ^= x<<7; x ^= x>>9; x ^= x<<8.
func xorshift16(x uint16) uint16 {
	x ^= x << 7
	x ^= x >> 9
	x ^= x << 8
	return x
}

func TestMathRandom(t *testing.T) {
	r := libtest.NewRunner(t, "MATH_RandSeed", "MATH_Rand16", "MATH_Rand8", "MATH_RandRange16", "MATH_RandRange8")

	// semente conhecida -> sequência idêntica à referência
	in := canary().WithHL(1234)
	out := r.Call("MATH_RandSeed", in)
	keep(t, "RandSeed", in, out, "ABCDEHL")
	x := uint16(1234)
	for i := 0; i < 2000; i++ {
		x = xorshift16(x)
		in := canary()
		out := r.Call("MATH_Rand16", in)
		if out.HL() != x {
			t.Fatalf("Rand16 #%d = %d, quer %d", i, out.HL(), x)
		}
		keep(t, "Rand16", in, out, "ABCDE")
	}

	// semente 0 é trocada por ACE1h
	r.Call("MATH_RandSeed", canary().WithHL(0))
	if got, want := r.Call("MATH_Rand16", canary()).HL(), xorshift16(0xACE1); got != want {
		t.Errorf("semente 0: primeiro valor %#x, quer %#x", got, want)
	}

	// período completo: 65535 valores distintos, nunca 0, e volta ao começo
	r.Call("MATH_RandSeed", canary().WithHL(1))
	seen := make([]bool, 65536)
	for i := 0; i < 65535; i++ {
		v := r.Call("MATH_Rand16", canary()).HL()
		if v == 0 || seen[v] {
			t.Fatalf("valor %d repetido/zero na posição %d", v, i)
		}
		seen[v] = true
	}
	if v := r.Call("MATH_Rand16", canary()).HL(); v != xorshift16(1) {
		t.Errorf("depois do período completo o gerador devia recomeçar: %d", v)
	}

	// Rand8 / RandRange16 / RandRange8 derivam do Rand16
	for _, max := range []uint16{0, 1, 2, 10, 100, 255, 256, 1000, 40000, 65535} {
		r.Call("MATH_RandSeed", canary().WithHL(777))
		x := uint16(777)
		for i := 0; i < 200; i++ {
			x = xorshift16(x)
			in := canary().WithHL(max)
			out := r.Call("MATH_RandRange16", in)
			want := uint16(0)
			if max != 0 {
				want = x % max
			}
			if out.HL() != want {
				t.Fatalf("RandRange16(max=%d) #%d = %d, quer %d", max, i, out.HL(), want)
			}
			keep(t, "RandRange16", in, out, "BCDE")
		}
	}
	r.Call("MATH_RandSeed", canary().WithHL(31))
	x = 31
	for i := 0; i < 300; i++ {
		x = xorshift16(x)
		out := r.Call("MATH_Rand8", canary())
		if want := byte(x>>8) ^ byte(x); out.A != want {
			t.Fatalf("Rand8 #%d = %d, quer %d", i, out.A, want)
		}
		keep(t, "Rand8", canary(), out, "BCDEHL")
	}
	r.Call("MATH_RandSeed", canary().WithHL(31))
	x = 31
	for _, max := range []byte{0, 1, 6, 100, 255} {
		if max != 0 { // max = 0 devolve 0 sem consumir um valor do gerador
			x = xorshift16(x)
		}
		in := canary()
		in.A = max
		out := r.Call("MATH_RandRange8", in)
		want := byte(0)
		if max != 0 {
			want = byte(x % uint16(max))
		}
		if out.A != want {
			t.Fatalf("RandRange8(max=%d) = %d, quer %d", max, out.A, want)
		}
		keep(t, "RandRange8", in, out, "BCDEHL")
	}
}

// Mul16 e Div16 (as rotinas originais da MSXLIB) ganham aqui a mesma checagem.
func TestMathMul16Div16(t *testing.T) {
	r := libtest.NewRunner(t, "Mul16", "Div16")
	for _, p := range pairs16(3000) {
		a, b := p[0], p[1]
		in := canary().WithHL(a).WithDE(b)
		out := r.Call("Mul16", in)
		if out.HL() != a*b {
			t.Fatalf("Mul16(%d,%d) = %d, quer %d", a, b, out.HL(), a*b)
		}
		keep(t, "Mul16", in, out, "BCDE")

		out = r.Call("Div16", in)
		wantQ, wantR := uint16(0xFFFF), uint16(0)
		if b != 0 {
			wantQ, wantR = a/b, a%b
		}
		if out.HL() != wantQ || out.DE() != wantR {
			t.Fatalf("Div16(%d,%d) = q %d r %d, quer q %d r %d", a, b, out.HL(), out.DE(), wantQ, wantR)
		}
		keep(t, "Div16", in, out, "BC")
	}
}

// --- bits, 8 bits, mod 10, faixa, ponto fixo ---------------------------------------------

func TestMathBits(t *testing.T) {
	r := libtest.NewRunner(t, "MATH_Flip8", "MATH_Flip16", "MATH_Swap16")
	rev8 := func(v byte) byte {
		var o byte
		for i := 0; i < 8; i++ {
			o = o<<1 | v>>i&1
		}
		return o
	}
	for v := 0; v < 256; v++ {
		in := canary()
		in.A = byte(v)
		out := r.Call("MATH_Flip8", in)
		if out.A != rev8(byte(v)) {
			t.Fatalf("Flip8(%08b) = %08b", v, out.A)
		}
		keep(t, "Flip8", in, out, "BCDEHL")
	}
	for v := 0; v < 65536; v++ {
		in := canary().WithHL(uint16(v))
		out := r.Call("MATH_Flip16", in)
		want := uint16(rev8(byte(v)))<<8 | uint16(rev8(byte(v>>8)))
		if out.HL() != want {
			t.Fatalf("Flip16(%016b) = %016b, quer %016b", v, out.HL(), want)
		}
		keep(t, "Flip16", in, out, "ABCDE")

		out = r.Call("MATH_Swap16", in)
		if out.HL() != uint16(v)<<8|uint16(v)>>8 {
			t.Fatalf("Swap16(%04X) = %04X", v, out.HL())
		}
		keep(t, "Swap16", in, out, "ABCDE")
	}
}

func TestMath8BitSignAndShifts(t *testing.T) {
	r := libtest.NewRunner(t, "MATH_Neg8", "MATH_Abs8", "MATH_Abs32", "MATH_Shl8", "MATH_Shr8", "MATH_Sar8")
	for v := 0; v < 256; v++ {
		in := canary()
		in.A = byte(v)
		s := int8(v)
		out := r.Call("MATH_Neg8", in)
		if out.A != byte(-s) {
			t.Fatalf("Neg8(%d) = %d", s, int8(out.A))
		}
		keep(t, "Neg8", in, out, "BCDEHL")
		out = r.Call("MATH_Abs8", in)
		wantAbs := s
		if s < 0 {
			wantAbs = -s
		}
		if out.A != byte(wantAbs) {
			t.Fatalf("Abs8(%d) = %d", s, int8(out.A))
		}
		keep(t, "Abs8", in, out, "BCDEHL")

		for _, b := range []byte{0, 1, 2, 3, 5, 7, 8, 9, 200, 255} {
			in.B = b
			wantShl, wantShr, wantSar := byte(0), byte(0), byte(0)
			if b < 8 {
				wantShl, wantShr, wantSar = byte(v)<<b, byte(v)>>b, byte(int8(v)>>b)
			} else if s < 0 {
				wantSar = 0xFF
			}
			for _, x := range []struct {
				n string
				w byte
			}{{"MATH_Shl8", wantShl}, {"MATH_Shr8", wantShr}, {"MATH_Sar8", wantSar}} {
				out := r.Call(x.n, in)
				if out.A != x.w {
					t.Fatalf("%s(%#02x, %d) = %#02x, quer %#02x", x.n, v, b, out.A, x.w)
				}
				keep(t, x.n, in, out, "BCDEHL")
			}
		}
	}
	rng := rand.New(rand.NewSource(6))
	vals := []uint32{0, 1, 0x7FFFFFFF, 0x80000000, 0xFFFFFFFF, 0xFFFF0000}
	for i := 0; i < 300; i++ {
		vals = append(vals, rng.Uint32())
	}
	for _, v := range vals {
		in := canary().WithDE(uint16(v >> 16)).WithHL(uint16(v))
		out := r.Call("MATH_Abs32", in)
		want := v
		if int32(v) < 0 {
			want = -v
		}
		if got := uint32(out.DE())<<16 | uint32(out.HL()); got != want {
			t.Fatalf("Abs32(%#x) = %#x, quer %#x", v, got, want)
		}
		keep(t, "Abs32", in, out, "ABC")
	}
}

func TestMathMod10AndDivS10(t *testing.T) {
	r := libtest.NewRunner(t, "MATH_Mod10", "MATH_DivS10")
	for v := 0; v < 65536; v++ {
		in := canary().WithHL(uint16(v))
		out := r.Call("MATH_Mod10", in)
		if out.A != byte(v%10) {
			t.Fatalf("Mod10(%d) = %d", v, out.A)
		}
		keep(t, "Mod10", in, out, "BCDEHL")

		s := int16(v)
		out = r.Call("MATH_DivS10", in)
		if int16(out.HL()) != s/10 || int8(out.A) != int8(s%10) {
			t.Fatalf("DivS10(%d) = q %d r %d, quer q %d r %d", s, int16(out.HL()), int8(out.A), s/10, s%10)
		}
		keep(t, "DivS10", in, out, "BCDE")
	}
}

func TestMathRandBetween(t *testing.T) {
	r := libtest.NewRunner(t, "MATH_RandSeed", "MATH_Rand16", "MATH_RandBetween16", "MATH_RandBetween8")
	for _, lo := range []uint16{0, 1, 5, 100, 30000, 65000} {
		for _, hi := range []uint16{0, 1, 5, 6, 101, 30001, 65535} {
			r.Call("MATH_RandSeed", canary().WithHL(4321))
			x := uint16(4321)
			for i := 0; i < 50; i++ {
				in := canary().WithHL(lo).WithDE(hi)
				out := r.Call("MATH_RandBetween16", in)
				var want uint16
				if hi > lo {
					x = xorshift16(x)
					want = lo + x%(hi-lo)
				} else {
					want = lo
					if hi == lo {
						want = lo // hi = lo: RandRange16(0) = 0 sem consumir o gerador
					}
				}
				if out.HL() != want {
					t.Fatalf("RandBetween16(%d,%d) #%d = %d, quer %d", lo, hi, i, out.HL(), want)
				}
				keep(t, "RandBetween16", in, out, "BCDE")
			}
		}
	}
	for _, c := range [][2]byte{{0, 6}, {10, 20}, {250, 255}, {5, 5}, {9, 3}} {
		for i := 0; i < 40; i++ {
			in := canary()
			in.A, in.E = c[0], c[1]
			out := r.Call("MATH_RandBetween8", in)
			if c[1] > c[0] {
				if out.A < c[0] || out.A >= c[1] {
					t.Fatalf("RandBetween8(%d,%d) = %d fora da faixa", c[0], c[1], out.A)
				}
			} else if out.A != c[0] {
				t.Fatalf("RandBetween8(%d,%d) = %d, quer %d", c[0], c[1], out.A, c[0])
			}
			keep(t, "RandBetween8", in, out, "BCDEHL")
		}
	}
}

func TestMathFixed88(t *testing.T) {
	r := libtest.NewRunner(t, "MATH_FixMul88", "MATH_FixDiv88")
	rng := rand.New(rand.NewSource(7))
	vals := []uint16{0, 1, 0x0080, 0x0100, 0x0180, 0x0200, 0x7FFF, 0x8000, 0x8001, 0xFF00, 0xFE80, 0xFFFF, 0x0A00, 0xF600}
	for i := 0; i < 60; i++ {
		vals = append(vals, uint16(rng.Intn(65536)))
	}
	for _, a := range vals {
		for _, b := range vals {
			in := canary().WithHL(a).WithDE(b)

			out := r.Call("MATH_FixMul88", in)
			wantMul := uint16((int32(int16(a)) * int32(int16(b))) >> 8)
			if out.HL() != wantMul {
				t.Fatalf("FixMul88(%04X,%04X) = %04X, quer %04X", a, b, out.HL(), wantMul)
			}
			keep(t, "FixMul88", in, out, "BCDE")

			out = r.Call("MATH_FixDiv88", in)
			sa, sb := int32(int16(a)), int32(int16(b))
			var wantDiv uint16
			switch {
			case sb == 0 && sa >= 0:
				wantDiv = 0x7FFF
			case sb == 0:
				wantDiv = 0x8000
			default: // (a << 8) / b, truncado para zero, so os 16 bits baixos
				wantDiv = uint16((sa << 8) / sb)
			}
			if out.HL() != wantDiv {
				t.Fatalf("FixDiv88(%04X,%04X) = %04X, quer %04X", a, b, out.HL(), wantDiv)
			}
			keep(t, "FixDiv88", in, out, "BCDE")
		}
	}
}
