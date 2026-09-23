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

// TestCharLiteralImmediate cobre um bug real encontrado em 2026-09-22:
// parseLine reconstruía operandos concatenando tokens.Value direto, e o
// lexer já devolve o conteúdo de um TokenString SEM as aspas (correto para
// DB, que usa os tokens crus). Isso deixava "'$'" indistinguível de um
// identificador solto para parseImm8, que silenciosamente devolvia 0 --
// "LD (HL), '$'" virava "LD (HL), 0" em vez de "LD (HL), 24h", sem erro de
// montagem. Descoberto porque sample/fileio (que usava esse terminador
// pra imprimir uma string lida de um arquivo via BDOS função 09h) imprimia
// lixo de memória indefinidamente em vez de parar no terminador.
func TestCharLiteralImmediate(t *testing.T) {
	src := `
MODULE CharLit
BANK 0
PUBLIC Start
Start:
    LD HL, 8000h
    LD (HL), '$'
    LD A, 'A'
    RET
ENDMOD
`
	asm := NewAssembler()
	obj, err := asm.Assemble(src)
	if err != nil {
		t.Fatalf("Assemble failed: %v", err)
	}

	data := obj.Segments[0].Data
	// LD HL,8000h (3) / LD (HL),n (2) / LD A,n (2) / RET (1)
	if len(data) != 8 {
		t.Fatalf("Expected 8 bytes, got %d: % X", len(data), data)
	}
	if data[3] != 0x36 || data[4] != '$' {
		t.Errorf("Expected LD (HL),'$' -> 36 24, got %02X %02X", data[3], data[4])
	}
	if data[5] != 0x3E || data[6] != 'A' {
		t.Errorf("Expected LD A,'A' -> 3E 41, got %02X %02X", data[5], data[6])
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

// TestLabelAfterLdAIndirectAndDb cobre os dois bugs de dessincronia Pass1/Pass2
// encontrados em 2026-09-10: "LD A, (rotulo)" era subestimado em 1 byte no
// Pass 1 (tratado como "LD A, n" imediato de 2 bytes em vez de "LD A, (nn)"
// de 3 bytes), e "rotulo: DB valor" tinha o proprio token do mnemonico "DB"
// contado como se fosse um byte de dado extra, tanto no Pass 1 quanto na
// emissao real do Pass 2. Qualquer rotulo declarado depois desses padroes
// ficava com o endereco errado na tabela de simbolos -- silenciosamente,
// sem erro de montagem -- corrompendo referencias cruzadas resolvidas pelo
// linker. O Assemble() agora verifica essa consistencia internamente; este
// teste fixa o comportamento correto dos rotulos em si.
func TestLabelAfterLdAIndirectAndDb(t *testing.T) {
	src := `
MODULE LabelSync
BANK 0
PUBLIC Start, AfterLdA, Scratch, AfterDb

Start:
    ld   a, (Scratch)
AfterLdA:
    nop
Scratch: db 00h
AfterDb:
    nop
`
	asm := NewAssembler()
	obj, err := asm.Assemble(src)
	if err != nil {
		t.Fatalf("Assemble failed: %v", err)
	}

	offsets := make(map[string]uint16)
	for _, sym := range obj.Symbols {
		offsets[sym.Name] = sym.Offset
	}

	// LD A,(nn) = 3 bytes (0x3A + endereco de 16 bits)
	if offsets["AfterLdA"] != 3 {
		t.Errorf("AfterLdA: esperado offset 3 (LD A,(nn) = 3 bytes), obtido %d", offsets["AfterLdA"])
	}
	// + NOP (1 byte)
	if offsets["Scratch"] != 4 {
		t.Errorf("Scratch: esperado offset 4, obtido %d", offsets["Scratch"])
	}
	// DB 00h = exatamente 1 byte, nao 2
	if offsets["AfterDb"] != 5 {
		t.Errorf("AfterDb: esperado offset 5 (DB 00h = 1 byte), obtido %d", offsets["AfterDb"])
	}
}

// TestLdRegisterFromAbsoluteAddressRejected: só "LD A,(nn)" tem forma de
// endereçamento absoluto de 16 bits pra um registrador de 8 bits no Z80 de
// verdade -- "LD B,(nn)"/"LD C,(nn)"/etc. não existem. Sem uma checagem
// explícita, isso caía no fallback de imediato (parseImm8, que não tem como
// reportar erro) e montava em silêncio como "LD B, 0" -- bug real
// encontrado em lib/src/float.asm (Float_Cmp32), onde 3 ocorrências de
// "LD B,(Float_X)" quebravam comparações sempre que o fluxo as alcançava,
// sem nenhum erro de montagem. Agora deve ser um erro de compilação claro.
func TestLdRegisterFromAbsoluteAddressRejected(t *testing.T) {
	src := `
MODULE BadLd
BANK 0
PUBLIC Start
Scratch: DB 00h
Start:
    LD B, (Scratch)
    RET
`
	asm := NewAssembler()
	_, err := asm.Assemble(src)
	if err == nil {
		t.Fatal("esperado erro de montagem para 'LD B,(Scratch)', mas montou com sucesso")
	}
}

func TestLdAFromAbsoluteAddressStillWorks(t *testing.T) {
	src := `
MODULE GoodLd
BANK 0
PUBLIC Start
Scratch: DB 00h
Start:
    LD A, (Scratch)
    RET
`
	asm := NewAssembler()
	if _, err := asm.Assemble(src); err != nil {
		t.Fatalf("LD A,(Scratch) deveria continuar funcionando, mas falhou: %v", err)
	}
}

// TestMsxlibModulesAssembleConsistently monta todos os fontes da MSXLIB e
// depende da verificacao interna de consistencia Pass1/Pass2 dentro de
// Assemble() para pegar qualquer futura dessincronia de tamanho de
// instrucao antes que ela corrompa silenciosamente algum rotulo.
func TestMsxlibModulesAssembleConsistently(t *testing.T) {
	libDir := filepath.Join("..", "..", "lib", "src")
	files := []string{"bdos.asm", "bios.asm", "vdp.asm", "psg.asm", "string.asm", "math.asm", "float.asm"}

	for _, f := range files {
		t.Run(f, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join(libDir, f))
			if err != nil {
				t.Fatalf("failed to read %s: %v", f, err)
			}
			asm := NewAssembler()
			if _, err := asm.Assemble(string(data)); err != nil {
				t.Fatalf("Assemble(%s) failed: %v", f, err)
			}
		})
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

