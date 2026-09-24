package msxlib_test

import (
	"math/rand"
	"testing"

	"github.com/wilsonpilon/kizuna/pkg/libtest"
	"github.com/wilsonpilon/kizuna/pkg/z80sim"
)

// gfx acessa os pixels de um modo bitmap direto na VRAM do simulador -- uma
// implementação independente do motor de comandos do modelo (só usa a
// organização da VRAM descrita na folha de dados).
type gfx struct {
	v           *z80sim.VDP
	mode        int
	w, h        int
	bpp, bpl    int
	pixelsPerBy int
}

func newGfx(v *z80sim.VDP, mode int) *gfx {
	g := &gfx{v: v, mode: mode, h: 212}
	switch mode {
	case 5:
		g.w, g.bpp, g.bpl = 256, 4, 128
	case 6:
		g.w, g.bpp, g.bpl = 512, 2, 128
	case 7:
		g.w, g.bpp, g.bpl = 512, 4, 256
	case 8:
		g.w, g.bpp, g.bpl = 256, 8, 256
	}
	g.pixelsPerBy = 8 / g.bpp
	return g
}

func (g *gfx) mask() byte { return byte(1<<uint(g.bpp) - 1) }

func (g *gfx) get(x, y int) byte {
	a := y*g.bpl + x/g.pixelsPerBy
	sh := uint(8 - g.bpp - (x%g.pixelsPerBy)*g.bpp)
	return g.v.VRAM[a] >> sh & g.mask()
}

func (g *gfx) put(x, y int, c byte) {
	a := y*g.bpl + x/g.pixelsPerBy
	sh := uint(8 - g.bpp - (x%g.pixelsPerBy)*g.bpp)
	g.v.VRAM[a] = g.v.VRAM[a]&^(g.mask()<<sh) | (c&g.mask())<<sh
}

func (g *gfx) clear() {
	for i := range g.v.VRAM {
		g.v.VRAM[i] = 0
	}
}

// image é a cópia em Go da tela: as operações de referência mexem nela e depois ela é
// comparada com a VRAM do simulador.
type image struct {
	g   *gfx
	pix [][]byte
}

func (g *gfx) snapshot() *image {
	im := &image{g: g, pix: make([][]byte, g.h)}
	for y := range im.pix {
		im.pix[y] = make([]byte, g.w)
		for x := range im.pix[y] {
			im.pix[y][x] = g.get(x, y)
		}
	}
	return im
}

func refOp(dst, src, op, mask byte) byte {
	src &= mask
	if op&8 != 0 && src == 0 {
		return dst
	}
	switch op & 7 {
	case 1:
		return dst & src
	case 2:
		return dst | src
	case 3:
		return dst ^ src
	case 4:
		return ^src & mask
	}
	return src
}

func (im *image) set(x, y int, c, op byte) {
	if x < 0 || y < 0 || x >= im.g.w || y >= im.g.h {
		return
	}
	im.pix[y][x] = refOp(im.pix[y][x], c, op, im.g.mask())
}

func (im *image) compare(t *testing.T, what string) {
	t.Helper()
	for y := 0; y < im.g.h; y++ {
		for x := 0; x < im.g.w; x++ {
			if got := im.g.get(x, y); got != im.pix[y][x] {
				t.Fatalf("%s: pixel (%d,%d) = %X, quer %X", what, x, y, got, im.pix[y][x])
			}
		}
	}
}

func cmdRunner(t *testing.T, extra ...string) (*libtest.Runner, *z80sim.VDP) {
	names := append([]string{"VDP_SetMode", "VDP_CmdWait", "VDP_CmdBusy", "VDP_CmdStop", "VDP_CmdRun"}, extra...)
	r, v := vdpRunner(t, names...)
	v.VBlankEvery = 0
	return r, v
}

func put16(r *libtest.Runner, addr int, words ...int) {
	for i, w := range words {
		r.M.Mem[addr+2*i] = byte(w)
		r.M.Mem[addr+2*i+1] = byte(w >> 8)
	}
}

var bitmapModes = []int{5, 6, 7, 8}

const blk = 0x9000

