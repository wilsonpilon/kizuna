package msxlib_test

import (
	"testing"

	"github.com/wilsonpilon/kizuna/pkg/libtest"
	"github.com/wilsonpilon/kizuna/pkg/z80sim"
)

// ---------------------------------------------------------------- tabelas

// refTableRegs é a referência (escrita direto da folha de dados, sem olhar o
// assembly): quais registradores cada tabela grava, dado o modo (0..9) e o
// endereço de 17 bits.
func refTableRegs(table string, mode int, addr uint32) map[int]byte {
	bitmap := mode >= 5 && mode <= 8
	switch table {
	case "name":
		shift, or := uint(10), byte(0)
		switch {
		case mode == 9:
			or = 0x03
		case bitmap:
			or = 0x1F
			if mode >= 7 {
				shift = 11
			}
		}
		return map[int]byte{2: byte(addr>>shift) | or}
	case "pattern":
		or := byte(0)
		if mode == 2 || mode == 4 {
			or = 0x03
		}
		return map[int]byte{4: byte(addr>>11) | or}
	case "color":
		or := byte(0)
		switch mode {
		case 2, 4:
			or = 0x7F
		case 9:
			or = 0x07
		}
		return map[int]byte{3: byte(addr>>6) | or, 10: byte(addr >> 14 & 7)}
	case "attr":
		or := byte(0)
		if mode >= 4 && mode <= 8 {
			or = 0x03
		}
		return map[int]byte{5: byte(addr>>7) | or, 11: byte(addr >> 15 & 3)}
	case "sprpat":
		return map[int]byte{6: byte(addr >> 11)}
	}
	panic(table)
}

// refTableGet: o que o getter devolve depois de um setter (os bits de máscara de cada modo somem).
func refTableGet(table string, mode int, addr uint32) uint32 {
	bitmap := mode >= 5 && mode <= 8
	g2 := mode == 2 || mode == 4
	switch table {
	case "name":
		switch {
		case mode == 9:
			return addr & 0x1F000 // R#2 bits 6..2 = A16..A12
		case bitmap && mode <= 6:
			return addr & 0x18000 // A16, A15
		case bitmap:
			return addr & 0x10000 // A16
		}
		return addr & 0x1FC00
	case "pattern":
		if g2 {
			return addr & 0x1E000 // R#4 bits 5..2 = A16..A13
		}
		return addr & 0x1F800
	case "color":
		switch {
		case g2:
			return addr & 0x1E000 // R#10 = A16..A14, R#3 bit 7 = A13
		case mode == 9:
			return addr & 0x1FE00 // R#3 bits 7..3 = A13..A9
		}
		return addr & 0x1FFC0
	case "attr":
		if mode >= 4 && mode <= 8 {
			return addr & 0x1FE00 // R#5 bits 1..0 forçados a 1
		}
		return addr & 0x1FF80
	case "sprpat":
		return addr & 0x1F800
	}
	panic(table)
}

var tableFns = []struct{ name, set, get string }{
	{"name", "VDP_SetNameTable", "VDP_GetNameTable"},
	{"pattern", "VDP_SetPatternTable", "VDP_GetPatternTable"},
	{"color", "VDP_SetColorTable", "VDP_GetColorTable"},
	{"attr", "VDP_SetSpriteAttrTable", "VDP_GetSpriteAttrTable"},
	{"sprpat", "VDP_SetSpritePatternTable", "VDP_GetSpritePatternTable"},
}

func setMode(r *libtest.Runner, mode int) {
	in := canary()
	in.A = byte(mode)
	r.Call("VDP_SetMode", in)
}

