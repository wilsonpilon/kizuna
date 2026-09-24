package z80sim

// Motor de comandos do V9938 (R#32-R#46), modelo funcional: cada comando roda
// inteiro na hora em que R#46 é escrito (CE volta a 0 já na leitura seguinte),
// sem modelar temporização. Vale nos modos bitmap (SCREEN 5 a 8); em qualquer
// outro modo o comando é aceito e ignorado, como no chip. Memória expandida
// (MXS/MXD) não é modelada -- só a VRAM.
//
// As transferências com a CPU (LMMC, HMMC, LMCM) seguem o protocolo real: o
// primeiro dado vai em R#44 junto com o comando; depois S#2 bit 7 (TR) diz que
// o chip quer o próximo byte (escrita em R#44) ou tem um pixel pronto em S#7
// (leitura); CE fica em 1 até o último byte.
//
// Os registradores de coordenada (R#32-R#43) NÃO são atualizados ao fim do
// comando (o chip deixa DY/SY/NY avançados); rotinas que dependam disso não
// são verificáveis aqui.

const (
	cmdSTOP  = 0x0
	cmdPOINT = 0x4
	cmdPSET  = 0x5
	cmdSRCH  = 0x6
	cmdLINE  = 0x7
	cmdLMMV  = 0x8
	cmdLMMM  = 0x9
	cmdLMCM  = 0xA
	cmdLMMC  = 0xB
	cmdHMMV  = 0xC
	cmdHMMM  = 0xD
	cmdYMMM  = 0xE
	cmdHMMC  = 0xF
)

// cmdXfer é uma transferência com a CPU em andamento (LMMC, HMMC, LMCM).
type cmdXfer struct {
	kind   byte
	op     byte
	coords [][2]int // pixels (LMMC/LMCM) ou (byte de linha, linha) em HMMC
	idx    int
}

func (v *VDP) geometry() (width, bpp, bytesPerLine int, ok bool) {
	switch v.Mode() {
	case 0x0C: // GRAPHIC 4
		return 256, 4, 128, true
	case 0x10: // GRAPHIC 5
		return 512, 2, 128, true
	case 0x14: // GRAPHIC 6
		return 512, 4, 256, true
	case 0x1C: // GRAPHIC 7
		return 256, 8, 256, true
	}
	return 0, 0, 0, false
}

func (v *VDP) reg16(lo int, mask int) int { return (int(v.Reg[lo]) | int(v.Reg[lo+1])<<8) & mask }

// pixelAddr devolve o endereço do byte e o deslocamento de bits de um pixel.
func (v *VDP) pixelAddr(x, y int) (addr int, shift uint, ok bool) {
	width, bpp, bpl, mode := v.geometry()
	if !mode || x < 0 || x >= width || y < 0 {
		return 0, 0, false
	}
	y &= 1023
	ppb := 8 / bpp
	addr = (y*bpl + x/ppb) & vramMask
	shift = uint(8 - bpp - (x%ppb)*bpp)
	return addr, shift, true
}

func (v *VDP) pget(x, y int) byte {
	_, bpp, _, _ := v.geometry()
	addr, sh, ok := v.pixelAddr(x, y)
	if !ok {
		return 0
	}
	return v.VRAM[addr] >> sh & byte(1<<uint(bpp)-1)
}

func (v *VDP) pset(x, y int, c, op byte) {
	_, bpp, _, _ := v.geometry()
	addr, sh, ok := v.pixelAddr(x, y)
	if !ok {
		return
	}
	mask := byte(1<<uint(bpp) - 1)
	c &= mask
	if op&8 != 0 && c == 0 { // variantes T: cor 0 é transparente
		return
	}
	d := v.VRAM[addr] >> sh & mask
	switch op & 7 {
	case 0:
		d = c
	case 1:
		d &= c
	case 2:
		d |= c
	case 3:
		d ^= c
	case 4:
		d = ^c & mask
	default: // 5..7 não são operações definidas: ficam como IMP
		d = c
	}
	v.VRAM[addr] = v.VRAM[addr]&^(mask<<sh) | d<<sh
}

func sgn(neg bool) int {
	if neg {
		return -1
	}
	return 1
}

// nDim lê NX/NY; 0 significa "nada a fazer" neste modelo.
func (v *VDP) dims() (nx, ny int) {
	return v.reg16(40, 0x1FF), v.reg16(42, 0x3FF)
}

