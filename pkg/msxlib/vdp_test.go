package msxlib_test

import (
	"testing"

	"github.com/wilsonpilon/kizuna/pkg/libtest"
	"github.com/wilsonpilon/kizuna/pkg/z80sim"
)

// vdpRunner liga as rotinas pedidas e um modelo de VDP às portas 98h-9Bh.
func vdpRunner(t *testing.T, names ...string) (*libtest.Runner, *z80sim.VDP) {
	t.Helper()
	r := libtest.NewRunner(t, names...)
	return r, r.M.AttachVDP()
}

// shadowOf lê a cópia sombra de um registrador direto da memória do simulador.
func shadowOf(t *testing.T, r *libtest.Runner, reg int) byte {
	t.Helper()
	return r.M.Mem[int(r.Addr("VDP_Shadow"))+reg]
}

func TestVDPRegisterShadow(t *testing.T) {
	r, v := vdpRunner(t, "VDP_SetReg", "VDP_GetReg", "VDP_UpdateReg", "VDP_Shadow")
	for _, reg := range []int{0, 1, 2, 7, 8, 9, 15, 16, 23, 25, 27, 32, 44, 46, 63} {
		for _, val := range []byte{0x00, 0x5A, 0xFF} {
			if reg == 46 {
				continue // R#46 dispara um comando; testado com o motor de comandos
			}
			in := canary()
			in.C, in.B = byte(reg), val
			out := r.Call("VDP_SetReg", in)
			want := val
			if reg == 14 {
				want &= 7
			}
			if v.Reg[reg] != want {
				t.Fatalf("SetReg(R#%d, %02X): VDP tem %02X", reg, val, v.Reg[reg])
			}
			if shadowOf(t, r, reg) != val {
				t.Fatalf("SetReg(R#%d, %02X): copia sombra = %02X", reg, val, shadowOf(t, r, reg))
			}
			keep(t, "SetReg", in, out, "ABCDEHL")

			in = canary()
			in.A = byte(reg)
			out = r.Call("VDP_GetReg", in)
			if out.A != val {
				t.Fatalf("GetReg(R#%d) = %02X, quer %02X", reg, out.A, val)
			}
			keep(t, "GetReg", in, out, "BCDEHL")
		}
	}

	// UpdateReg: muda só os bits da máscara
	set := func(reg int, val byte) {
		in := canary()
		in.C, in.B = byte(reg), val
		r.Call("VDP_SetReg", in)
	}
	for _, c := range []struct{ old, mask, val, want byte }{
		{0xA5, 0x40, 0x40, 0xE5}, {0xE5, 0x40, 0x00, 0xA5}, {0xFF, 0x0F, 0x03, 0xF3}, {0x00, 0xF0, 0xFF, 0xF0},
		{0x12, 0x00, 0xFF, 0x12}, {0x12, 0xFF, 0x34, 0x34}, {0xAA, 0x55, 0xFF, 0xFF},
	} {
		set(1, c.old)
		in := canary()
		in.C, in.D, in.E = 1, c.mask, c.val
		out := r.Call("VDP_UpdateReg", in)
		if v.Reg[1] != c.want || shadowOf(t, r, 1) != c.want {
			t.Fatalf("UpdateReg(R#1, antes %02X, mascara %02X, valor %02X): VDP %02X sombra %02X, quer %02X",
				c.old, c.mask, c.val, v.Reg[1], shadowOf(t, r, 1), c.want)
		}
		keep(t, "UpdateReg", in, out, "ABCDEHL")
	}
}

func TestVDPReadStatus(t *testing.T) {
	r, v := vdpRunner(t, "VDP_ReadStatus")
	v.VBlankEvery = 0
	for n := 0; n <= 9; n++ {
		v.Status[n] = byte(0x10 + n)
	}
	for n := 0; n <= 9; n++ {
		in := canary()
		in.A = byte(n)
		v.Reg[15] = 0x07 // qualquer valor anterior
		out := r.Call("VDP_ReadStatus", in)
		want := byte(0x10 + n)
		if n == 0 {
			want = 0x10
		}
		if out.A != want {
			t.Fatalf("ReadStatus(%d) = %02X, quer %02X", n, out.A, want)
		}
		if v.Reg[15] != 0 {
			t.Fatalf("ReadStatus(%d) deixou R#15 = %d, quer 0", n, v.Reg[15])
		}
		keep(t, "ReadStatus", in, out, "BCDEHL")
	}
	// ler S#0 apaga F, 5S e C
	v.Status[0] = 0xE5
	in := canary()
	in.A = 0
	if out := r.Call("VDP_ReadStatus", in); out.A != 0xE5 {
		t.Errorf("S#0 = %02X", out.A)
	}
	if v.Status[0] != 0x05 {
		t.Errorf("S#0 depois da leitura = %02X, quer 05", v.Status[0])
	}
}