func TestVDPTableSetters(t *testing.T) {
	r, v := vdpRunner(t, "VDP_SetMode", "VDP_SetNameTable", "VDP_SetPatternTable", "VDP_SetColorTable",
		"VDP_SetSpriteAttrTable", "VDP_SetSpritePatternTable", "VDP_GetNameTable", "VDP_GetPatternTable",
		"VDP_GetColorTable", "VDP_GetSpriteAttrTable", "VDP_GetSpritePatternTable")
	align := map[string]uint32{"name": 0x400, "pattern": 0x800, "color": 0x40, "attr": 0x80, "sprpat": 0x800}
	for mode := 0; mode <= 9; mode++ {
		setMode(r, mode)
		for _, tb := range tableFns {
			step := align[tb.name]
			// todos os múltiplos de 2 KB (e, para as tabelas de granularidade menor, também os intermediários
			// da faixa baixa)
			var addrs []uint32
			for a := uint32(0); a < 1<<17; a += 0x800 {
				addrs = append(addrs, a)
			}
			for a := uint32(0); a < 0x4000; a += step {
				addrs = append(addrs, a)
			}
			addrs = append(addrs, 0x1FFC0&^(step-1), 0x10000+step, 0x1FC00)
			for _, addr := range addrs {
				addr &^= step - 1
				in := canary()
				in.A, in.H, in.L = byte(addr>>16), byte(addr>>8), byte(addr)
				out := r.Call(tb.set, in)
				for reg, want := range refTableRegs(tb.name, mode, addr) {
					if v.Reg[reg] != want {
						t.Fatalf("%s(modo %d, %05X): R#%d = %02X, quer %02X", tb.set, mode, addr, reg, v.Reg[reg], want)
					}
					if got := shadowOf(t, r, reg); got != want {
						t.Fatalf("%s(modo %d, %05X): sombra de R#%d = %02X, quer %02X", tb.set, mode, addr, reg, got, want)
					}
				}
				keep(t, tb.set, in, out, "ABCDEHL")

				got := r.Call(tb.get, canary())
				gaddr := uint32(got.A&1)<<16 | uint32(got.HL())
				if want := refTableGet(tb.name, mode, addr); gaddr != want {
					t.Fatalf("%s depois de %s(modo %d, %05X) = %05X, quer %05X", tb.get, tb.set, mode, addr, gaddr, want)
				}
				keep(t, tb.get, canary(), got, "BCDE")
			}
		}
	}
}

// Depois de VDP_SetMode os getters têm que devolver os endereços padrão do MSX-BASIC.
func TestVDPTableGettersAfterSetMode(t *testing.T) {
	r, _ := vdpRunner(t, "VDP_SetMode", "VDP_GetNameTable", "VDP_GetPatternTable", "VDP_GetColorTable",
		"VDP_GetSpriteAttrTable", "VDP_GetSpritePatternTable")
	// nome, padrão, cor, atributos de sprite, padrões de sprite
	want := [10][5]uint32{
		0: {0x0000, 0x0800, 0x0000, 0x0000, 0x0000},
		1: {0x1800, 0x0000, 0x2000, 0x1B00, 0x3800},
		2: {0x1800, 0x0000, 0x2000, 0x1B00, 0x3800},
		3: {0x0800, 0x0000, 0x0000, 0x1B00, 0x3800},
		4: {0x1800, 0x0000, 0x2000, 0x1E00, 0x3800},
		5: {0x0000, 0x0000, 0x0000, 0x7600, 0x7800},
		6: {0x0000, 0x0000, 0x0000, 0x7600, 0x7800},
		7: {0x0000, 0x0000, 0x0000, 0xFA00, 0xF000},
		8: {0x0000, 0x0000, 0x0000, 0xFA00, 0xF000},
		9: {0x0000, 0x1000, 0x0800, 0x0000, 0x0000},
	}
	for mode := 0; mode <= 9; mode++ {
		setMode(r, mode)
		for i, tb := range tableFns {
			got := r.Call(tb.get, canary())
			gaddr := uint32(got.A&1)<<16 | uint32(got.HL())
			if gaddr != want[mode][i] {
				t.Errorf("modo %d: %s = %05X, quer %05X", mode, tb.get, gaddr, want[mode][i])
			}
		}
	}
}

