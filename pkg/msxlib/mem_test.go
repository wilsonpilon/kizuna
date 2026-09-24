package msxlib_test

import (
	"bytes"
	"fmt"
	"math/rand"
	"testing"

	"github.com/wilsonpilon/kizuna/pkg/libtest"
)

const memBase = 0x8000 // área de trabalho dos testes de memória

func fillPattern(mem []byte, seed int64) {
	r := rand.New(rand.NewSource(seed))
	for i := range mem {
		mem[i] = byte(r.Intn(256))
	}
}

// --- cópia -------------------------------------------------------------------

func TestMemCopyOverlapAndDirection(t *testing.T) {
	r := libtest.NewRunner(t, "MEM_Copy", "MEM_CopyFast", "MEM_CopyRev")
	rng := rand.New(rand.NewSource(10))
	area := r.M.Mem[memBase : memBase+2048]

	type call struct{ src, dst, n int }
	var cases []call
	// sobreposições e vizinhanças, em torno de tamanhos interessantes
	for _, n := range []int{0, 1, 2, 3, 7, 16, 100, 255, 256, 257} {
		for _, d := range []int{-300, -256, -100, -10, -2, -1, 0, 1, 2, 10, 100, 256, 300} {
			cases = append(cases, call{src: 600, dst: 600 + d, n: n})
		}
	}
	for i := 0; i < 400; i++ {
		cases = append(cases, call{src: rng.Intn(1500), dst: rng.Intn(1500), n: rng.Intn(500)})
	}
	for ci, c := range cases {
		if c.src+c.n > len(area) || c.dst+c.n > len(area) || c.src < 0 || c.dst < 0 {
			continue
		}
		for _, fn := range []string{"MEM_Copy", "MEM_CopyFast", "MEM_CopyRev"} {
			// só chama as variantes rápidas nos casos em que são seguras
			switch fn {
			case "MEM_CopyFast":
				if c.dst > c.src && c.dst < c.src+c.n { // destino dentro da origem: LDIR corromperia
					continue
				}
			case "MEM_CopyRev":
				if c.dst < c.src && c.dst+c.n > c.src { // destino antes e invadindo: LDDR corromperia
					continue
				}
			}
			fillPattern(area, int64(ci))
			want := append([]byte(nil), area...)
			copy(want[c.dst:c.dst+c.n], want[c.src:c.src+c.n]) // copy() do Go = memmove
			in := canary().WithHL(uint16(memBase + c.src)).WithDE(uint16(memBase + c.dst)).WithBC(uint16(c.n))
			r.Call(fn, in)
			if !bytes.Equal(area, want) {
				t.Fatalf("%s(src=%d dst=%d n=%d): memória diferente da esperada", fn, c.src, c.dst, c.n)
			}
		}
	}
}

// --- preenchimento, comparação, busca, troca --------------------------------------

func TestMemFillZeroFill16(t *testing.T) {
	r := libtest.NewRunner(t, "MEM_Fill", "MEM_Zero", "MEM_Fill16")
	area := r.M.Mem[memBase : memBase+600]
	for _, n := range []int{0, 1, 2, 3, 100, 255, 256, 257, 500} {
		for _, val := range []byte{0x00, 0x5A, 0xFF} {
			fillPattern(area, int64(n))
			want := append([]byte(nil), area...)
			for i := 0; i < n; i++ {
				want[10+i] = val
			}
			in := canary().WithHL(memBase + 10).WithBC(uint16(n))
			in.A = val
			out := r.Call("MEM_Fill", in)
			if !bytes.Equal(area, want) {
				t.Fatalf("Fill(n=%d, val=%02X): memória errada", n, val)
			}
			keep(t, "Fill", in, out, "DE")
		}
		fillPattern(area, 99)
		want := append([]byte(nil), area...)
		for i := 0; i < n; i++ {
			want[10+i] = 0
		}
		in := canary().WithHL(memBase + 10).WithBC(uint16(n))
		out := r.Call("MEM_Zero", in)
		if !bytes.Equal(area, want) {
			t.Fatalf("Zero(n=%d): memória errada", n)
		}
		keep(t, "Zero", in, out, "DE")
	}
	for _, n := range []int{0, 1, 2, 5, 100, 250} { // n PALAVRAS
		fillPattern(area, int64(n))
		want := append([]byte(nil), area...)
		for i := 0; i < n; i++ {
			want[20+2*i], want[21+2*i] = 0x34, 0x12
		}
		in := canary().WithHL(memBase + 20).WithBC(uint16(n)).WithDE(0x1234)
		r.Call("MEM_Fill16", in)
		if !bytes.Equal(area, want) {
			t.Fatalf("Fill16(n=%d): memória errada", n)
		}
	}
}