func TestVDPWaitVBlank(t *testing.T) {
	r, _ := vdpRunner(t, "VDP_WaitVBlank", "VDP_WaitFrames")
	r.M.JiffyEvery = 40
	jiffy := func() uint16 { return uint16(r.M.Mem[0xFC9E]) | uint16(r.M.Mem[0xFC9F])<<8 }
	before := jiffy()
	in := canary()
	out := r.Call("VDP_WaitVBlank", in)
	if jiffy() == before {
		t.Error("WaitVBlank voltou sem esperar um vblank")
	}
	keep(t, "WaitVBlank", in, out, "ABCDEHL")
	for _, n := range []byte{0, 1, 3, 10} {
		before = jiffy()
		in := canary()
		in.B = n
		out := r.Call("VDP_WaitFrames", in)
		if got := jiffy() - before; got < uint16(n) {
			t.Fatalf("WaitFrames(%d) esperou so %d quadros", n, got)
		}
		keep(t, "WaitFrames", in, out, "ABCDEHL")
	}
}

func TestVDPDisplayFlags(t *testing.T) {
	names := []string{"VDP_SetReg", "VDP_DisplayOn", "VDP_DisplayOff", "VDP_VBlankIntOn", "VDP_VBlankIntOff",
		"VDP_SpritesOn", "VDP_SpritesOff", "VDP_SetLines", "VDP_SetRefresh", "VDP_Backdrop", "VDP_TextColor", "VDP_SetDisplayPage"}
	r, v := vdpRunner(t, names...)
	set := func(reg int, val byte) {
		in := canary()
		in.C, in.B = byte(reg), val
		r.Call("VDP_SetReg", in)
	}
	type step struct {
		fn      string
		a       byte
		reg     int
		before  byte
		want    byte
		useA    bool
		comment string
	}
	steps := []step{
		{"VDP_DisplayOn", 0, 1, 0x80, 0xC0, false, "BL liga, K16 fica"},
		{"VDP_DisplayOff", 0, 1, 0xE1, 0xA1, false, "BL desliga, o resto fica"},
		{"VDP_VBlankIntOn", 0, 1, 0xC0, 0xE0, false, "IE0 liga"},
		{"VDP_VBlankIntOff", 0, 1, 0xF8, 0xD8, false, "IE0 desliga"},
		{"VDP_SpritesOff", 0, 8, 0x08, 0x0A, false, "SPD liga (sprites off)"},
		{"VDP_SpritesOn", 0, 8, 0x0B, 0x09, false, "SPD desliga (sprites on)"},
		{"VDP_SetLines", 212, 9, 0x0A, 0x8A, true, "212 linhas"},
		{"VDP_SetLines", 192, 9, 0x8A, 0x0A, true, "192 linhas"},
		{"VDP_SetLines", 100, 9, 0x8A, 0x0A, true, "valor invalido conta como 192"},
		{"VDP_SetRefresh", 50, 9, 0x88, 0x8A, true, "50 Hz (NT=1)"},
		{"VDP_SetRefresh", 60, 9, 0x8A, 0x88, true, "60 Hz (NT=0)"},
		{"VDP_Backdrop", 0x0D, 7, 0xF4, 0xFD, true, "so o nibble baixo"},
		{"VDP_Backdrop", 0x1F, 7, 0x40, 0x4F, true, "mascara em 4 bits"},
		{"VDP_TextColor", 0x03, 7, 0x1E, 0x3E, true, "so o nibble alto"},
		{"VDP_SetDisplayPage", 0, 2, 0x00, 0x1F, true, "pagina 0"},
		{"VDP_SetDisplayPage", 1, 2, 0x00, 0x3F, true, "pagina 1"},
		{"VDP_SetDisplayPage", 2, 2, 0x00, 0x5F, true, "pagina 2"},
		{"VDP_SetDisplayPage", 3, 2, 0x00, 0x7F, true, "pagina 3"},
	}
	for _, s := range steps {
		set(s.reg, s.before)
		in := canary()
		if s.useA {
			in.A = s.a
		}
		out := r.Call(s.fn, in)
		if v.Reg[s.reg] != s.want {
			t.Fatalf("%s(%d) [%s]: R#%d = %02X, quer %02X", s.fn, s.a, s.comment, s.reg, v.Reg[s.reg], s.want)
		}
		if shadowOf(t, r, s.reg) != s.want {
			t.Fatalf("%s(%d): copia sombra R#%d = %02X, quer %02X", s.fn, s.a, s.reg, shadowOf(t, r, s.reg), s.want)
		}
		keep(t, s.fn, in, out, "BCDEHL")
		if s.useA {
			keep(t, s.fn, in, out, "A")
		}
	}
}