func TestVDPIsBitmapMode(t *testing.T) {
	r, _ := vdpRunner(t, "VDP_SetMode", "VDP_IsBitmapMode")
	for mode := 0; mode <= 9; mode++ {
		setMode(r, mode)
		in := canary()
		out := r.Call("VDP_IsBitmapMode", in)
		want := byte(0)
		if mode >= 5 && mode <= 8 {
			want = 1
		}
		if out.A != want {
			t.Errorf("IsBitmapMode(modo %d) = %d, quer %d", mode, out.A, want)
		}
		keep(t, "IsBitmapMode", in, out, "BCDEHL")
	}
}

// ---------------------------------------------------------------- sprites

func fillVRAM(v *z80sim.VDP, b byte) {
	for i := range v.VRAM {
		v.VRAM[i] = b
	}
}

// vramDiff devolve os endereços cujo valor difere de base, exceto os de mudanças permitidas.
func vramDiff(v *z80sim.VDP, base byte) map[int]byte {
	d := map[int]byte{}
	for i, b := range v.VRAM {
		if b != base {
			d[i] = b
		}
	}
	return d
}

func expectVRAM(t *testing.T, what string, v *z80sim.VDP, base byte, want map[int]byte) {
	t.Helper()
	got := vramDiff(v, base)
	for a, b := range want {
		if b == base {
			delete(got, a)
			continue
		}
		if g, ok := got[a]; !ok || g != b {
			t.Fatalf("%s: VRAM[%05X] = %02X, quer %02X", what, a, v.VRAM[a], b)
		}
		delete(got, a)
	}
	if len(got) != 0 {
		for a, b := range got {
			t.Fatalf("%s: VRAM[%05X] = %02X mudou sem querer (%d bytes a mais)", what, a, b, len(got))
		}
	}
}

var spriteNames = []string{"VDP_SetMode", "VDP_SetSpriteAttrTable", "VDP_SetSpritePatternTable", "VDP_SpriteSetPos",
	"VDP_SpriteSetY", "VDP_SpriteSetX", "VDP_SpriteSetPattern", "VDP_SpriteSetColor", "VDP_SpriteSetLineColors",
	"VDP_SpriteSetAll", "VDP_SpriteDisableFrom", "VDP_SpritePatternLoad", "VDP_SpriteAttrAddr", "VDP_SpriteColorAddr"}