func TestVDPCmdPrimitives(t *testing.T) {
	r, v := cmdRunner(t, "VDP_HwPlot", "VDP_HwPoint")
	for _, mode := range bitmapModes {
		setMode(r, mode)
		g := newGfx(v, mode)
		g.clear()
		rng := rand.New(rand.NewSource(int64(mode)))
		im := g.snapshot()
		for i := 0; i < 300; i++ {
			x, y := rng.Intn(g.w), rng.Intn(g.h)
			c := byte(rng.Intn(256))
			op := []byte{0, 1, 2, 3, 4, 8, 9, 0xB}[rng.Intn(8)]
			in := canary()
			in.B, in.C = byte(x>>8), byte(x)
			in.D, in.E = byte(y>>8), byte(y)
			in.A, in.H = c, op
			out := r.Call("VDP_HwPlot", in)
			keep(t, "HwPlot", in, out, "ABCDEHL")
			im.set(x, y, c, op)

			in = canary()
			in.B, in.C = byte(x>>8), byte(x)
			in.D, in.E = byte(y>>8), byte(y)
			out = r.Call("VDP_HwPoint", in)
			if out.A != im.pix[y][x] {
				t.Fatalf("modo %d: HwPoint(%d,%d) = %X, quer %X", mode, x, y, out.A, im.pix[y][x])
			}
			keep(t, "HwPoint", in, out, "BCDEHL")
		}
		im.compare(t, "HwPlot")
		if v.Reg[15] != 0 {
			t.Fatalf("R#15 = %d depois dos comandos", v.Reg[15])
		}
	}
	// fora de modo bitmap o comando é ignorado (o chip não desenha)
	setMode(r, 1)
	before := v.Commands
	in := canary()
	in.A, in.H = 5, 0
	r.Call("VDP_HwPlot", in)
	if v.Commands != before {
		t.Error("comando aceito em SCREEN 1")
	}
}

func TestVDPCmdWaitBusyStopRun(t *testing.T) {
	r, v := cmdRunner(t)
	setMode(r, 5)
	in := canary()
	out := r.Call("VDP_CmdBusy", in)
	if out.A != 0 {
		t.Fatalf("CmdBusy com o motor livre = %d", out.A)
	}
	keep(t, "CmdBusy", in, out, "BCDEHL")
	out = r.Call("VDP_CmdWait", in)
	keep(t, "CmdWait", in, out, "ABCDEHL")

	// bloco montado à mão: LMMV 4x3 na posição (20,30) com a cor 9
	r.M.Mem[blk-1] = 0xEE
	block := []byte{0, 0, 0, 0, 20, 0, 30, 0, 4, 0, 3, 0, 9, 0, 0x80}
	copy(r.M.Mem[blk:], block)
	in = canary()
	in.H, in.L = blk>>8, blk&0xFF
	out = r.Call("VDP_CmdRun", in)
	keep(t, "CmdRun", in, out, "ABCHL")
	g := newGfx(v, 5)
	for y := 30; y < 33; y++ {
		for x := 20; x < 24; x++ {
			if g.get(x, y) != 9 {
				t.Fatalf("CmdRun não pintou (%d,%d)", x, y)
			}
		}
	}
	for i, b := range block {
		if v.Reg[32+i] != b {
			t.Fatalf("R#%d = %02X, quer %02X", 32+i, v.Reg[32+i], b)
		}
	}
	// CmdStop escreve R#46 = 0
	r.Call("VDP_CmdStop", canary())
	if v.Reg[46] != 0 {
		t.Fatalf("CmdStop: R#46 = %02X", v.Reg[46])
	}
}

