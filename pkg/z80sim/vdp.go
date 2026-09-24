package z80sim

// VDP é um modelo funcional do V9938 (o suficiente para verificar as rotinas da
// MSXLIB pelo CONTEÚDO da VRAM, não só pelo traço de portas): 128 KB de VRAM com
// contador de endereço de 17 bits e auto-incremento, registradores R#0-R#63,
// registradores de status, paleta de 16 cores e a escrita indireta de
// registrador (R#17). Não modela temporização, sprites nem vídeo; o motor de
// comandos está em vdpcmd.go.
//
// Portas: 98h dados da VRAM, 99h endereço/registrador (escrita) e status
// (leitura), 9Ah paleta, 9Bh registrador indireto.
type VDP struct {
	VRAM   [1 << 17]byte
	Reg    [64]byte
	Status [10]byte
	// Pal[i] = {R, G, B}, cada um 0..7
	Pal [16][3]byte

	addr      uint32
	latch     byte
	latched   bool
	readAhead byte
	palHalf   bool
	palRB     byte

	// S0Reads conta as leituras do status S#0. VBlankEvery > 0 faz o flag F
	// (vblank, bit 7 de S#0) aparecer a cada VBlankEvery-ésima leitura, para
	// que laços que esperam o retraço vertical terminem no simulador.
	S0Reads     int
	VBlankEvery int

	// RegWrites registra, em ordem, cada escrita de registrador (número, valor).
	RegWrites [][2]byte
}

// NewVDP cria um VDP zerado.
func NewVDP() *VDP { return &VDP{VBlankEvery: 4} }

// AttachVDP liga um VDP novo às portas 98h-9Bh da máquina e o devolve.
func (m *Machine) AttachVDP() *VDP {
	m.VDP = NewVDP()
	return m.VDP
}

const vramMask = 1<<17 - 1

func (v *VDP) writeReg(n, val byte) {
	n &= 0x3F
	if n == 14 {
		val &= 0x07
	}
	v.Reg[n] = val
	v.RegWrites = append(v.RegWrites, [2]byte{n, val})
	if n == 46 {
		v.runCommand()
	}
}

// Mode devolve o código de modo (M5..M1) como o V9938 documenta: bits 4..0 =
// M5 M4 M3 M2 M1.
func (v *VDP) Mode() byte {
	r0, r1 := v.Reg[0], v.Reg[1]
	m3, m4, m5 := r0>>1&1, r0>>2&1, r0>>3&1
	m1, m2 := r1>>4&1, r1>>3&1
	return m5<<4 | m4<<3 | m3<<2 | m2<<1 | m1
}

// Write98 é a escrita na porta de dados da VRAM.
func (v *VDP) Write98(b byte) {
	v.VRAM[v.addr&vramMask] = b
	v.addr = (v.addr + 1) & vramMask
	v.Reg[14] = byte(v.addr >> 14 & 7) // o contador de 17 bits leva o vai-um para R#14
	v.latched = false
}

// Read98 é a leitura da porta de dados (com pré-leitura, como o V9938).
func (v *VDP) Read98() byte {
	b := v.readAhead
	v.readAhead = v.VRAM[v.addr&vramMask]
	v.addr = (v.addr + 1) & vramMask
	v.Reg[14] = byte(v.addr >> 14 & 7)
	v.latched = false
	return b
}

// Write99 é a escrita na porta de comando: o 1º byte fica retido; o 2º decide
// entre escrita de registrador (bit 7 = 1) e ajuste do endereço da VRAM.
func (v *VDP) Write99(b byte) {
	if !v.latched {
		v.latch = b
		v.latched = true
		return
	}
	v.latched = false
	if b&0x80 != 0 {
		v.writeReg(b&0x3F, v.latch)
		return
	}
	v.addr = uint32(v.Reg[14]&7)<<14 | uint32(b&0x3F)<<8 | uint32(v.latch)
	if b&0x40 == 0 { // modo de leitura: pré-lê o primeiro byte
		v.readAhead = v.VRAM[v.addr&vramMask]
		v.addr = (v.addr + 1) & vramMask
	}
}

// Read99 lê o registrador de status selecionado por R#15.
func (v *VDP) Read99() byte {
	v.latched = false
	sel := v.Reg[15] & 0x0F
	if sel > 9 {
		return 0
	}
	if sel == 0 {
		v.S0Reads++
		if v.VBlankEvery > 0 && v.S0Reads%v.VBlankEvery == 0 {
			v.Status[0] |= 0x80
		}
		b := v.Status[0]
		v.Status[0] &= 0x1F // ler S#0 limpa F, 5S e C
		return b
	}
	return v.Status[sel]
}

// Write9A é a escrita de paleta: dois bytes por cor (R e B, depois G); o
// ponteiro R#16 avança sozinho.
func (v *VDP) Write9A(b byte) {
	if !v.palHalf {
		v.palRB = b
		v.palHalf = true
		return
	}
	v.palHalf = false
	i := v.Reg[16] & 15
	v.Pal[i] = [3]byte{v.palRB >> 4 & 7, b & 7, v.palRB & 7}
	v.Reg[16] = (v.Reg[16] + 1) & 15
}

// Write9B é a escrita indireta de registrador: grava no registrador apontado
// por R#17 e, se o bit 7 de R#17 for 0, avança o ponteiro.
func (v *VDP) Write9B(b byte) {
	reg := v.Reg[17] & 0x3F
	v.writeReg(reg, b)
	if v.Reg[17]&0x80 == 0 {
		v.Reg[17] = v.Reg[17]&0xC0 | (reg+1)&0x3F
	}
}

// SetAddr ajusta o contador de endereço de escrita diretamente (para testes).
func (v *VDP) SetAddr(a uint32) { v.addr = a & vramMask }

// Addr devolve o contador de endereço atual.
func (v *VDP) Addr() uint32 { return v.addr }
