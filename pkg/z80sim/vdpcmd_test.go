package z80sim

import "testing"

// cmdVDP devolve um VDP em GRAPHIC 4 (SCREEN 5) limpo.
func cmdVDP(mode byte) *VDP {
	v := NewVDP()
	v.VBlankEvery = 0
	r0 := map[byte]byte{0x0C: 0x06, 0x10: 0x08, 0x14: 0x0A, 0x1C: 0x0E}[mode]
	out99(v, r0, 0x80)
	return v
}

func setCmdRegs(v *VDP, sx, sy, dx, dy, nx, ny int, clr, arg, cmd byte) {
	w := func(reg int, val int) {
		v.writeReg(byte(reg), byte(val))
		v.writeReg(byte(reg+1), byte(val>>8))
	}
	v.Reg[44], v.Reg[45] = clr, arg
	w(32, sx)
	w(34, sy)
	w(36, dx)
	w(38, dy)
	w(40, nx)
	w(42, ny)
	v.writeReg(44, clr)
	v.writeReg(45, arg)
	v.writeReg(46, cmd)
}

func TestCmdPixelFormats(t *testing.T) {
	for _, c := range []struct {
		mode       byte
		w, bpp     int
		maxc       byte
		bpl        int
		addrOf     func(x, y int) int
		shiftOf    func(x int) uint
		xLimit     int
		colorLimit byte
	}{
		{0x0C, 256, 4, 15, 128, func(x, y int) int { return y*128 + x/2 }, func(x int) uint { return uint(4 - 4*(x&1)) }, 256, 15},
		{0x10, 512, 2, 3, 128, func(x, y int) int { return y*128 + x/4 }, func(x int) uint { return uint(6 - 2*(x&3)) }, 512, 3},
		{0x14, 512, 4, 15, 256, func(x, y int) int { return y*256 + x/2 }, func(x int) uint { return uint(4 - 4*(x&1)) }, 512, 15},
		{0x1C, 256, 8, 255, 256, func(x, y int) int { return y*256 + x }, func(x int) uint { return 0 }, 256, 255},
	} {
		v := cmdVDP(c.mode)
		if v.Mode() != c.mode {
			t.Fatalf("modo %05b", v.Mode())
		}
		for _, p := range [][2]int{{0, 0}, {1, 0}, {3, 5}, {c.w - 1, 0}, {c.w - 1, 211}, {7, 300}} {
			for i := range v.VRAM {
				v.VRAM[i] = 0
			}
			setCmdRegs(v, 0, 0, p[0], p[1], 0, 0, 0xFF, 0, cmdPSET<<4)
			want := byte(0xFF) & c.maxc
			got := v.VRAM[c.addrOf(p[0], p[1])] >> c.shiftOf(p[0]) & c.maxc
			if got != want {
				t.Fatalf("modo %02X PSET(%d,%d): pixel %02X, quer %02X", c.mode, p[0], p[1], got, want)
			}
			// só os bits desse pixel mudaram
			n := 0
			for _, b := range v.VRAM {
				if b != 0 {
					n++
				}
			}
			if n != 1 {
				t.Fatalf("modo %02X PSET escreveu %d bytes", c.mode, n)
			}
			setCmdRegs(v, p[0], p[1], 0, 0, 0, 0, 0, 0, cmdPOINT<<4)
			if v.Status[7] != want {
				t.Fatalf("POINT(%d,%d) = %02X, quer %02X", p[0], p[1], v.Status[7], want)
			}
		}
		// fora da tela: ignorado
		setCmdRegs(v, 0, 0, c.w, 0, 0, 0, 0xFF, 0, cmdPSET<<4)
	}
}