func TestVDPSpritesMode1(t *testing.T) {
	r, v := vdpRunner(t, spriteNames...)
	setMode(r, 1) // atributos em 1B00h
	const attr = 0x1B00
	const base = 0x77
	fillVRAM(v, base)

	call := func(fn string, in libtest.Regs, keepRegs string) {
		t.Helper()
		out := r.Call(fn, in)
		keep(t, fn, in, out, keepRegs)
	}
	in := canary()
	in.A, in.H, in.L = 3, 0x55, 0x66
	call("VDP_SpriteSetPos", in, "ABCDEHL")
	expectVRAM(t, "SetPos", v, base, map[int]byte{attr + 12: 0x55, attr + 13: 0x66})

	in = canary()
	in.A, in.B = 3, 0x11
	call("VDP_SpriteSetY", in, "ABCDEHL")
	in = canary()
	in.A, in.B = 3, 0x22
	call("VDP_SpriteSetX", in, "ABCDEHL")
	in = canary()
	in.A, in.B = 3, 0x33
	call("VDP_SpriteSetPattern", in, "ABCDEHL")
	in = canary()
	in.A, in.B = 3, 0x8F
	call("VDP_SpriteSetColor", in, "ABCDEHL")
	expectVRAM(t, "SetY/X/Pattern/Color", v, base, map[int]byte{attr + 12: 0x11, attr + 13: 0x22, attr + 14: 0x33, attr + 15: 0x8F})

	// SetAll: 5 parâmetros, 4 bytes de atributo
	in = canary()
	in.A, in.H, in.L, in.D, in.E = 31, 0xA1, 0xA2, 0xA3, 0x0C
	call("VDP_SpriteSetAll", in, "ABCDEHL")
	expectVRAM(t, "SetAll", v, base, map[int]byte{attr + 12: 0x11, attr + 13: 0x22, attr + 14: 0x33, attr + 15: 0x8F,
		attr + 124: 0xA1, attr + 125: 0xA2, attr + 126: 0xA3, attr + 127: 0x0C})

	// SetLineColors não faz nada no modo 1
	in = canary()
	in.A = 4
	in.H, in.L = 0x90, 0x00
	call("VDP_SpriteSetLineColors", in, "ABCDEHL")

	// fim de lista: Y = D0h; índice >= 32 não escreve
	in = canary()
	in.A = 6
	call("VDP_SpriteDisableFrom", in, "ABCDEHL")
	before := vramDiff(v, base)
	in = canary()
	in.A = 32
	call("VDP_SpriteDisableFrom", in, "ABCDEHL")
	if len(vramDiff(v, base)) != len(before) {
		t.Fatal("DisableFrom(32) escreveu na VRAM")
	}
	if v.VRAM[attr+24] != 0xD0 {
		t.Fatalf("DisableFrom(6) modo 1: Y = %02X, quer D0", v.VRAM[attr+24])
	}

	// padrões: tabela em 3800h, padrão 4 => +32
	data := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
	copy(r.M.Mem[0x9000:], data)
	in = canary()
	in.A = 4
	in.H, in.L = 0x90, 0x00
	in.B, in.C = 0, byte(len(data))
	call("VDP_SpritePatternLoad", in, "ABCDEHL")
	for i, b := range data {
		if v.VRAM[0x3800+32+i] != b {
			t.Fatalf("PatternLoad: VRAM[%04X] = %02X, quer %02X", 0x3800+32+i, v.VRAM[0x3800+32+i], b)
		}
	}
	if v.VRAM[0x3800+32-1] != base || v.VRAM[0x3800+32+len(data)] != base {
		t.Fatal("PatternLoad escreveu fora do intervalo")
	}
}