func TestMemCompareFindSwap(t *testing.T) {
	r := libtest.NewRunner(t, "MEM_Compare", "MEM_Find", "MEM_Swap")
	a := r.M.Mem[memBase : memBase+300]
	b := r.M.Mem[memBase+300 : memBase+600]

	for _, n := range []int{0, 1, 2, 50, 256, 300} {
		fillPattern(a, 1)
		copy(b, a)
		in := canary().WithHL(memBase).WithDE(memBase + 300).WithBC(uint16(n))
		out := r.Call("MEM_Compare", in)
		if out.A != 0 || !out.Zero() {
			t.Fatalf("Compare(iguais, n=%d): A=%d Z=%v", n, out.A, out.Zero())
		}
		for pos := 0; pos < n; pos += 1 + n/7 {
			for _, delta := range []int{+1, -1, +100, -100} {
				copy(b, a)
				nv := int(a[pos]) + delta
				if nv < 0 || nv > 255 {
					continue
				}
				b[pos] = byte(nv)
				b[n-1] ^= 0 // (só para deixar claro: o resto continua igual)
				out := r.Call("MEM_Compare", in)
				want := byte(1) // a > b
				if nv > int(a[pos]) {
					want = 0xFF // a < b
				}
				if out.A != want || out.Zero() {
					t.Fatalf("Compare(n=%d, diferença em %d: a=%d b=%d): A=%d Z=%v, quer A=%d", n, pos, a[pos], nv, out.A, out.Zero(), want)
				}
			}
		}
	}

	// Find
	fillPattern(a, 2)
	for i := range a {
		if a[i] == 0xC3 || a[i] == 0xC5 {
			a[i] = 0xC4
		}
	}
	a[57], a[120], a[299] = 0xC3, 0xC3, 0xC3
	find := func(n int, val byte) (uint16, bool) {
		in := canary().WithHL(memBase).WithBC(uint16(n))
		in.A = val
		out := r.Call("MEM_Find", in)
		return out.HL(), out.Zero()
	}
	if hl, z := find(300, 0xC3); hl != memBase+57 || !z {
		t.Errorf("Find primeiro: HL=%04X Z=%v", hl, z)
	}
	if hl, z := find(57, 0xC3); hl != 0 || z {
		t.Errorf("Find fora da faixa (n=57): HL=%04X Z=%v", hl, z)
	}
	if hl, z := find(58, 0xC3); hl != memBase+57 || !z {
		t.Errorf("Find no ultimo byte da faixa: HL=%04X Z=%v", hl, z)
	}
	if hl, z := find(300, 0xC5); hl != 0 || z {
		t.Errorf("Find ausente: HL=%04X Z=%v", hl, z)
	}
	if hl, z := find(0, 0xC3); hl != 0 || z {
		t.Errorf("Find n=0: HL=%04X Z=%v", hl, z)
	}

	// Swap
	for _, n := range []int{0, 1, 2, 77, 300} {
		fillPattern(a, 3)
		fillPattern(b, 4)
		wa, wb := append([]byte(nil), a...), append([]byte(nil), b...)
		for i := 0; i < n; i++ {
			wa[i], wb[i] = wb[i], wa[i]
		}
		r.Call("MEM_Swap", canary().WithHL(memBase).WithDE(memBase+300).WithBC(uint16(n)))
		if !bytes.Equal(a, wa) || !bytes.Equal(b, wb) {
			t.Fatalf("Swap(n=%d): blocos errados", n)
		}
	}
}