func TestVDPVram17(t *testing.T) {
	r, v := vdpRunner(t, "VDP_VramSetWrite", "VDP_VramSetRead", "VDP_VramPut", "VDP_VramGet", "VDP_VramWriteStream",
		"VDP_VramReadStream", "VDP_VramFillStream", "VDP_VPoke", "VDP_VPeek")
	pattern := make([]byte, 70000)
	for i := range pattern {
		pattern[i] = byte(i*7 + i>>8)
	}
	copy(r.M.Mem[0x8000:], pattern[:0x7000]) // origem na RAM (até 28 KB é o que cabe aqui)

	starts := []uint32{0x00000, 0x00001, 0x03FF0, 0x0FFF0, 0x10000, 0x13FFE, 0x1FF00}
	for _, start := range starts {
		for _, n := range []int{0, 1, 2, 255, 256, 257, 1000, 5000} {
			if start+uint32(n) > 0x20000 {
				continue
			}
			in := canary().WithHL(uint16(start))
			in.A = byte(start >> 16)
			out := r.Call("VDP_VramSetWrite", in)
			keep(t, "VramSetWrite", in, out, "ABCDEHL")
			in = canary().WithHL(0x8000).WithBC(uint16(n))
			r.Call("VDP_VramWriteStream", in)
			for i := 0; i < n; i++ {
				if v.VRAM[start+uint32(i)] != r.M.Mem[0x8000+i] {
					t.Fatalf("WriteStream(inicio %05X, n %d): VRAM[%05X] = %02X, quer %02X", start, n, start+uint32(i), v.VRAM[start+uint32(i)], r.M.Mem[0x8000+i])
				}
			}

			// lê de volta com o ponteiro de leitura
			in = canary().WithHL(uint16(start))
			in.A = byte(start >> 16)
			r.Call("VDP_VramSetRead", in)
			clear(r.M.Mem[0xC000:0xE000])
			r.Call("VDP_VramReadStream", canary().WithDE(0xC000).WithBC(uint16(n)))
			for i := 0; i < n; i++ {
				if r.M.Mem[0xC000+i] != v.VRAM[start+uint32(i)] {
					t.Fatalf("ReadStream(inicio %05X, n %d): byte %d = %02X, quer %02X", start, n, i, r.M.Mem[0xC000+i], v.VRAM[start+uint32(i)])
				}
			}

			// preenche
			in = canary().WithHL(uint16(start))
			in.A = byte(start >> 16)
			r.Call("VDP_VramSetWrite", in)
			fill := canary().WithBC(uint16(n))
			fill.A = 0xC3
			r.Call("VDP_VramFillStream", fill)
			for i := 0; i < n; i++ {
				if v.VRAM[start+uint32(i)] != 0xC3 {
					t.Fatalf("FillStream(inicio %05X, n %d): VRAM[%05X] = %02X", start, n, start+uint32(i), v.VRAM[start+uint32(i)])
				}
			}
		}
	}

	// VPoke / VPeek / Put / Get
	for _, addr := range []uint32{0, 1, 0x3FFF, 0x4000, 0xFFFF, 0x10000, 0x1FFFF} {
		in := canary().WithHL(uint16(addr))
		in.D, in.A = byte(addr>>16), 0x9D
		out := r.Call("VDP_VPoke", in)
		if v.VRAM[addr] != 0x9D {
			t.Fatalf("VPoke(%05X): VRAM = %02X", addr, v.VRAM[addr])
		}
		keep(t, "VPoke", in, out, "ABCDEHL")
		v.VRAM[addr] = 0x6E
		in = canary().WithHL(uint16(addr))
		in.D = byte(addr >> 16)
		out = r.Call("VDP_VPeek", in)
		if out.A != 0x6E {
			t.Fatalf("VPeek(%05X) = %02X, quer 6E", addr, out.A)
		}
		keep(t, "VPeek", in, out, "BCDEHL")
	}
	v.SetAddr(0x2000)
	in := canary()
	in.A = 0x11
	r.Call("VDP_VramPut", in)
	r.Call("VDP_VramPut", in)
	if v.VRAM[0x2000] != 0x11 || v.VRAM[0x2001] != 0x11 {
		t.Error("VramPut nao escreveu no ponteiro atual")
	}
}