func TestCmdLogicalOps(t *testing.T) {
	v := cmdVDP(0x0C)
	for _, c := range []struct {
		op        byte
		dst, src  byte
		want      byte
		transp    bool
		wantTouch bool
	}{
		{0, 0x9, 0x6, 0x6, false, true}, {1, 0xC, 0xA, 0x8, false, true}, {2, 0xC, 0xA, 0xE, false, true},
		{3, 0xC, 0xA, 0x6, false, true}, {4, 0xC, 0xA, 0x5, false, true},
		{8, 0x9, 0x0, 0x9, true, false}, {8, 0x9, 0x6, 0x6, true, true}, {0xB, 0xC, 0x0, 0xC, true, false},
	} {
		for i := range v.VRAM {
			v.VRAM[i] = 0
		}
		v.pset(10, 4, c.dst, 0)
		setCmdRegs(v, 0, 0, 10, 4, 0, 0, c.src, 0, cmdPSET<<4|c.op)
		if got := v.pget(10, 4); got != c.want {
			t.Errorf("op %X dst %X src %X: %X, quer %X", c.op, c.dst, c.src, got, c.want)
		}
	}
}

func TestCmdRectangles(t *testing.T) {
	v := cmdVDP(0x0C)
	// LMMV
	setCmdRegs(v, 0, 0, 10, 20, 5, 3, 7, 0, cmdLMMV<<4)
	for y := 0; y < 30; y++ {
		for x := 0; x < 30; x++ {
			want := byte(0)
			if x >= 10 && x < 15 && y >= 20 && y < 23 {
				want = 7
			}
			if v.pget(x, y) != want {
				t.Fatalf("LMMV (%d,%d) = %X, quer %X", x, y, v.pget(x, y), want)
			}
		}
	}
	// LMMM: copia para outro lugar; com DIX/DIY para cima/esquerda a partir do canto oposto
	setCmdRegs(v, 10, 20, 100, 50, 5, 3, 0, 0, cmdLMMM<<4)
	for y := 0; y < 3; y++ {
		for x := 0; x < 5; x++ {
			if v.pget(100+x, 50+y) != 7 {
				t.Fatalf("LMMM (%d,%d)", x, y)
			}
		}
	}
	if v.pget(105, 50) != 0 || v.pget(99, 50) != 0 || v.pget(100, 53) != 0 {
		t.Fatal("LMMM escreveu fora do retângulo")
	}
	// sobreposição: fonte e destino se cruzam; com DIX o resultado é o de uma cópia "para trás"
	for x := 0; x < 8; x++ {
		v.pset(x, 100, byte(x+1), 0)
	}
	setCmdRegs(v, 7, 100, 9, 100, 8, 1, 0, 0x04, cmdLMMM<<4) // DIX: do último para o primeiro
	for x := 0; x < 8; x++ {
		if v.pget(2+x, 100) != byte(x+1) {
			t.Fatalf("LMMM sobreposto (DIX): pixel %d = %X", 2+x, v.pget(2+x, 100))
		}
	}
	// operação lógica no retângulo: XOR duas vezes restaura
	setCmdRegs(v, 0, 0, 10, 20, 5, 3, 0xF, 0, cmdLMMV<<4|3)
	if v.pget(10, 20) != 7^0xF {
		t.Fatalf("LMMV XOR = %X", v.pget(10, 20))
	}
	setCmdRegs(v, 0, 0, 10, 20, 5, 3, 0xF, 0, cmdLMMV<<4|3)
	if v.pget(10, 20) != 7 {
		t.Fatalf("LMMV XOR 2x = %X", v.pget(10, 20))
	}
}

