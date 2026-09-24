package kaji80

import (
	"bytes"
	"strings"
	"testing"
)

// Codificações da documentação oficial do Z80 (Zilog UM0080 / "The Undocumented
// Z80 Documented", só as formas documentadas).
func TestExtraInstructionEncodings(t *testing.T) {
	cases := []struct {
		src  string
		want []byte
	}{
		// operações de bloco
		{"ldi", []byte{0xED, 0xA0}}, {"ldir", []byte{0xED, 0xB0}}, {"ldd", []byte{0xED, 0xA8}}, {"lddr", []byte{0xED, 0xB8}},
		{"cpi", []byte{0xED, 0xA1}}, {"cpir", []byte{0xED, 0xB1}}, {"cpd", []byte{0xED, 0xA9}}, {"cpdr", []byte{0xED, 0xB9}},
		{"ini", []byte{0xED, 0xA2}}, {"inir", []byte{0xED, 0xB2}}, {"ind", []byte{0xED, 0xAA}}, {"indr", []byte{0xED, 0xBA}},
		{"outi", []byte{0xED, 0xA3}}, {"otir", []byte{0xED, 0xB3}}, {"outd", []byte{0xED, 0xAB}}, {"otdr", []byte{0xED, 0xBB}},
		// outras sem operando
		{"rld", []byte{0xED, 0x6F}}, {"rrd", []byte{0xED, 0x67}}, {"reti", []byte{0xED, 0x4D}}, {"retn", []byte{0xED, 0x45}},
		{"daa", []byte{0x27}},
		// modo de interrupção, RST, registradores I e R
		{"im 0", []byte{0xED, 0x46}}, {"im 1", []byte{0xED, 0x56}}, {"im 2", []byte{0xED, 0x5E}},
		{"rst 00h", []byte{0xC7}}, {"rst 08h", []byte{0xCF}}, {"rst 10h", []byte{0xD7}}, {"rst 18h", []byte{0xDF}},
		{"rst 20h", []byte{0xE7}}, {"rst 28h", []byte{0xEF}}, {"rst 30h", []byte{0xF7}}, {"rst 38h", []byte{0xFF}},
		{"ld i, a", []byte{0xED, 0x47}}, {"ld r, a", []byte{0xED, 0x4F}},
		// LD rr,(nn) e (nn),rr com BC/DE/SP e IX/IY
		{"ld bc, (1234h)", []byte{0xED, 0x4B, 0x34, 0x12}}, {"ld de, (1234h)", []byte{0xED, 0x5B, 0x34, 0x12}},
		{"ld sp, (1234h)", []byte{0xED, 0x7B, 0x34, 0x12}},
		{"ld (1234h), bc", []byte{0xED, 0x43, 0x34, 0x12}}, {"ld (1234h), de", []byte{0xED, 0x53, 0x34, 0x12}},
		{"ld (1234h), sp", []byte{0xED, 0x73, 0x34, 0x12}},
		{"ld ix, (1234h)", []byte{0xDD, 0x2A, 0x34, 0x12}}, {"ld (1234h), ix", []byte{0xDD, 0x22, 0x34, 0x12}},
		{"ld iy, (1234h)", []byte{0xFD, 0x2A, 0x34, 0x12}}, {"ld (1234h), iy", []byte{0xFD, 0x22, 0x34, 0x12}},
		// saltos indiretos e ADD IX,IX
		{"jp (hl)", []byte{0xE9}}, {"jp (ix)", []byte{0xDD, 0xE9}}, {"jp (iy)", []byte{0xFD, 0xE9}},
		{"add ix, ix", []byte{0xDD, 0x29}}, {"add iy, iy", []byte{0xFD, 0x29}},
		// IN r,(C) / OUT (C),r
		{"in b, (c)", []byte{0xED, 0x40}}, {"in c, (c)", []byte{0xED, 0x48}}, {"in d, (c)", []byte{0xED, 0x50}},
		{"in e, (c)", []byte{0xED, 0x58}}, {"in h, (c)", []byte{0xED, 0x60}}, {"in l, (c)", []byte{0xED, 0x68}},
		{"in a, (c)", []byte{0xED, 0x78}},
		{"out (c), b", []byte{0xED, 0x41}}, {"out (c), c", []byte{0xED, 0x49}}, {"out (c), d", []byte{0xED, 0x51}},
		{"out (c), e", []byte{0xED, 0x59}}, {"out (c), h", []byte{0xED, 0x61}}, {"out (c), l", []byte{0xED, 0x69}},
		{"out (c), a", []byte{0xED, 0x79}},
		// formas indexadas
		{"inc (ix+1)", []byte{0xDD, 0x34, 0x01}}, {"dec (iy+1)", []byte{0xFD, 0x35, 0x01}},
		{"inc (ix-1)", []byte{0xDD, 0x34, 0xFF}}, {"dec (ix+0)", []byte{0xDD, 0x35, 0x00}},
		{"rlc (ix+2)", []byte{0xDD, 0xCB, 0x02, 0x06}}, {"rrc (ix+2)", []byte{0xDD, 0xCB, 0x02, 0x0E}},
		{"rl (ix+2)", []byte{0xDD, 0xCB, 0x02, 0x16}}, {"rr (iy+2)", []byte{0xFD, 0xCB, 0x02, 0x1E}},
		{"sla (iy+2)", []byte{0xFD, 0xCB, 0x02, 0x26}}, {"sra (ix+2)", []byte{0xDD, 0xCB, 0x02, 0x2E}},
		{"srl (ix+2)", []byte{0xDD, 0xCB, 0x02, 0x3E}},
		{"bit 3, (ix+2)", []byte{0xDD, 0xCB, 0x02, 0x5E}}, {"res 3, (ix+2)", []byte{0xDD, 0xCB, 0x02, 0x9E}},
		{"set 3, (iy+2)", []byte{0xFD, 0xCB, 0x02, 0xDE}}, {"bit 0, (ix-2)", []byte{0xDD, 0xCB, 0xFE, 0x46}},
		{"set 7, (ix+0)", []byte{0xDD, 0xCB, 0x00, 0xFE}},
		// registrador alternativo
		{"ex af, af'", []byte{0x08}},
	}
	for _, c := range cases {
		t.Run(c.src, func(t *testing.T) {
			src := "MODULE X\nBANK 0\nPUBLIC S\nS:\n    " + c.src + "\n"
			obj, err := NewAssembler().Assemble(src)
			if err != nil {
				t.Fatalf("Assemble: %v", err)
			}
			if got := obj.Segments[0].Data; !bytes.Equal(got, c.want) {
				t.Errorf("bytes = % X, quer % X", got, c.want)
			}
		})
	}
}