func TestMemSysInfo(t *testing.T) {
	r := libtest.NewRunner(t, "MEM_GetSP", "MEM_TPATop")
	r.M.Mem[6], r.M.Mem[7] = 0x00, 0xC4
	in := canary()
	out := r.Call("MEM_TPATop", in)
	if out.HL() != 0xC400 {
		t.Errorf("TPATop = %04X, quer C400", out.HL())
	}
	keep(t, "TPATop", in, out, "ABCDE")
	// Machine.Call arma SP = FF00-2 com o retorno; depois do RET o SP do chamador é FF00
	out = r.Call("MEM_GetSP", in)
	if out.HL() != 0xFF00 {
		t.Errorf("GetSP = %04X, quer FF00", out.HL())
	}
	keep(t, "GetSP", in, out, "ABCDE")
}

// --- heap ------------------------------------------------------------------------------

// heapModel é a mesma política do heap em Go (primeiro que couber, fusão
// preguiçosa de livres vizinhos, sobra vira bloco livre a partir de 4 bytes),
// escrita sobre uma cópia da memória para poder ser comparada byte a byte com a
// memória do simulador depois de cada operação.
type heapModel struct {
	mem        [65536]byte
	start, end uint16
}

func (h *heapModel) rd(a uint16) uint16 { return uint16(h.mem[a]) | uint16(h.mem[a+1])<<8 }
func (h *heapModel) wr(a, v uint16)     { h.mem[a], h.mem[a+1] = byte(v), byte(v>>8) }

func (h *heapModel) init(start, size uint16) byte {
	if size == 0 {
		h.start, h.end = 0, 0
		return 1
	}
	if start&1 != 0 {
		start++
		size--
	}
	size &^= 1
	if size < 6 {
		h.start, h.end = 0, 0
		return 1
	}
	if uint32(start)+uint32(size) > 0xFFFF {
		h.start, h.end = 0, 0
		return 1
	}
	h.end = start + size
	h.wr(h.end-2, 1)
	h.start = start
	h.wr(start, size-4)
	return 0
}

func (h *heapModel) alloc(size uint16) uint16 {
	if size == 0 || size == 0xFFFF {
		return 0
	}
	need := (size + 1) &^ 1
	p := h.start
	if p == 0 {
		return 0
	}
	for {
		hdr := h.rd(p)
		if hdr&1 == 0 {
			for {
				nh := h.rd(p + 2 + hdr)
				if nh&1 != 0 {
					break
				}
				hdr = hdr + 2 + nh
				h.wr(p, hdr)
			}
			if hdr >= need {
				rest := hdr - need
				if rest >= 4 {
					h.wr(p+2+need, rest-2)
					h.wr(p, need|1)
				} else {
					h.wr(p, hdr|1)
				}
				return p + 2
			}
			p += 2 + hdr
			continue
		}
		sz := hdr &^ 1
		if sz == 0 {
			return 0
		}
		p += 2 + sz
	}
}

func (h *heapModel) free(ptr uint16) byte {
	if ptr == 0 {
		return 0
	}
	if ptr&1 != 0 || ptr < h.start+2 || ptr >= h.end {
		return 1
	}
	if h.rd(ptr-2)&1 == 0 {
		return 1
	}
	h.wr(ptr-2, h.rd(ptr-2)&^1)
	return 0
}

func (h *heapModel) compact() {
	if h.start == 0 {
		return
	}
	p := h.start
	for {
		hdr := h.rd(p)
		if hdr&1 == 0 {
			for {
				nh := h.rd(p + 2 + hdr)
				if nh&1 != 0 {
					break
				}
				hdr = hdr + 2 + nh
				h.wr(p, hdr)
			}
		} else if hdr&^1 == 0 {
			return
		}
		p += 2 + hdr&^1
	}
}

