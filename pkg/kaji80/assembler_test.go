package kaji80

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/wilsonpilon/kizuna/pkg/mob"
)

func TestAssembleBasicInstructions(t *testing.T) {
	src := `
MODULE TestMod
BANK 0
PUBLIC Start
EXTERN SubRoutine

Start:
    nop
    halt
    di
    ei
    ld   a, 42
    ld   b, a
    push af
    pop  bc
    call SubRoutine
    jp   EndLabel
EndLabel:
    ret
`
	asm := NewAssembler()
	obj, err := asm.Assemble(src)
	if err != nil {
		t.Fatalf("Assemble failed: %v", err)
	}

	if len(obj.Segments) != 1 {
		t.Fatalf("Expected 1 segment, got %d", len(obj.Segments))
	}
	seg := obj.Segments[0]
	if seg.Bank != 0 {
		t.Errorf("Expected Bank 0, got %d", seg.Bank)
	}

	// Verificar se símbolos Start e SubRoutine existem
	var foundStart, foundSubRoutine bool
	for _, sym := range obj.Symbols {
		if sym.Name == "Start" && sym.Class == mob.SymbolPublic {
			foundStart = true
		}
		if sym.Name == "SubRoutine" && sym.Class == mob.SymbolExtern {
			foundSubRoutine = true
		}
	}
	if !foundStart {
		t.Errorf("Symbol 'Start' not found as PUBLIC")
	}
	if !foundSubRoutine {
		t.Errorf("Symbol 'SubRoutine' not found as EXTERN")
	}

	// Deve ter uma relocation para o CALL SubRoutine e outra para JP EndLabel
	if len(obj.Relocations) < 1 {
		t.Errorf("Expected at least 1 relocation for external CALL, got %d", len(obj.Relocations))
	}
}

func TestAssembleDemoScreenAsm(t *testing.T) {
	// Testar diretamente o arquivo demo/screen.asm do projeto
	demoPath := filepath.Join("..", "..", "demo", "screen.asm")
	content, err := os.ReadFile(demoPath)
	if err != nil {
		t.Fatalf("Could not read demo/screen.asm: %v", err)
	}

	asm := NewAssembler()
	obj, err := asm.Assemble(string(content))
	if err != nil {
		t.Fatalf("Assemble demo/screen.asm failed: %v", err)
	}

	if len(obj.Segments) != 1 {
		t.Fatalf("Expected 1 segment, got %d", len(obj.Segments))
	}
	seg := obj.Segments[0]
	if seg.Bank != 1 {
		t.Errorf("Expected BANK 1 as specified in screen.asm, got %d", seg.Bank)
	}

	// Verificar os símbolos exportados e importados
	symbolNames := make(map[string]mob.SymbolClass)
	for _, sym := range obj.Symbols {
		symbolNames[sym.Name] = sym.Class
	}

	if symbolNames["Setup"] != mob.SymbolPublic {
		t.Errorf("Expected 'Setup' to be PUBLIC")
	}
	if symbolNames["BIOS_CHGMOD"] != mob.SymbolExtern {
		t.Errorf("Expected 'BIOS_CHGMOD' to be EXTERN")
	}
	if symbolNames["BIOS_CHGCLR"] != mob.SymbolExtern {
		t.Errorf("Expected 'BIOS_CHGCLR' to be EXTERN")
	}
	if symbolNames["BIOS_WIDTH"] != mob.SymbolExtern {
		t.Errorf("Expected 'BIOS_WIDTH' to be EXTERN")
	}
	if symbolNames["BIOS_KEYOFF"] != mob.SymbolExtern {
		t.Errorf("Expected 'BIOS_KEYOFF' to be EXTERN")
	}

	// Validar que o .MOB gerado pode ser serializado e deserializado
	encoded, err := mob.Encode(obj)
	if err != nil {
		t.Fatalf("mob.Encode failed: %v", err)
	}

	decoded, err := mob.Decode(encoded)
	if err != nil {
		t.Fatalf("mob.Decode failed: %v", err)
	}

	if decoded.Segments[0].Bank != 1 {
		t.Errorf("Expected decoded bank to be 1, got %d", decoded.Segments[0].Bank)
	}
	if !bytes.Equal(decoded.Segments[0].Data, seg.Data) {
		t.Errorf("Decoded segment data differs from original")
	}
}

