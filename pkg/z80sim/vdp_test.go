package z80sim

import "testing"

func out99(v *VDP, a, b byte) { v.Write99(a); v.Write99(b) }

func TestVDPWriteRegisterAndAddress(t *testing.T) {
	v := NewVDP()
	out99(v, 0x60, 0x87) // R#7 = 60h
	if v.Reg[7] != 0x60 {
		t.Fatalf("R#7 = %02X", v.Reg[7])
	}
	// R#14 = 4 (bit 2 = A16), depois endereço de escrita 0x0123 => 0x10123
	out99(v, 0x04, 0x8E)
	out99(v, 0x23, 0x41) // 0x0123 | 0x4000 (bit 6 = escrita)
	v.Write98(0xAA)
	v.Write98(0xBB)
	if v.VRAM[0x10123] != 0xAA || v.VRAM[0x10124] != 0xBB {
		t.Fatalf("VRAM[10123..] = %02X %02X", v.VRAM[0x10123], v.VRAM[0x10124])
	}
	if v.Addr() != 0x10125 {
		t.Errorf("Addr = %05X", v.Addr())
	}
}

func TestVDPAutoIncrementCarriesIntoR14(t *testing.T) {
	v := NewVDP()
	v.SetAddr(0x3FFF)
	v.Write98(1)
	v.Write98(2) // cruzou 16K: A14 sobe
	if v.VRAM[0x3FFF] != 1 || v.VRAM[0x4000] != 2 || v.Reg[14] != 1 {
		t.Errorf("VRAM=%02X %02X R#14=%d", v.VRAM[0x3FFF], v.VRAM[0x4000], v.Reg[14])
	}
	v.SetAddr(0x1FFFF)
	v.Write98(9)
	v.Write98(8) // volta a 0
	if v.VRAM[0x1FFFF] != 9 || v.VRAM[0] != 8 {
		t.Errorf("wrap de 128K falhou: %02X %02X", v.VRAM[0x1FFFF], v.VRAM[0])
	}
}

func TestVDPReadUsesReadAhead(t *testing.T) {
	v := NewVDP()
	copy(v.VRAM[0x100:], []byte{10, 11, 12, 13})
	out99(v, 0x00, 0x01) // R#... endereço 0x0100 leitura: low=00, high=01|00 (bit 6 = 0)
	got := []byte{v.Read98(), v.Read98(), v.Read98()}
	if got[0] != 10 || got[1] != 11 || got[2] != 12 {
		t.Errorf("leitura = %v, quer [10 11 12]", got)
	}
}

func TestVDPStatusAndVBlank(t *testing.T) {
	v := NewVDP()
	out99(v, 0x00, 0x8F) // R#15 = 0 (S#0)
	frames := 0
	for i := 0; i < 20; i++ {
		if v.Read99()&0x80 != 0 {
			frames++
		}
	}
	if frames != 5 { // a cada 4a leitura
		t.Errorf("F apareceu %d vezes em 20 leituras, quer 5", frames)
	}
	v.Status[2] = 0x81
	out99(v, 0x02, 0x8F) // R#15 = 2
	if v.Read99() != 0x81 {
		t.Error("S#2 nao foi lido")
	}
}

func TestVDPPalette(t *testing.T) {
	v := NewVDP()
	out99(v, 0x03, 0x90) // R#16 = 3
	v.Write9A(0x52)      // R=5, B=2
	v.Write9A(0x07)      // G=7
	v.Write9A(0x11)
	v.Write9A(0x02)
	if v.Pal[3] != [3]byte{5, 7, 2} || v.Pal[4] != [3]byte{1, 2, 1} || v.Reg[16] != 5 {
		t.Errorf("paleta = %v %v R#16=%d", v.Pal[3], v.Pal[4], v.Reg[16])
	}
}

func TestVDPIndirectRegisterWrite(t *testing.T) {
	v := NewVDP()
	out99(v, 0x20, 0x91) // R#17 = 20h => aponta R#32, auto-incremento
	v.Write9B(1)
	v.Write9B(2)
	v.Write9B(3)
	if v.Reg[32] != 1 || v.Reg[33] != 2 || v.Reg[34] != 3 {
		t.Errorf("R#32..34 = %d %d %d", v.Reg[32], v.Reg[33], v.Reg[34])
	}
	out99(v, 0xA0, 0x91) // bit 7: sem auto-incremento, ainda em R#32
	v.Write9B(7)
	v.Write9B(8)
	if v.Reg[32] != 8 || v.Reg[33] != 2 {
		t.Errorf("sem auto-incremento: R#32=%d R#33=%d", v.Reg[32], v.Reg[33])
	}
}

func TestVDPModeBits(t *testing.T) {
	v := NewVDP()
	// GRAPHIC 4 (SCREEN 5): M3+M4 => R#0 = 06h  => (M5 M4 M3 M2 M1) = 01100
	out99(v, 0x06, 0x80)
	if v.Mode() != 0x0C {
		t.Errorf("Mode = %05b, quer 01100", v.Mode())
	}
	out99(v, 0x08, 0x80) // GRAPHIC 5
	if v.Mode() != 0x10 {
		t.Errorf("Mode = %05b, quer 10000", v.Mode())
	}
	out99(v, 0x00, 0x80)
	out99(v, 0x10, 0x81) // TEXT 1 (M1 em R#1)
	if v.Mode() != 0x01 {
		t.Errorf("Mode = %05b, quer 00001", v.Mode())
	}
}

func TestMachineRoutesVDPPorts(t *testing.T) {
	m := New()
	v := m.AttachVDP()
	// OUT (99h),20h ; OUT (99h),87h  => R#7 = 20h
	m.Load(0x0100, []byte{0x3E, 0x20, 0xD3, 0x99, 0x3E, 0x87, 0xD3, 0x99, 0x76})
	m.CPU.SetPC(0x0100)
	if err := m.Run(100); err != nil {
		t.Fatal(err)
	}
	if v.Reg[7] != 0x20 {
		t.Errorf("R#7 = %02X", v.Reg[7])
	}
	if len(m.Ports) != 2 {
		t.Errorf("o traço de portas deve continuar sendo gravado: %v", m.Ports)
	}
}

func TestJiffyTick(t *testing.T) {
	m := New()
	m.JiffyEvery = 10
	m.Load(0x0100, []byte{0x18, 0xFE}) // JR $
	m.CPU.SetPC(0x0100)
	_ = m.Run(105)
	if j := uint16(m.Mem[0xFC9E]) | uint16(m.Mem[0xFC9F])<<8; j != 11 {
		t.Errorf("JIFFY = %d, quer 11", j)
	}
}