func TestVDPSpritesMode2(t *testing.T) {
	r, v := vdpRunner(t, spriteNames...)
	const base = 0x77
	call := func(fn string, in libtest.Regs) {
		t.Helper()
		out := r.Call(fn, in)
		keep(t, fn, in, out, "ABCDEHL")
	}

	// tabelas: SCREEN 5 (atributos 7600h, cores 7400h), SCREEN 7 (FA00h/F800h) e duas
	// posições que atravessam a fronteira de 64 KB
	cases := []struct {
		mode   int
		setup  uint32
		custom bool // false = tabela padrão do modo
	}{{5, 0, false}, {7, 0, false}, {7, 0x10400, true}, {7, 0x10000, true}, {8, 0x10200, true}, {4, 0, false}, {5, 0x200, true}}
	for _, c := range cases {
		setMode(r, c.mode)
		if c.custom {
			in := canary()
			in.A, in.H, in.L = byte(c.setup>>16), byte(c.setup>>8), byte(c.setup)
			call("VDP_SetSpriteAttrTable", in)
		}
		fillVRAM(v, base)
		attr := map[int]uint32{4: 0x1E00, 5: 0x7600, 7: 0xFA00}[c.mode]
		if c.custom {
			attr = c.setup
		}
		color := (attr - 0x200) & 0x1FFFF
		want := map[int]byte{}

		// posição/padrão vão para os atributos; a cor para a tabela de cores
		in := canary()
		in.A, in.H, in.L = 2, 0x10, 0x20
		call("VDP_SpriteSetPos", in)
		want[int(attr)+8], want[int(attr)+9] = 0x10, 0x20
		in = canary()
		in.A, in.B = 2, 0x44
		call("VDP_SpriteSetPattern", in)
		want[int(attr)+10] = 0x44
		in = canary()
		in.A, in.B = 2, 0x4B
		call("VDP_SpriteSetColor", in)
		for i := 0; i < 16; i++ {
			want[int(color)+2*16+i] = 0x4B
		}
		expectVRAM(t, "modo 2 SetPos/Pattern/Color", v, base, want)

		// cores por linha
		lines := make([]byte, 16)
		for i := range lines {
			lines[i] = byte(0x40 + i)
		}
		copy(r.M.Mem[0x9000:], lines)
		in = canary()
		in.A, in.H, in.L = 31, 0x90, 0x00
		call("VDP_SpriteSetLineColors", in)
		for i := range lines {
			want[int(color)+31*16+i] = lines[i]
		}
		expectVRAM(t, "modo 2 SetLineColors", v, base, want)

		// SetAll: o 4º byte do atributo fica 0 e a cor vai para a tabela de cores
		in = canary()
		in.A, in.H, in.L, in.D, in.E = 0, 0xC1, 0xC2, 0xC3, 0x0A
		call("VDP_SpriteSetAll", in)
		want[int(attr)+0], want[int(attr)+1], want[int(attr)+2], want[int(attr)+3] = 0xC1, 0xC2, 0xC3, 0
		for i := 0; i < 16; i++ {
			want[int(color)+i] = 0x0A
		}
		expectVRAM(t, "modo 2 SetAll", v, base, want)

		// fim de lista D8h
		in = canary()
		in.A = 9
		call("VDP_SpriteDisableFrom", in)
		want[int(attr)+36] = 0xD8
		expectVRAM(t, "modo 2 DisableFrom", v, base, want)

		// os helpers de endereço
		for idx := 0; idx < 32; idx += 7 {
			in := canary()
			in.A = byte(idx)
			out := r.Call("VDP_SpriteAttrAddr", in)
			if got, w := uint32(out.A&1)<<16|uint32(out.HL()), (attr+uint32(idx)*4)&0x1FFFF; got != w {
				t.Fatalf("SpriteAttrAddr(%d) = %05X, quer %05X", idx, got, w)
			}
			keep(t, "SpriteAttrAddr", in, out, "BCDE")
			out = r.Call("VDP_SpriteColorAddr", in)
			if got, w := uint32(out.A&1)<<16|uint32(out.HL()), (color+uint32(idx)*16)&0x1FFFF; got != w {
				t.Fatalf("SpriteColorAddr(%d) = %05X, quer %05X", idx, got, w)
			}
			keep(t, "SpriteColorAddr", in, out, "BCDE")
		}
	}

	// padrões de sprite acima de 64 KB (SCREEN 7 com a tabela em 1F800h)
	setMode(r, 7)
	in := canary()
	in.A, in.H, in.L = 1, 0xF8, 0x00
	call("VDP_SetSpritePatternTable", in)
	fillVRAM(v, base)
	data := []byte{9, 8, 7, 6}
	copy(r.M.Mem[0x9000:], data)
	in = canary()
	in.A = 255 // padrão 255 => 1F800h + 2040 = 1FFF8h; 4 bytes cruzam o fim da VRAM
	in.H, in.L = 0x90, 0x00
	in.B, in.C = 0, 4
	call("VDP_SpritePatternLoad", in)
	for i, b := range data {
		if v.VRAM[0x1F800+255*8+i] != b {
			t.Fatalf("PatternLoad acima de 64K: VRAM[%05X] = %02X", 0x1F800+255*8+i, v.VRAM[0x1F800+255*8+i])
		}
	}
}

// ---------------------------------------------------------------- piscar