func TestDataDirectives(t *testing.T) {
	src := `
MODULE DataMod
BANK 0
PUBLIC MyString, MyWord

MyString:
    db 13, "Hello, World!", 0
MyWord:
    dw 0x1234, 0xABCD
`
	asm := NewAssembler()
	obj, err := asm.Assemble(src)
	if err != nil {
		t.Fatalf("Assemble failed: %v", err)
	}

	seg := obj.Segments[0]
	// 1 byte (13) + 13 bytes ("Hello, World!") + 1 byte (0) + 4 bytes (2 words) = 19 bytes
	expectedLen := 1 + 13 + 1 + 4
	if len(seg.Data) != expectedLen {
		t.Errorf("Expected data size %d, got %d", expectedLen, len(seg.Data))
	}

	// Verificar little-endian de 0x1234: 0x34, 0x12
	word0Lo := seg.Data[15]
	word0Hi := seg.Data[16]
	if word0Lo != 0x34 || word0Hi != 0x12 {
		t.Errorf("Expected 0x34 0x12 for word 0x1234, got 0x%02X 0x%02X", word0Lo, word0Hi)
	}
}

func TestRotateShiftBitInstructions(t *testing.T) {
	src := `
MODULE BitOps
BANK 0
PUBLIC TestBitOps

TestBitOps:
    rla
    rra
    cpl
    scf
    ccf
    neg
    sla c
    srl a
    bit 3, a
    res 2, b
    set 7, (hl)
`
	asm := NewAssembler()
	obj, err := asm.Assemble(src)
	if err != nil {
		t.Fatalf("Assemble failed: %v", err)
	}

	expected := []byte{
		0x17,       // rla
		0x1F,       // rra
		0x2F,       // cpl
		0x37,       // scf
		0x3F,       // ccf
		0xED, 0x44, // neg
		0xCB, 0x21, // sla c
		0xCB, 0x3F, // srl a
		0xCB, 0x5F, // bit 3, a
		0xCB, 0x90, // res 2, b
		0xCB, 0xFE, // set 7, (hl)
	}

	seg := obj.Segments[0]
	if !bytes.Equal(seg.Data, expected) {
		t.Errorf("Byte mismatch.\nGot:      % X\nExpected: % X", seg.Data, expected)
	}
}

func TestPass1AndPass2Sync(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "lib", "src", "string.asm"))
	if err != nil {
		t.Fatalf("Failed to read string.asm: %v", err)
	}

	asm := NewAssembler()
	obj, err := asm.Assemble(string(data))
	if err != nil {
		t.Fatalf("Assemble failed: %v", err)
	}

	t.Logf("Total segment size: %d", len(obj.Segments[0].Data))
	for _, sym := range obj.Symbols {
		if sym.Class == mob.SymbolPublic {
			t.Logf("Symbol %-15s -> Offset: 0x%04X (%d) -> First byte: 0x%02X", sym.Name, sym.Offset, sym.Offset, obj.Segments[0].Data[sym.Offset])
		}
	}
	for _, sym := range obj.Symbols {
		if sym.Name == "PrintDec16" {
			if obj.Segments[0].Data[sym.Offset] != 0xF5 {
				t.Fatalf("Expected PrintDec16 to start with 0xF5 (PUSH AF), got 0x%02X", obj.Segments[0].Data[sym.Offset])
			}
		}
	}
}

func Test16BitAluInstructions(t *testing.T) {
	src := `
MODULE Test16
BANK 0
PUBLIC Start
Start:
    add  hl, bc
    add  hl, de
    adc  hl, bc
    adc  hl, de
    sbc  hl, bc
    sbc  hl, de
    ld   hl, (1234h)
    ld   (1234h), hl
`
	asm := NewAssembler()
	obj, err := asm.Assemble(src)
	if err != nil {
		t.Fatalf("Assemble failed: %v", err)
	}

	data := obj.Segments[0].Data
	expected := []byte{
		0x09,             // ADD HL, BC
		0x19,             // ADD HL, DE
		0xED, 0x4A,       // ADC HL, BC
		0xED, 0x5A,       // ADC HL, DE
		0xED, 0x42,       // SBC HL, BC
		0xED, 0x52,       // SBC HL, DE
		0x2A, 0x34, 0x12, // LD HL, (1234h)
		0x22, 0x34, 0x12, // LD (1234h), HL
	}

	if !bytes.Equal(data, expected) {
		t.Fatalf("Byte mismatch:\nGot:      % X\nExpected: % X", data, expected)
	}
}