func TestVDPHwRects(t *testing.T) {
	r, v := cmdRunner(t, "VDP_HwFillRect", "VDP_HwFillRectFast", "VDP_HwBoxFill")
	rng := rand.New(rand.NewSource(7))
	for _, mode := range bitmapModes {
		setMode(r, mode)
		g := newGfx(v, mode)
		g.clear()
		// fundo aleatório
		for y := 0; y < g.h; y++ {
			for x := 0; x < g.w; x++ {
				g.put(x, y, byte(rng.Intn(256)))
			}
		}
		im := g.snapshot()
		for i := 0; i < 40; i++ {
			x, y := rng.Intn(g.w-40), rng.Intn(g.h-30)
			w, h := 1+rng.Intn(39), 1+rng.Intn(29)
			c := byte(rng.Intn(256))
			op := []byte{0, 1, 2, 3, 4, 8, 0xB}[rng.Intn(7)]
			put16(r, blk, x, y, w, h)
			in := canary()
			in.H, in.L, in.A, in.B = blk>>8, blk&0xFF, c, op
			out := r.Call("VDP_HwFillRect", in)
			keep(t, "HwFillRect", in, out, "ABCDEHL")
			for j := 0; j < h; j++ {
				for k := 0; k < w; k++ {
					im.set(x+k, y+j, c, op)
				}
			}
			im.compare(t, "HwFillRect")
		}
		// retângulo por cantos, em qualquer ordem, cheio
		for i := 0; i < 40; i++ {
			x1, y1 := 5+rng.Intn(g.w-10), 5+rng.Intn(g.h-10)
			x2, y2 := 5+rng.Intn(g.w-10), 5+rng.Intn(g.h-10)
			c := byte(rng.Intn(256))
			op := []byte{0, 2, 3}[rng.Intn(3)]
			put16(r, blk, x1, y1, x2, y2)
			in := canary()
			in.H, in.L, in.A, in.B = blk>>8, blk&0xFF, c, op
			out := r.Call("VDP_HwBoxFill", in)
			keep(t, "HwBoxFill", in, out, "ABCDEHL")
			lx, hx, ly, hy := min(x1, x2), max(x1, x2), min(y1, y2), max(y1, y2)
			for yy := ly; yy <= hy; yy++ {
				for xx := lx; xx <= hx; xx++ {
					im.set(xx, yy, c, op)
				}
			}
			im.compare(t, "HwBoxFill")
		}
		// preenchimento rápido: por bytes
		pp := g.pixelsPerBy
		for i := 0; i < 20; i++ {
			x, y := pp*rng.Intn((g.w-40)/pp), rng.Intn(g.h-30)
			w, h := pp*(1+rng.Intn(39/pp)), 1+rng.Intn(29)
			b := byte(rng.Intn(256))
			put16(r, blk, x, y, w, h)
			in := canary()
			in.H, in.L, in.A = blk>>8, blk&0xFF, b
			out := r.Call("VDP_HwFillRectFast", in)
			keep(t, "HwFillRectFast", in, out, "ABCDEHL")
			for j := 0; j < h; j++ {
				for k := 0; k < w; k++ {
					sh := uint(8 - g.bpp - ((x+k)%pp)*g.bpp)
					im.pix[y+j][x+k] = b >> sh & g.mask()
				}
			}
			im.compare(t, "HwFillRectFast")
		}
	}
}

func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}

// linePoints confere as propriedades de uma linha de Bresenham no conjunto de pontos
// desenhados: extremos incluídos, um ponto por coordenada do lado maior, nenhum a mais
// de meio ponto da reta ideal.
func checkLine(t *testing.T, what string, g *gfx, color byte, x1, y1, x2, y2 int) {
	t.Helper()
	dx, dy := x2-x1, y2-y1
	long := max(abs(dx), abs(dy))
	var pts [][2]int
	for y := 0; y < g.h; y++ {
		for x := 0; x < g.w; x++ {
			if g.get(x, y) == color {
				pts = append(pts, [2]int{x, y})
			}
		}
	}
	if len(pts) != long+1 {
		t.Fatalf("%s (%d,%d)-(%d,%d): %d pontos, quer %d", what, x1, y1, x2, y2, len(pts), long+1)
	}
	has := func(x, y int) bool { return g.get(x, y) == color }
	if !has(x1, y1) || !has(x2, y2) {
		t.Fatalf("%s (%d,%d)-(%d,%d): faltam extremos", what, x1, y1, x2, y2)
	}
	for _, p := range pts {
		// distância à reta ideal, no eixo menor: |(y-y1)*dx - (x-x1)*dy| * 2 <= long * ... (normalizada)
		var e int
		if abs(dx) >= abs(dy) {
			e = abs((p[1]-y1)*dx-(p[0]-x1)*dy) * 2
			if e > abs(dx) {
				t.Fatalf("%s (%d,%d)-(%d,%d): ponto (%d,%d) longe da reta", what, x1, y1, x2, y2, p[0], p[1])
			}
		} else {
			e = abs((p[0]-x1)*dy-(p[1]-y1)*dx) * 2
			if e > abs(dy) {
				t.Fatalf("%s (%d,%d)-(%d,%d): ponto (%d,%d) longe da reta", what, x1, y1, x2, y2, p[0], p[1])
			}
		}
		if p[0] < min(x1, x2) || p[0] > max(x1, x2) || p[1] < min(y1, y2) || p[1] > max(y1, y2) {
			t.Fatalf("%s: ponto (%d,%d) fora da caixa dos extremos", what, p[0], p[1])
		}
	}
}