func TestVDPClearVRAM(t *testing.T) {
	r, v := vdpRunner(t, "VDP_ClearVRAM")
	for _, blocks := range []int{0, 1, 2, 3, 4, 5, 7, 8} {
		for i := range v.VRAM {
			v.VRAM[i] = 0xEE
		}
		in := canary()
		in.A = byte(blocks)
		out := r.Call("VDP_ClearVRAM", in)
		for i := range v.VRAM {
			want := byte(0xEE)
			if i < blocks*0x4000 {
				want = 0
			}
			if v.VRAM[i] != want {
				t.Fatalf("ClearVRAM(%d blocos): VRAM[%05X] = %02X, quer %02X", blocks, i, v.VRAM[i], want)
			}
		}
		keep(t, "ClearVRAM", in, out, "ABCDEHL")
	}
}

var (
	palMSX2 = [16][3]byte{{0, 0, 0}, {0, 0, 0}, {1, 6, 1}, {3, 7, 3}, {1, 1, 7}, {2, 3, 7}, {5, 1, 1}, {2, 6, 7}, {7, 1, 1}, {7, 3, 3}, {6, 6, 1}, {6, 6, 4}, {1, 4, 1}, {6, 2, 5}, {5, 5, 5}, {7, 7, 7}}
	palMSX1 = [16][3]byte{{0, 0, 0}, {0, 0, 0}, {1, 5, 1}, {3, 6, 3}, {2, 2, 6}, {3, 3, 7}, {5, 2, 2}, {2, 6, 7}, {6, 2, 2}, {6, 3, 3}, {5, 5, 2}, {6, 6, 3}, {1, 4, 1}, {5, 2, 5}, {5, 5, 5}, {7, 7, 7}}
)

func TestVDPPalette(t *testing.T) {
	r, v := vdpRunner(t, "VDP_SetPaletteEntry", "VDP_SetPaletteBlock", "VDP_SetDefaultPalette", "VDP_SetMSX1Palette", "VDP_PalShadow")
	shadow := func(i int) [2]byte {
		a := int(r.Addr("VDP_PalShadow")) + 2*i
		return [2]byte{r.M.Mem[a], r.M.Mem[a+1]}
	}
	for idx := 0; idx < 16; idx++ {
		for _, c := range [][3]byte{{0, 0, 0}, {7, 7, 7}, {1, 2, 3}, {5, 0, 7}, {12, 9, 15}} { // 12,9,15 testam a mascara de 3 bits
			in := canary()
			in.A, in.B, in.C, in.D = byte(idx), c[0], c[1], c[2]
			out := r.Call("VDP_SetPaletteEntry", in)
			want := [3]byte{c[0] & 7, c[1] & 7, c[2] & 7}
			if v.Pal[idx] != want {
				t.Fatalf("SetPaletteEntry(%d, %v): VDP %v, quer %v", idx, c, v.Pal[idx], want)
			}
			if got, w := shadow(idx), [2]byte{want[0]<<4 | want[2], want[1]}; got != w {
				t.Fatalf("copia sombra da cor %d = %02X, quer %02X", idx, got, w)
			}
			keep(t, "SetPaletteEntry", in, out, "ABCDEHL")
		}
	}
	// bloco de 32 bytes
	blk := make([]byte, 32)
	for i := range blk {
		blk[i] = byte(i * 5 & 0x77)
	}
	copy(r.M.Mem[0x8000:], blk)
	in := canary().WithHL(0x8000)
	out := r.Call("VDP_SetPaletteBlock", in)
	for i := 0; i < 16; i++ {
		want := [3]byte{blk[2*i] >> 4 & 7, blk[2*i+1] & 7, blk[2*i] & 7}
		if v.Pal[i] != want {
			t.Fatalf("SetPaletteBlock: cor %d = %v, quer %v", i, v.Pal[i], want)
		}
		if shadow(i) != [2]byte{blk[2*i], blk[2*i+1]} {
			t.Fatalf("SetPaletteBlock: copia sombra da cor %d errada", i)
		}
	}
	keep(t, "SetPaletteBlock", in, out, "ABCDEHL")
	// paletas prontas
	for _, c := range []struct {
		fn   string
		want [16][3]byte
	}{{"VDP_SetDefaultPalette", palMSX2}, {"VDP_SetMSX1Palette", palMSX1}} {
		v.Pal = [16][3]byte{}
		in := canary()
		out := r.Call(c.fn, in)
		if v.Pal != c.want {
			t.Fatalf("%s: paleta %v, quer %v", c.fn, v.Pal, c.want)
		}
		keep(t, c.fn, in, out, "ABCDEHL")
	}
}