func (h *heapModel) freeTotal() (total, largest uint16) {
	h.compact()
	if h.start == 0 {
		return 0, 0
	}
	for p := h.start; ; {
		hdr := h.rd(p)
		if hdr&1 == 0 {
			total += hdr
			if hdr > largest {
				largest = hdr
			}
		} else if hdr&^1 == 0 {
			return
		}
		p += 2 + hdr&^1
	}
}

// checkHeapLayout confere, direto na memória do simulador e sem usar o modelo,
// que o heap está bem formado: os blocos ladrilham a região exatamente, todos os
// tamanhos são pares e a sentinela está no fim.
func checkHeapLayout(t *testing.T, mem []byte, start, end uint16) {
	t.Helper()
	p := int(start)
	for {
		if p+2 > int(end) {
			t.Fatalf("lista de blocos passou do fim do heap (p=%04X, fim=%04X)", p, end)
		}
		hdr := int(mem[p]) | int(mem[p+1])<<8
		sz := hdr &^ 1
		if hdr&1 == 1 && sz == 0 {
			if p != int(end)-2 {
				t.Fatalf("sentinela fora do lugar: %04X, quer %04X", p, end-2)
			}
			return
		}
		if sz%2 != 0 || sz < 0 {
			t.Fatalf("bloco em %04X com tamanho ímpar %d", p, sz)
		}
		p += 2 + sz
	}
}

func TestMemHeapInit(t *testing.T) {
	r := libtest.NewRunner(t, "MEM_HeapInit", "MEM_HeapInitToStack", "MEM_HeapSize", "MEM_Alloc")
	cases := []struct{ start, size uint16 }{
		{0xA000, 4096}, {0xA001, 4096}, {0xA000, 4095}, {0xA001, 4097}, {0xA000, 6}, {0xA000, 7}, {0xA000, 5},
		{0xA000, 0}, {0xA001, 1}, {0xA001, 6}, {0xA000, 2}, {0xFFF0, 100}, {0xFFF0, 16}, {0x9000, 0x6000},
	}
	for _, c := range cases {
		var h heapModel
		clear(r.M.Mem[0x8000:0xFE00]) // sem lixo de casos anteriores
		wantA := h.init(c.start, c.size)
		in := canary().WithHL(c.start).WithBC(c.size)
		out := r.Call("MEM_HeapInit", in)
		if out.A != wantA {
			t.Fatalf("HeapInit(%04X, %d): A=%d, quer %d", c.start, c.size, out.A, wantA)
		}
		hs := r.Call("MEM_HeapSize", canary()).HL()
		if hs != h.end-h.start {
			t.Fatalf("HeapSize depois de Init(%04X,%d) = %d, quer %d", c.start, c.size, hs, h.end-h.start)
		}
		if wantA == 0 {
			checkHeapLayout(t, r.M.Mem[:], h.start, h.end)
			if !bytes.Equal(r.M.Mem[h.start:h.start+2], h.mem[h.start:h.start+2]) ||
				!bytes.Equal(r.M.Mem[h.end-2:h.end], h.mem[h.end-2:h.end]) {
				t.Fatalf("cabecalho/sentinela depois de Init(%04X,%d) diferem do modelo", c.start, c.size)
			}
		} else if out := r.Call("MEM_Alloc", canary().WithHL(2)); out.HL() != 0 {
			t.Fatalf("Alloc num heap nao iniciado devolveu %04X", out.HL())
		}
	}

	// InitToStack: o SP do chamador em Runner.Call é FF00
	for _, c := range []struct {
		start, margin uint16
		ok            bool
	}{{0xC000, 0x100, true}, {0xC001, 0x0200, true}, {0xFE00, 0x100, false}, {0xFF00, 0, false}, {0xF000, 0x1000, false}, {0xFE00, 0xF8, true}} {
		out := r.Call("MEM_HeapInitToStack", canary().WithHL(c.start).WithBC(c.margin))
		if (out.A == 0) != c.ok {
			t.Fatalf("InitToStack(%04X, margem %04X): A=%d, ok esperado=%v", c.start, c.margin, out.A, c.ok)
		}
		if c.ok {
			var h heapModel
			h.init(c.start, 0xFF00-c.margin-c.start)
			if got := r.Call("MEM_HeapSize", canary()).HL(); got != h.end-h.start {
				t.Errorf("InitToStack(%04X, %04X): tamanho %d, quer %d", c.start, c.margin, got, h.end-h.start)
			}
		}
	}
}