func TestCmdHighSpeed(t *testing.T) {
	v := cmdVDP(0x0C) // 2 pixels por byte
	setCmdRegs(v, 0, 0, 8, 10, 6, 2, 0x5A, 0, cmdHMMV<<4)
	for y := 10; y < 12; y++ {
		for b := 3; b < 8; b++ {
			want := byte(0)
			if b >= 4 && b < 7 {
				want = 0x5A
			}
			if v.VRAM[y*128+b] != want {
				t.Fatalf("HMMV byte %d,%d = %02X, quer %02X", b, y, v.VRAM[y*128+b], want)
			}
		}
	}
	setCmdRegs(v, 8, 10, 40, 60, 6, 2, 0, 0, cmdHMMM<<4)
	for y := 60; y < 62; y++ {
		for b := 20; b < 23; b++ {
			if v.VRAM[y*128+b] != 0x5A {
				t.Fatalf("HMMM byte %d,%d = %02X", b, y, v.VRAM[y*128+b])
			}
		}
	}
	if v.VRAM[60*128+23] != 0 || v.VRAM[60*128+19] != 0 {
		t.Fatal("HMMM escreveu fora")
	}
	// YMMM: linha 10 -> 200, a partir de x = 64 até a borda direita
	v.VRAM[10*128+100] = 0x77
	v.VRAM[10*128+10] = 0x33
	setCmdRegs(v, 0, 10, 64, 200, 0, 1, 0, 0, cmdYMMM<<4)
	if v.VRAM[200*128+100] != 0x77 {
		t.Fatalf("YMMM: %02X", v.VRAM[200*128+100])
	}
	if v.VRAM[200*128+10] != 0 {
		t.Fatal("YMMM copiou a parte esquerda")
	}
}

func TestCmdLine(t *testing.T) {
	v := cmdVDP(0x0C)
	// horizontal: (10,5) a (20,5): NX = 10 => 11 pontos
	setCmdRegs(v, 0, 0, 10, 5, 10, 0, 9, 0, cmdLINE<<4)
	for x := 8; x < 23; x++ {
		want := byte(0)
		if x >= 10 && x <= 20 {
			want = 9
		}
		if v.pget(x, 5) != want {
			t.Fatalf("linha horizontal x=%d: %X", x, v.pget(x, 5))
		}
	}
	// diagonal (0,0)-(5,5): NX = 5, NY = 5
	for i := range v.VRAM {
		v.VRAM[i] = 0
	}
	setCmdRegs(v, 0, 0, 0, 0, 5, 5, 3, 0, cmdLINE<<4)
	for i := 0; i <= 5; i++ {
		if v.pget(i, i) != 3 {
			t.Fatalf("diagonal (%d,%d)", i, i)
		}
	}
	// (10,10)-(14,12): NX = 4, NY = 2 -- o exemplo do texto do modelo
	for i := range v.VRAM {
		v.VRAM[i] = 0
	}
	setCmdRegs(v, 0, 0, 10, 10, 4, 2, 5, 0, cmdLINE<<4)
	for _, p := range [][2]int{{10, 10}, {11, 11}, {12, 11}, {13, 12}, {14, 12}} {
		if v.pget(p[0], p[1]) != 5 {
			t.Fatalf("linha (%d,%d) não desenhada", p[0], p[1])
		}
	}
	// lado maior em Y (MAJ=1) e direções invertidas: (50,50) -> (48,44): NX = 6 (dy), NY = 2 (dx), DIX, DIY, MAJ
	for i := range v.VRAM {
		v.VRAM[i] = 0
	}
	setCmdRegs(v, 0, 0, 50, 50, 6, 2, 8, 0x01|0x04|0x08, cmdLINE<<4)
	n := 0
	for y := 40; y < 55; y++ {
		for x := 40; x < 55; x++ {
			if v.pget(x, y) != 0 {
				n++
			}
		}
	}
	if n != 7 || v.pget(50, 50) != 8 || v.pget(48, 44) != 8 {
		t.Fatalf("linha MAJ: %d pontos, extremos %X %X", n, v.pget(50, 50), v.pget(48, 44))
	}
}