// modeExpect: M5..M1 e R#2..R#6, R#10, R#11 de cada modo (valores do MSX-BASIC).
var modeExpect = []struct {
	name string
	mode byte // (M5 M4 M3 M2 M1)
	regs [9]byte
}{
	{"SCREEN 0 (40 col)", 0b00001, [9]byte{2: 0x00, 3: 0x00, 4: 0x01, 5: 0x00, 6: 0x00}},
	{"SCREEN 1", 0b00000, [9]byte{2: 0x06, 3: 0x80, 4: 0x00, 5: 0x36, 6: 0x07}},
	{"SCREEN 2", 0b00100, [9]byte{2: 0x06, 3: 0xFF, 4: 0x03, 5: 0x36, 6: 0x07}},
	{"SCREEN 3", 0b00010, [9]byte{2: 0x02, 3: 0x00, 4: 0x00, 5: 0x36, 6: 0x07}},
	{"SCREEN 4", 0b01000, [9]byte{2: 0x06, 3: 0xFF, 4: 0x03, 5: 0x3F, 6: 0x07}},
	{"SCREEN 5", 0b01100, [9]byte{2: 0x1F, 5: 0xEF, 6: 0x0F, 8: 0x00}},
	{"SCREEN 6", 0b10000, [9]byte{2: 0x1F, 5: 0xEF, 6: 0x0F, 8: 0x00}},
	{"SCREEN 7", 0b10100, [9]byte{2: 0x1F, 5: 0xF7, 6: 0x1E, 8: 0x01}},
	{"SCREEN 8", 0b11100, [9]byte{2: 0x1F, 5: 0xF7, 6: 0x1E, 8: 0x01}},
	{"SCREEN 0 (80 col)", 0b01001, [9]byte{2: 0x03, 3: 0x27, 4: 0x02, 5: 0x00, 6: 0x00}},
}

func TestVDPSetMode(t *testing.T) {
	r, v := vdpRunner(t, "VDP_SetMode", "VDP_GetMode", "VDP_SetReg")
	set := func(reg int, val byte) {
		in := canary()
		in.C, in.B = byte(reg), val
		r.Call("VDP_SetReg", in)
	}
	for mode, e := range modeExpect {
		// R#0 e R#1 já têm bits que NÃO são de modo (IE1/IE2, K16/BL/IE0/MAG): têm que sobreviver
		set(0, 0x30|0x01)
		set(1, 0xE0|0x03)
		in := canary()
		in.A = byte(mode)
		out := r.Call("VDP_SetMode", in)
		if out.Carry() {
			t.Fatalf("SetMode(%d) recusou um modo valido", mode)
		}
		if got := v.Mode(); got != e.mode {
			t.Fatalf("%s: bits de modo M5..M1 = %05b, quer %05b", e.name, got, e.mode)
		}
		if v.Reg[0]&^0x0E != 0x31 {
			t.Fatalf("%s: R#0 perdeu bits fora do modo: %02X", e.name, v.Reg[0])
		}
		if v.Reg[1]&^0x18 != 0xE3 {
			t.Fatalf("%s: R#1 perdeu bits fora do modo (tela ligada, IE0, K16, MAG): %02X", e.name, v.Reg[1])
		}
		for reg := 2; reg <= 6; reg++ {
			if v.Reg[reg] != e.regs[reg] {
				t.Fatalf("%s: R#%d = %02X, quer %02X", e.name, reg, v.Reg[reg], e.regs[reg])
			}
			if shadowOf(t, r, reg) != e.regs[reg] {
				t.Fatalf("%s: copia sombra de R#%d = %02X", e.name, reg, shadowOf(t, r, reg))
			}
		}
		if v.Reg[11] != e.regs[8] || v.Reg[10] != e.regs[7] {
			t.Fatalf("%s: R#10=%02X R#11=%02X, quer %02X %02X", e.name, v.Reg[10], v.Reg[11], e.regs[7], e.regs[8])
		}
		keep(t, "SetMode", in, out, "ABCDEHL")

		got := r.Call("VDP_GetMode", canary())
		if got.A != byte(mode) {
			t.Fatalf("GetMode = %d, quer %d", got.A, mode)
		}
	}

	// modo inexistente: carry = 1 e nenhuma escrita
	n := len(v.RegWrites)
	in := canary()
	in.A = 10
	out := r.Call("VDP_SetMode", in)
	if !out.Carry() || len(v.RegWrites) != n {
		t.Errorf("SetMode(10): carry=%v, %d escritas novas", out.Carry(), len(v.RegWrites)-n)
	}
	in.A = 0xFF
	if out := r.Call("VDP_SetMode", in); !out.Carry() {
		t.Error("SetMode(255) deveria recusar")
	}
}