func TestVDPBlink(t *testing.T) {
	r, v := vdpRunner(t, "VDP_SetMode", "VDP_SetColorTable", "VDP_BlinkFill", "VDP_BlinkLine", "VDP_BlinkCell",
		"VDP_SetBlinkColor", "VDP_SetBlinkTime", "VDP_BlinkOff")
	setMode(r, 9)
	const tab = 0x0800
	const base = 0x5A
	call := func(fn string, in libtest.Regs) {
		t.Helper()
		out := r.Call(fn, in)
		keep(t, fn, in, out, "ABCDEHL")
	}
	// registradores
	in := canary()
	in.A, in.B = 0x0F, 0x04
	call("VDP_SetBlinkColor", in)
	if v.Reg[12] != 0xF4 {
		t.Errorf("SetBlinkColor: R#12 = %02X, quer F4", v.Reg[12])
	}
	in = canary()
	in.A, in.B = 0xF3, 0x12 // só os 4 bits baixos de cada um
	call("VDP_SetBlinkColor", in)
	if v.Reg[12] != 0x32 {
		t.Errorf("SetBlinkColor mascara: R#12 = %02X, quer 32", v.Reg[12])
	}
	in = canary()
	in.A, in.B = 0x07, 0x09
	call("VDP_SetBlinkTime", in)
	if v.Reg[13] != 0x79 {
		t.Errorf("SetBlinkTime: R#13 = %02X, quer 79", v.Reg[13])
	}
	call("VDP_BlinkOff", canary())
	if v.Reg[13] != 0 || shadowOf(t, r, 13) != 0 {
		t.Errorf("BlinkOff: R#13 = %02X", v.Reg[13])
	}

	// tabela em 0800h
	fillVRAM(v, base)
	in = canary()
	in.A, in.H, in.L = 0, 0x08, 0x00
	call("VDP_SetColorTable", in)
	in = canary()
	in.A = 0xFF
	call("VDP_BlinkFill", in)
	want := map[int]byte{}
	for i := 0; i < 270; i++ {
		want[tab+i] = 0xFF
	}
	expectVRAM(t, "BlinkFill", v, base, want)
	in = canary()
	in.A = 0
	call("VDP_BlinkFill", in)
	want = map[int]byte{}
	for i := 0; i < 270; i++ {
		want[tab+i] = 0
	}
	expectVRAM(t, "BlinkFill(0)", v, base, want)

	// uma linha
	in = canary()
	in.A, in.B = 3, 0xAA
	call("VDP_BlinkLine", in)
	for i := 0; i < 10; i++ {
		want[tab+30+i] = 0xAA
	}
	expectVRAM(t, "BlinkLine", v, base, want)

	// células: modelo em Go do bit certo (bit 7 = coluna mais à esquerda do grupo de 8)
	model := make([]byte, 270)
	for i := 0; i < 10; i++ {
		model[30+i] = 0xAA
	}
	type cell struct{ x, y, on byte }
	for _, c := range []cell{{13, 2, 1}, {0, 0, 1}, {79, 26, 1}, {7, 0, 1}, {8, 0, 1}, {13, 2, 0}, {1, 3, 0}, {0, 3, 0}, {77, 3, 1}, {1, 3, 1}} {
		in = canary()
		in.A, in.B, in.C = c.x, c.y, c.on
		call("VDP_BlinkCell", in)
		idx := int(c.y)*10 + int(c.x)/8
		bit := byte(0x80) >> (c.x & 7)
		if c.on != 0 {
			model[idx] |= bit
		} else {
			model[idx] &^= bit
		}
		w := map[int]byte{}
		for i, b := range model {
			w[tab+i] = b
		}
		expectVRAM(t, "BlinkCell", v, base, w)
	}
}

// ---------------------------------------------------------------- registradores diversos