func TestVDPHwLineAndBox(t *testing.T) {
	r, v := cmdRunner(t, "VDP_HwLine", "VDP_HwBox")
	rng := rand.New(rand.NewSource(11))
	for _, mode := range bitmapModes {
		setMode(r, mode)
		g := newGfx(v, mode)
		color := byte(1)
		if mode == 8 {
			color = 0xE3
		}
		lines := [][4]int{
			{10, 10, 10, 10}, {10, 10, 20, 10}, {20, 10, 10, 10}, {10, 10, 10, 20}, {10, 20, 10, 10}, {10, 10, 25, 25},
			{25, 25, 10, 10}, {25, 10, 10, 25}, {10, 25, 25, 10}, {0, 0, g.w - 1, g.h - 1}, {g.w - 1, 0, 0, g.h - 1},
		}
		for i := 0; i < 250; i++ {
			lines = append(lines, [4]int{rng.Intn(g.w), rng.Intn(g.h), rng.Intn(g.w), rng.Intn(g.h)})
		}
		for _, l := range lines {
			g.clear()
			put16(r, blk, l[0], l[1], l[2], l[3])
			in := canary()
			in.H, in.L, in.A, in.B = blk>>8, blk&0xFF, color, 0
			out := r.Call("VDP_HwLine", in)
			keep(t, "HwLine", in, out, "ABCDEHL")
			checkLine(t, "HwLine", g, color, l[0], l[1], l[2], l[3])
		}
		// contorno de retângulo: o conjunto exato de pontos do perímetro
		for i := 0; i < 30; i++ {
			g.clear()
			x1, y1, x2, y2 := 3+rng.Intn(g.w-6), 3+rng.Intn(g.h-6), 3+rng.Intn(g.w-6), 3+rng.Intn(g.h-6)
			put16(r, blk, x1, y1, x2, y2)
			in := canary()
			in.H, in.L, in.A, in.B = blk>>8, blk&0xFF, color, 0
			out := r.Call("VDP_HwBox", in)
			keep(t, "HwBox", in, out, "ABCDEHL")
			lx, hx, ly, hy := min(x1, x2), max(x1, x2), min(y1, y2), max(y1, y2)
			for y := 0; y < g.h; y++ {
				for x := 0; x < g.w; x++ {
					onEdge := x >= lx && x <= hx && y >= ly && y <= hy && (x == lx || x == hx || y == ly || y == hy)
					if (g.get(x, y) == color) != onEdge {
						t.Fatalf("HwBox (%d,%d)-(%d,%d): pixel (%d,%d) = %X", x1, y1, x2, y2, x, y, g.get(x, y))
					}
				}
			}
		}
	}
}