func (v *VDP) runCommand() {
	v.xfer = nil
	v.Status[2] &^= 0x81 // CE = 0, TR = 0
	cmd := v.Reg[46] >> 4
	op := v.Reg[46] & 0x0F
	if cmd == cmdSTOP {
		return
	}
	if _, _, _, ok := v.geometry(); !ok {
		return
	}
	v.Commands++
	arg := v.Reg[45]
	dix, diy := sgn(arg&0x04 != 0), sgn(arg&0x08 != 0)
	sx, sy := v.reg16(32, 0x1FF), v.reg16(34, 0x3FF)
	dx, dy := v.reg16(36, 0x1FF), v.reg16(38, 0x3FF)
	nx, ny := v.dims()
	clr := v.Reg[44]
	_, bpp, bpl, _ := v.geometry()
	ppb := 8 / bpp

	switch cmd {
	case cmdPSET:
		v.pset(dx, dy, clr, op)
	case cmdPOINT:
		v.Status[7] = v.pget(sx, sy)
	case cmdLMMV:
		for j := 0; j < ny; j++ {
			for i := 0; i < nx; i++ {
				v.pset(dx+dix*i, dy+diy*j, clr, op)
			}
		}
	case cmdLMMM:
		for j := 0; j < ny; j++ {
			for i := 0; i < nx; i++ {
				v.pset(dx+dix*i, dy+diy*j, v.pget(sx+dix*i, sy+diy*j), op)
			}
		}
	case cmdHMMV, cmdHMMM:
		nb := nx / ppb
		bx0, sbx0 := dx/ppb, sx/ppb
		for j := 0; j < ny; j++ {
			for i := 0; i < nb; i++ {
				bx := bx0 + dix*i
				if bx < 0 || bx >= v.lineBytes() {
					continue
				}
				da := (((dy+diy*j)&1023)*bpl + bx) & vramMask
				if cmd == cmdHMMV {
					v.VRAM[da] = clr
					continue
				}
				sbx := sbx0 + dix*i
				if sbx < 0 || sbx >= v.lineBytes() {
					continue
				}
				v.VRAM[da] = v.VRAM[(((sy+diy*j)&1023)*bpl+sbx)&vramMask]
			}
		}
	case cmdYMMM:
		bx0 := dx / ppb
		lb := v.lineBytes()
		for j := 0; j < ny; j++ {
			srow, drow := (sy+diy*j)&1023, (dy+diy*j)&1023
			for bx := bx0; bx >= 0 && bx < lb; bx += dix {
				v.VRAM[(drow*bpl+bx)&vramMask] = v.VRAM[(srow*bpl+bx)&vramMask]
			}
		}
	case cmdLINE:
		long, short := nx, ny
		asx := (long - 1) >> 1
		x, y := dx, dy
		for i := 0; i <= long; i++ {
			v.pset(x, y, clr, op)
			if i == long {
				break
			}
			if arg&0x01 == 0 { // lado maior em X
				x += dix
			} else {
				y += diy
			}
			asx -= short
			if asx < 0 {
				asx += long
				if arg&0x01 == 0 {
					y += diy
				} else {
					x += dix
				}
			}
		}
	case cmdSRCH:
		eq := arg&0x02 != 0
		width, _, _, _ := v.geometry()
		found := false
		x := sx
		for ; x >= 0 && x < width; x += dix {
			if (v.pget(x, sy) == clr&byte(1<<uint(bpp)-1)) != eq {
				found = true
				break
			}
		}
		if !found {
			if x < 0 {
				x = 0
			} else if x >= width {
				x = width - 1
			}
			v.Status[2] &^= 0x10
		} else {
			v.Status[2] |= 0x10
		}
		v.Status[8], v.Status[9] = byte(x), byte(x>>8)&1
	case cmdLMMC, cmdLMCM:
		xf := &cmdXfer{kind: cmd, op: op}
		for j := 0; j < ny; j++ {
			for i := 0; i < nx; i++ {
				xf.coords = append(xf.coords, [2]int{dx + dix*i, dy + diy*j})
				if cmd == cmdLMCM {
					xf.coords[len(xf.coords)-1] = [2]int{sx + dix*i, sy + diy*j}
				}
			}
		}
		v.startXfer(xf, clr)
	case cmdHMMC:
		xf := &cmdXfer{kind: cmd, op: op}
		nb := nx / ppb
		for j := 0; j < ny; j++ {
			for i := 0; i < nb; i++ {
				xf.coords = append(xf.coords, [2]int{dx/ppb + dix*i, (dy + diy*j) & 1023})
			}
		}
		v.startXfer(xf, clr)
	}
}

// lineBytes: bytes por linha de imagem (largura em pixels / pixels por byte).
func (v *VDP) lineBytes() int {
	width, bpp, _, _ := v.geometry()
	return width / (8 / bpp)
}

func (v *VDP) startXfer(xf *cmdXfer, first byte) {
	if len(xf.coords) == 0 {
		return
	}
	v.xfer = xf
	v.Status[2] |= 0x81 // CE = 1, TR = 1
	if xf.kind == cmdLMCM {
		v.Status[7] = v.pget(xf.coords[0][0], xf.coords[0][1])
		return
	}
	v.xferWrite(first)
}

// xferWrite entrega um byte de dados (escrita em R#44) a LMMC/HMMC.
func (v *VDP) xferWrite(b byte) {
	xf := v.xfer
	if xf == nil || xf.kind == cmdLMCM || xf.idx >= len(xf.coords) {
		return
	}
	c := xf.coords[xf.idx]
	if xf.kind == cmdLMMC {
		v.pset(c[0], c[1], b, xf.op)
	} else if c[0] >= 0 && c[0] < v.lineBytes() {
		_, _, bpl, _ := v.geometry()
		v.VRAM[(c[1]*bpl+c[0])&vramMask] = b
	}
	xf.idx++
	if xf.idx >= len(xf.coords) {
		v.xfer = nil
		v.Status[2] &^= 0x81
	}
}

// xferRead devolve o pixel de S#7 durante um LMCM e avança para o seguinte.
func (v *VDP) xferRead() byte {
	xf := v.xfer
	b := v.Status[7]
	if xf == nil || xf.kind != cmdLMCM {
		return b
	}
	xf.idx++
	if xf.idx >= len(xf.coords) {
		v.xfer = nil
		v.Status[2] &^= 0x81
	} else {
		c := xf.coords[xf.idx]
		v.Status[7] = v.pget(c[0], c[1])
	}
	return b
}