// LD DE,(rótulo) e LD (rótulo),BC geram relocation (o endereço só é conhecido na ligação).
func TestExtraLdWithSymbolRelocates(t *testing.T) {
	src := "MODULE X\nBANK 0\nPUBLIC S\nEXTERN Far\nS:\n    ld de, (Far)\n    ld (Far), bc\n    ld ix, (Far)\n"
	obj, err := NewAssembler().Assemble(src)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(obj.Segments[0].Data), 4+4+4; got != want {
		t.Fatalf("tamanho = %d, quer %d", got, want)
	}
	if len(obj.Relocations) != 3 {
		t.Fatalf("relocations = %d, quer 3", len(obj.Relocations))
	}
	// posições dos operandos de endereço: depois de 2 bytes de prefixo em cada uma
	for i, off := range []uint16{2, 6, 10} {
		if obj.Relocations[i].Offset != off {
			t.Errorf("relocation %d em %d, quer %d", i, obj.Relocations[i].Offset, off)
		}
	}
}

func TestExtraInstructionErrors(t *testing.T) {
	cases := []struct{ src, want string }{
		{"im 3", "inválido"},
		{"rst 07h", "inválido"},
		{"rst 40h", "inválido"},
		{"bit 8, (ix+1)", "inválido"},
	}
	for _, c := range cases {
		_, err := NewAssembler().Assemble("MODULE X\nBANK 0\nPUBLIC S\nS:\n    " + c.src + "\n")
		if err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%q: erro = %v, quer conter %q", c.src, err, c.want)
		}
	}
}