func TestMemHeapAgainstModel(t *testing.T) {
	r := libtest.NewRunner(t, "MEM_HeapInit", "MEM_Alloc", "MEM_Free", "MEM_HeapCompact", "MEM_HeapSize", "MEM_HeapFree", "MEM_HeapLargest")
	const start, size = 0xA000, 3000
	for seed := int64(1); seed <= 12; seed++ {
		rng := rand.New(rand.NewSource(seed))
		var h heapModel
		clear(r.M.Mem[start : start+size+16])
		h.init(start, size)
		if out := r.Call("MEM_HeapInit", canary().WithHL(start).WithBC(size)); out.A != 0 {
			t.Fatal("HeapInit falhou")
		}
		type blk struct {
			ptr  uint16
			size int
			id   byte
		}
		var live []blk
		nextID := byte(1)

		sameImage := func(what string) {
			t.Helper()
			if !bytes.Equal(r.M.Mem[start:start+size], h.mem[start:start+size]) {
				for i := start; i < start+size; i++ {
					if r.M.Mem[i] != h.mem[i] {
						t.Fatalf("seed %d, %s: heap difere do modelo no offset %d (%02X x %02X)", seed, what, i-start, r.M.Mem[i], h.mem[i])
					}
				}
			}
			checkHeapLayout(t, r.M.Mem[:], h.start, h.end)
		}

		for step := 0; step < 600; step++ {
			switch op := rng.Intn(10); {
			case op < 5 || len(live) == 0: // alloc
				n := 1 + rng.Intn(40)
				switch rng.Intn(8) {
				case 0:
					n = 100 + rng.Intn(600)
				case 1:
					n = 1 + rng.Intn(3)
				}
				want := h.alloc(uint16(n))
				in := canary().WithHL(uint16(n))
				out := r.Call("MEM_Alloc", in)
				keep(t, "Alloc", in, out, "BCDE")
				if out.HL() != want {
					t.Fatalf("seed %d passo %d: Alloc(%d) = %04X, modelo %04X", seed, step, n, out.HL(), want)
				}
				if want != 0 {
					id := nextID
					nextID++
					if nextID == 0 {
						nextID = 1
					}
					for i := 0; i < n; i++ { // marca o conteúdo nas duas memórias
						r.M.Mem[int(want)+i] = id
						h.mem[int(want)+i] = id
					}
					live = append(live, blk{want, n, id})
				}
				sameImage(fmt.Sprintf("apos Alloc(%d)", n))
			case op < 9: // free de um bloco vivo
				i := rng.Intn(len(live))
				b := live[i]
				live = append(live[:i], live[i+1:]...)
				wantA := h.free(b.ptr)
				in := canary().WithHL(b.ptr)
				out := r.Call("MEM_Free", in)
				keep(t, "Free", in, out, "BCDEHL")
				if out.A != wantA || wantA != 0 {
					t.Fatalf("seed %d: Free(%04X) = %d, modelo %d", seed, b.ptr, out.A, wantA)
				}
				sameImage("apos Free")
			default: // ponteiros ruins e liberação dupla
				bad := []uint16{0, 1, start, start + 1, start + 3, start + size, start + size + 2, 0x100, 0xFFFF}
				p := bad[rng.Intn(len(bad))]
				wantA := h.free(p)
				out := r.Call("MEM_Free", canary().WithHL(p))
				if out.A != wantA {
					t.Fatalf("seed %d: Free ruim(%04X) = %d, modelo %d", seed, p, out.A, wantA)
				}
				sameImage("apos Free ruim")
			}
			// integridade independente do modelo: o conteúdo de todo bloco vivo continua intacto
			if step%25 == 0 {
				for _, b := range live {
					for i := 0; i < b.size; i++ {
						if r.M.Mem[int(b.ptr)+i] != b.id {
							t.Fatalf("seed %d passo %d: conteudo do bloco %04X corrompido no byte %d", seed, step, b.ptr, i)
						}
					}
				}
			}
			if step%97 == 0 { // estatísticas
				wantTotal, wantLargest := func() (uint16, uint16) {
					// o modelo funde antes de medir; copia para não alterar o estado do modelo real
					c := *h.clone()
					return c.freeTotal()
				}()
				gotFree := r.Call("MEM_HeapFree", canary()).HL()
				// HeapFree soma SEM fundir; recalcula no modelo sem fundir
				var sum uint16
				for p := h.start; ; {
					hdr := h.rd(p)
					if hdr&1 == 0 {
						sum += hdr
					} else if hdr&^1 == 0 {
						break
					}
					p += 2 + hdr&^1
				}
				if gotFree != sum {
					t.Fatalf("seed %d passo %d: HeapFree = %d, quer %d", seed, step, gotFree, sum)
				}
				gotLargest := r.Call("MEM_HeapLargest", canary()).HL()
				h.compact() // HeapLargest funde os livres vizinhos
				if gotLargest != wantLargest {
					t.Fatalf("seed %d passo %d: HeapLargest = %d, quer %d", seed, step, gotLargest, wantLargest)
				}
				_ = wantTotal
				sameImage("apos HeapLargest (com fusao)")
			}
		}
		// libera tudo, funde e confere que sobrou um bloco único do tamanho original
		for _, b := range live {
			h.free(b.ptr)
			if out := r.Call("MEM_Free", canary().WithHL(b.ptr)); out.A != 0 {
				t.Fatalf("seed %d: Free final falhou", seed)
			}
		}
		r.Call("MEM_HeapCompact", canary())
		h.compact()
		sameImage("apos liberar tudo e compactar")
		if got := r.Call("MEM_HeapLargest", canary()).HL(); got != size-4 {
			t.Fatalf("seed %d: depois de liberar tudo, maior bloco = %d, quer %d", seed, got, size-4)
		}
	}
}