func TestVDPModeBitsAndScroll(t *testing.T) {
	names := []string{"VDP_SetReg", "VDP_SpriteSize", "VDP_SpriteMag", "VDP_HBlankInt", "VDP_Interlace",
		"VDP_PageAlternate", "VDP_Transparency", "VDP_GrayScale", "VDP_LeftMask", "VDP_SetVerticalOffset",
		"VDP_SetHScrollCoarse", "VDP_SetHScrollFine", "VDP_SetAdjustRaw", "VDP_SetHBlankLine"}
	r, v := vdpRunner(t, names...)
	set := func(reg int, val byte) {
		in := canary()
		in.C, in.B = byte(reg), val
		r.Call("VDP_SetReg", in)
	}
	steps := []struct {
		fn            string
		a             byte
		reg           int
		before, after byte
	}{
		{"VDP_SpriteSize", 1, 1, 0xC0, 0xC2}, {"VDP_SpriteSize", 0, 1, 0xC3, 0xC1}, {"VDP_SpriteSize", 200, 1, 0x00, 0x02},
		{"VDP_SpriteMag", 5, 1, 0xC0, 0xC1}, {"VDP_SpriteMag", 0, 1, 0xC3, 0xC2},
		{"VDP_HBlankInt", 1, 0, 0x00, 0x10}, {"VDP_HBlankInt", 0, 0, 0x1E, 0x0E},
		{"VDP_Interlace", 1, 9, 0x80, 0x88}, {"VDP_Interlace", 0, 9, 0xFF, 0xF7},
		{"VDP_PageAlternate", 1, 9, 0x00, 0x04}, {"VDP_PageAlternate", 0, 9, 0xFF, 0xFB},
		{"VDP_Transparency", 1, 8, 0x2A, 0x0A}, {"VDP_Transparency", 0, 8, 0x0A, 0x2A},
		{"VDP_GrayScale", 1, 8, 0x00, 0x01}, {"VDP_GrayScale", 0, 8, 0xFF, 0xFE},
		{"VDP_LeftMask", 1, 25, 0x00, 0x02}, {"VDP_LeftMask", 0, 25, 0xFF, 0xFD},
		{"VDP_SetVerticalOffset", 0x9C, 23, 0x00, 0x9C},
		{"VDP_SetHScrollCoarse", 0xFF, 26, 0x00, 0x3F}, {"VDP_SetHScrollCoarse", 0x15, 26, 0xFF, 0x15},
		{"VDP_SetHScrollFine", 0xFF, 27, 0x00, 0x07}, {"VDP_SetHScrollFine", 0x03, 27, 0xFF, 0x03},
		{"VDP_SetAdjustRaw", 0x5A, 18, 0x00, 0x5A},
		{"VDP_SetHBlankLine", 77, 19, 0x00, 77},
	}
	for _, s := range steps {
		set(s.reg, s.before)
		in := canary()
		in.A = s.a
		out := r.Call(s.fn, in)
		if v.Reg[s.reg] != s.after || shadowOf(t, r, s.reg) != s.after {
			t.Fatalf("%s(%d): R#%d = %02X (sombra %02X), quer %02X", s.fn, s.a, s.reg, v.Reg[s.reg], shadowOf(t, r, s.reg), s.after)
		}
		keep(t, s.fn, in, out, "ABCDEHL")
	}
}

func TestVDPGetVersion(t *testing.T) {
	r, v := vdpRunner(t, "VDP_GetVersion")
	v.VBlankEvery = 0
	for _, c := range []struct{ s1, want byte }{{0x00, 0}, {0x04, 2}, {0x3E, 31}, {0xC1, 0}, {0x06, 3}} {
		v.Status[1] = c.s1
		in := canary()
		out := r.Call("VDP_GetVersion", in)
		if out.A != c.want {
			t.Errorf("S#1 = %02X: GetVersion = %d, quer %d", c.s1, out.A, c.want)
		}
		keep(t, "GetVersion", in, out, "BCDEHL")
		if v.Reg[15] != 0 {
			t.Errorf("GetVersion deixou R#15 = %d", v.Reg[15])
		}
	}
}