func TestCmdSearch(t *testing.T) {
	v := cmdVDP(0x0C)
	for x := 0; x < 20; x++ {
		v.pset(x, 30, 1, 0)
	}
	v.pset(20, 30, 6, 0)
	// procura a cor 6 para a direita a partir de 3 (EQ = 0: pára quando igual)
	setCmdRegs(v, 3, 30, 0, 0, 0, 0, 6, 0, cmdSRCH<<4)
	if v.Status[2]&0x10 == 0 || v.Status[8] != 20 {
		t.Fatalf("SRCH: BD=%02X X=%d", v.Status[2], v.Status[8])
	}
	// EQ = 1: pára no primeiro pixel diferente de 1
	setCmdRegs(v, 3, 30, 0, 0, 0, 0, 1, 0x02, cmdSRCH<<4)
	if v.Status[2]&0x10 == 0 || v.Status[8] != 20 {
		t.Fatalf("SRCH EQ: %02X %d", v.Status[2], v.Status[8])
	}
	// para a esquerda, sem achar a cor 9: BD = 0
	setCmdRegs(v, 10, 30, 0, 0, 0, 0, 9, 0x04, cmdSRCH<<4)
	if v.Status[2]&0x10 != 0 {
		t.Fatal("SRCH sem achar deixou BD ligado")
	}
}

func TestCmdCPUTransfers(t *testing.T) {
	v := cmdVDP(0x0C)
	// LMMC 3x2: o primeiro dado vai junto, depois espera TR para cada um
	setCmdRegs(v, 0, 0, 5, 7, 3, 2, 1, 0, cmdLMMC<<4)
	if v.Status[2]&0x81 != 0x81 {
		t.Fatalf("depois do LMMC S#2 = %02X, quer CE e TR", v.Status[2])
	}
	for _, b := range []byte{2, 3, 4, 5} {
		v.writeReg(44, b)
	}
	if v.Status[2]&0x81 != 0x81 {
		t.Fatalf("ainda falta 1 byte: S#2 = %02X", v.Status[2])
	}
	v.writeReg(44, 6)
	if v.Status[2]&0x81 != 0 {
		t.Fatalf("terminou: S#2 = %02X", v.Status[2])
	}
	want := []byte{1, 2, 3, 4, 5, 6}
	for i, w := range want {
		if got := v.pget(5+i%3, 7+i/3); got != w {
			t.Fatalf("LMMC pixel %d = %X, quer %X", i, got, w)
		}
	}
	// LMCM: lê os mesmos 6 pixels de volta
	setCmdRegs(v, 5, 7, 0, 0, 3, 2, 0, 0, cmdLMCM<<4)
	out99(v, 0x07, 0x8F) // R#15 = 7
	for i, w := range want {
		if v.Status[2]&0x80 == 0 {
			t.Fatalf("TR baixo antes do pixel %d", i)
		}
		if got := v.Read99(); got != w {
			t.Fatalf("LMCM pixel %d = %X, quer %X", i, got, w)
		}
	}
	if v.Status[2]&0x81 != 0 {
		t.Fatalf("LMCM terminou com S#2 = %02X", v.Status[2])
	}
	// HMMC: 2 bytes por linha (4 pixels), 2 linhas
	setCmdRegs(v, 0, 0, 20, 40, 4, 2, 0x11, 0, cmdHMMC<<4)
	for _, b := range []byte{0x22, 0x33, 0x44} {
		v.writeReg(44, b)
	}
	for i, w := range []byte{0x11, 0x22, 0x33, 0x44} {
		if got := v.VRAM[(40+i/2)*128+10+i%2]; got != w {
			t.Fatalf("HMMC byte %d = %02X", i, got)
		}
	}
	// STOP interrompe
	setCmdRegs(v, 0, 0, 5, 7, 3, 2, 1, 0, cmdLMMC<<4)
	v.writeReg(46, 0)
	if v.Status[2]&0x81 != 0 {
		t.Fatal("STOP não cancelou")
	}
}

func TestCmdIgnoredOutsideBitmap(t *testing.T) {
	v := NewVDP() // modo 0: texto
	setCmdRegs(v, 0, 0, 5, 5, 3, 3, 7, 0, cmdLMMV<<4)
	if v.Commands != 0 {
		t.Fatal("comando aceito fora de um modo bitmap")
	}
	for _, b := range v.VRAM {
		if b != 0 {
			t.Fatal("VRAM alterada")
		}
	}
}