func TestVDPHwCopies(t *testing.T) {
	r, v := cmdRunner(t, "VDP_HwCopyRect", "VDP_HwMoveRect", "VDP_HwCopyLines")
	rng := rand.New(rand.NewSource(13))
	const dix, diy = 0x04, 0x08
	for _, mode := range bitmapModes {
		setMode(r, mode)
		g := newGfx(v, mode)
		g.clear()
		for y := 0; y < g.h; y++ {
			for x := 0; x < g.w; x++ {
				g.put(x, y, byte(rng.Intn(256)))
			}
		}
		im := g.snapshot()
		// cópias sem sobreposição e com sobreposição nos 4 sentidos
		for i := 0; i < 60; i++ {
			w, h := 1+rng.Intn(30), 1+rng.Intn(20)
			sx, sy := 40+rng.Intn(g.w-90), 30+rng.Intn(g.h-70)
			dx, dy := sx+rng.Intn(25)-12, sy+rng.Intn(21)-10
			arg := 0
			ssx, ssy, ddx, ddy := sx, sy, dx, dy
			if dx > sx {
				arg |= dix
				ssx, ddx = sx+w-1, dx+w-1
			}
			if dy > sy {
				arg |= diy
				ssy, ddy = sy+h-1, dy+h-1
			}
			op := []byte{0, 0, 0, 2, 3, 8}[rng.Intn(6)]
			put16(r, blk, ssx, ssy, ddx, ddy, w, h)
			in := canary()
			in.H, in.L, in.A, in.B = blk>>8, blk&0xFF, byte(arg), op
			out := r.Call("VDP_HwCopyRect", in)
			keep(t, "HwCopyRect", in, out, "ABCDEHL")
			// modelo: copia "como se a origem fosse lida antes de qualquer escrita" só vale sem
			// sobreposição; com sobreposição e sentido certo o resultado é o mesmo
			src := make([][]byte, h)
			for j := 0; j < h; j++ {
				src[j] = make([]byte, w)
				for k := 0; k < w; k++ {
					src[j][k] = im.pix[sy+j][sx+k]
				}
			}
			for j := 0; j < h; j++ {
				for k := 0; k < w; k++ {
					im.set(dx+k, dy+j, src[j][k], op)
				}
			}
			im.compare(t, "HwCopyRect")
		}
		// cópia por bytes
		pp := g.pixelsPerBy
		for i := 0; i < 20; i++ {
			w, h := pp*(1+rng.Intn(20/pp)), 1+rng.Intn(15)
			sx, sy := pp*(20+rng.Intn(20)), 30+rng.Intn(50)
			dx, dy := pp*(60+rng.Intn(20)), 100+rng.Intn(50)
			put16(r, blk, sx, sy, dx, dy, w, h)
			in := canary()
			in.H, in.L, in.A = blk>>8, blk&0xFF, 0
			out := r.Call("VDP_HwMoveRect", in)
			keep(t, "HwMoveRect", in, out, "ABCDEHL")
			for j := 0; j < h; j++ {
				for k := 0; k < w; k++ {
					im.pix[dy+j][dx+k] = im.pix[sy+j][sx+k]
				}
			}
			im.compare(t, "HwMoveRect")
		}
		// linhas inteiras a partir de uma coluna
		put16(r, blk, 20, 2*pp, 5, 4) // SY, DX, DY, linhas
		in := canary()
		in.H, in.L, in.A = blk>>8, blk&0xFF, 0
		out := r.Call("VDP_HwCopyLines", in)
		keep(t, "HwCopyLines", in, out, "ABCDEHL")
		for j := 0; j < 4; j++ {
			for x := 2 * pp; x < g.w; x++ {
				im.pix[5+j][x] = im.pix[20+j][x]
			}
		}
		im.compare(t, "HwCopyLines")
	}
}

func TestVDPHwSearch(t *testing.T) {
	r, v := cmdRunner(t, "VDP_HwSearch")
	for _, mode := range bitmapModes {
		setMode(r, mode)
		g := newGfx(v, mode)
		g.clear()
		border := byte(3)
		for x := 0; x < 40; x++ {
			g.put(x, 20, 1)
		}
		g.put(40, 20, border)
		call := func(x, y int, color byte, args byte) libtest.Regs {
			in := canary()
			in.B, in.C, in.D, in.E, in.A, in.H = byte(x>>8), byte(x), byte(y>>8), byte(y), color, args
			out := r.Call("VDP_HwSearch", in)
			keep(t, "HwSearch", in, out, "BCDE")
			return out
		}
		out := call(5, 20, border, 0)
		if !out.Carry() || out.HL() != 40 {
			t.Fatalf("modo %d: procura a borda: carry=%v X=%d", mode, out.Carry(), out.HL())
		}
		out = call(5, 20, 1, 0x02) // até o primeiro ponto diferente de 1
		if !out.Carry() || out.HL() != 40 {
			t.Fatalf("modo %d: EQ: carry=%v X=%d", mode, out.Carry(), out.HL())
		}
		out = call(30, 20, border, 0x04) // para a esquerda: não acha
		if out.Carry() {
			t.Fatalf("modo %d: achou o que não existe à esquerda", mode)
		}
		g.put(3, 20, border)
		out = call(30, 20, border, 0x04)
		if !out.Carry() || out.HL() != 3 {
			t.Fatalf("modo %d: para a esquerda: carry=%v X=%d", mode, out.Carry(), out.HL())
		}
	}
}