func (h *heapModel) clone() *heapModel {
	c := *h
	return &c
}

func TestMemCopyWordsAndBlockSize(t *testing.T) {
	r := libtest.NewRunner(t, "MEM_CopyWords", "MEM_CopyFastWords", "MEM_HeapInit", "MEM_Alloc", "MEM_BlockSize")
	area := r.M.Mem[memBase : memBase+800]
	for _, fn := range []string{"MEM_CopyWords", "MEM_CopyFastWords"} {
		for _, n := range []int{0, 1, 2, 50, 200} {
			fillPattern(area, int64(n))
			want := append([]byte(nil), area...)
			copy(want[400:400+2*n], want[0:2*n])
			r.Call(fn, canary().WithHL(memBase).WithDE(memBase+400).WithBC(uint16(n)))
			if !bytes.Equal(area, want) {
				t.Fatalf("%s(n=%d palavras): memoria errada", fn, n)
			}
		}
	}

	clear(r.M.Mem[0xA000:0xA200])
	r.Call("MEM_HeapInit", canary().WithHL(0xA000).WithBC(0x100))
	for _, n := range []uint16{1, 2, 3, 10, 33, 100} {
		p := r.Call("MEM_Alloc", canary().WithHL(n)).HL()
		if p == 0 {
			t.Fatalf("Alloc(%d) falhou", n)
		}
		in := canary().WithHL(p)
		out := r.Call("MEM_BlockSize", in)
		if want := (n + 1) &^ 1; out.HL() != want {
			t.Errorf("BlockSize(Alloc(%d)) = %d, quer %d", n, out.HL(), want)
		}
		keep(t, "BlockSize", in, out, "ABCDE")
	}
}