func TestVDPHwTransfers(t *testing.T) {
	r, v := cmdRunner(t, "VDP_HwLoadRect", "VDP_HwLoadFast", "VDP_HwReadRect")
	rng := rand.New(rand.NewSource(17))
	const data, back = 0xA000, 0xB000
	for _, mode := range bitmapModes {
		setMode(r, mode)
		g := newGfx(v, mode)
		g.clear()
		im := g.snapshot()
		for i := 0; i < 15; i++ {
			x, y := rng.Intn(g.w-30), rng.Intn(g.h-20)
			w, h := 1+rng.Intn(29), 1+rng.Intn(19)
			op := []byte{0, 0, 2, 3, 8}[rng.Intn(5)]
			buf := make([]byte, w*h)
			for j := range buf {
				buf[j] = byte(rng.Intn(256))
			}
			copy(r.M.Mem[data:], buf)
			put16(r, blk, x, y, w, h)
			in := canary()
			in.H, in.L, in.D, in.E, in.B, in.C, in.A = blk>>8, blk&0xFF, data>>8, data&0xFF, byte(len(buf)>>8), byte(len(buf)), op
			out := r.Call("VDP_HwLoadRect", in)
			keep(t, "HwLoadRect", in, out, "ABCDEHL")
			for j := 0; j < h; j++ {
				for k := 0; k < w; k++ {
					im.set(x+k, y+j, buf[j*w+k], op)
				}
			}
			im.compare(t, "HwLoadRect")
			if v.Reg[15] != 0 {
				t.Fatalf("HwLoadRect deixou R#15 = %d", v.Reg[15])
			}

			// lê de volta com LMCM
			for j := range r.M.Mem[back : back+len(buf)] {
				r.M.Mem[back+j] = 0xEE
			}
			in = canary()
			in.H, in.L, in.D, in.E, in.B, in.C = blk>>8, blk&0xFF, back>>8, back&0xFF, byte(len(buf)>>8), byte(len(buf))
			out = r.Call("VDP_HwReadRect", in)
			keep(t, "HwReadRect", in, out, "ABCDEHL")
			for j := 0; j < h; j++ {
				for k := 0; k < w; k++ {
					if got := r.M.Mem[back+j*w+k]; got != im.pix[y+j][x+k] {
						t.Fatalf("modo %d HwReadRect (%d,%d) = %X, quer %X", mode, k, j, got, im.pix[y+j][x+k])
					}
				}
			}
			if v.Reg[15] != 0 {
				t.Fatalf("HwReadRect deixou R#15 = %d", v.Reg[15])
			}
			if v.Status[2]&0x81 != 0 {
				t.Fatalf("HwReadRect deixou o motor ocupado: S#2 = %02X", v.Status[2])
			}
		}

		// HMMC: bytes crus
		pp := g.pixelsPerBy
		x, y, w, h := pp*10, 100, pp*6, 4
		buf := make([]byte, w/pp*h)
		for j := range buf {
			buf[j] = byte(rng.Intn(256))
		}
		copy(r.M.Mem[data:], buf)
		put16(r, blk, x, y, w, h)
		in := canary()
		in.H, in.L, in.D, in.E, in.B, in.C = blk>>8, blk&0xFF, data>>8, data&0xFF, byte(len(buf)>>8), byte(len(buf))
		out := r.Call("VDP_HwLoadFast", in)
		keep(t, "HwLoadFast", in, out, "ABCDEHL")
		for j := 0; j < h; j++ {
			for k := 0; k < w/pp; k++ {
				if got := v.VRAM[(y+j)*g.bpl+x/pp+k]; got != buf[j*(w/pp)+k] {
					t.Fatalf("modo %d HwLoadFast byte (%d,%d) = %02X, quer %02X", mode, k, j, got, buf[j*(w/pp)+k])
				}
			}
		}

		// quantidade zero: não faz nada
		n := v.Commands
		in = canary()
		in.H, in.L, in.D, in.E, in.B, in.C = blk>>8, blk&0xFF, data>>8, data&0xFF, 0, 0
		r.Call("VDP_HwLoadRect", in)
		r.Call("VDP_HwReadRect", in)
		if v.Commands != n {
			t.Fatal("transferência de 0 bytes disparou um comando")
		}

		// leitura de menos bytes que a área: o resto do comando é cancelado
		put16(r, blk, 0, 0, 10, 10)
		in = canary()
		in.H, in.L, in.D, in.E, in.B, in.C = blk>>8, blk&0xFF, back>>8, back&0xFF, 0, 7
		r.Call("VDP_HwReadRect", in)
		if v.Status[2]&0x81 != 0 {
			t.Fatalf("leitura parcial deixou o motor ocupado: S#2 = %02X", v.Status[2])
		}
	}
}
