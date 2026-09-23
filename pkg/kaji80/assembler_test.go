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

// TestMalformedAddressOperandRejected: auditoria feita depois do bug real de
// "LD B,(nn)" (achado em lib/src/float.asm, testado em hardware) -- qualquer
// operando de endereço que não seja número/constante EQU/nome de símbolo
// válido (ex.: um operando de memória mal-formado como "(Algo)" acabando
// onde um símbolo era esperado) agora é um erro de compilação claro em
// CALL/JP/LD rr,nn, não um "símbolo" fantasma silenciosamente aceito que só
// falharia (ou pior, resolveria por acidente) na hora da linkagem.
func TestMalformedAddressOperandRejected(t *testing.T) {
	cases := []string{
		"    CALL (Algo)\n",
		"    JP (Algo)\n",
		"    LD DE, (Algo)\n",
		"    LD BC, (Algo)\n",
	}
	for _, instr := range cases {
		src := "MODULE BadAddr\nBANK 0\nPUBLIC Start\nAlgo: DB 00h\nStart:\n" + instr
		asm := NewAssembler()
		if _, err := asm.Assemble(src); err == nil {
			t.Errorf("esperado erro de montagem para %q, mas montou com sucesso", instr)
		}
	}
}

// TestAluIndirectAbsoluteAddressRejected: mesma classe de bug do teste
// acima, mas no lado das operações ALU de 8 bits (ADD/ADC/SUB/SBC/AND/XOR/
// OR/CP) -- o Z80 só tem forma indireta via (HL) ou (IX+d)/(IY+d), nunca
// endereço absoluto. Sem a checagem, caía no fallback de imediato
// (parseImm8, sem como reportar erro) e virava "CP 0" em silêncio -- mesma
// causa raiz do bug gráfico de SCREEN 2 já documentado para o caso
// (IX+d)/(IY+d); este teste cobre o caso de endereço absoluto puro.
func TestAluIndirectAbsoluteAddressRejected(t *testing.T) {
	src := `
MODULE BadAlu
BANK 0
PUBLIC Start
Algo: DB 00h
Start:
    CP (Algo)
    RET
`
	asm := NewAssembler()
	if _, err := asm.Assemble(src); err == nil {
		t.Fatal("esperado erro de montagem para 'CP (Algo)', mas montou com sucesso")
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

// TestEquWithExpression: EQU passa a aceitar uma expressão completa, não só
// um literal único -- resolve o exemplo motivador do Wilson
// ((2*8)/(1+3))<<2.
func TestEquWithExpression(t *testing.T) {
	src := `
MODULE EquExpr
BANK 0
PUBLIC Start
VAL EQU ((2*8)/(1+3))<<2
Start:
    LD A, VAL
    RET
`
	asm := NewAssembler()
	obj, err := asm.Assemble(src)
	if err != nil {
		t.Fatalf("Assemble failed: %v", err)
	}
	// LD A,n = 3E nn -- VAL deve valer 16 (0x10).
	expected := []byte{0x3E, 0x10, 0xC9}
	if !bytes.Equal(obj.Segments[0].Data, expected) {
		t.Fatalf("Byte mismatch:\nGot:      % X\nExpected: % X", obj.Segments[0].Data, expected)
	}
}

// TestEquWithUnknownSymbolErrors: EQU não aceita mais nome de símbolo solto
// silenciosamente virando zero (comportamento antigo de parseConstant) --
// agora é um erro de compilação claro.
func TestEquWithUnknownSymbolErrors(t *testing.T) {
	src := `
MODULE BadEqu
BANK 0
PUBLIC Start
VAL EQU RotuloQueNaoExiste
Start:
    RET
`
	asm := NewAssembler()
	if _, err := asm.Assemble(src); err == nil {
		t.Fatal("esperado erro de montagem para EQU com símbolo desconhecido, mas montou com sucesso")
	}
}

// TestVariableAssignAndReassign: "Nome = expressão" -- variável
// reatribuível, reavaliada em cada DB subsequente na ordem sequencial em
// que aparecem (não só o valor final depois de todo o Pass 1 rodar).
func TestVariableAssignAndReassign(t *testing.T) {
	src := `
MODULE VarAssign
BANK 0
PUBLIC Start
Start:
X = 5
    DB X
X = X + 1
    DB X
X = X * 10
    DB X
    RET
`
	asm := NewAssembler()
	obj, err := asm.Assemble(src)
	if err != nil {
		t.Fatalf("Assemble failed: %v", err)
	}
	expected := []byte{5, 6, 60, 0xC9}
	if !bytes.Equal(obj.Segments[0].Data, expected) {
		t.Fatalf("Byte mismatch:\nGot:      % X\nExpected: % X", obj.Segments[0].Data, expected)
	}
}

// TestDbWithMultiArgFunctionExpression: confirma que parseLine agora separa
// operandos por vírgula respeitando profundidade de parênteses -- antes
// desta correção, "DB POW(2,3)" quebrava incorretamente em dois operandos
// ("POW(2" e "3)") em vez de um único operando de expressão.
func TestDbWithMultiArgFunctionExpression(t *testing.T) {
	src := `
MODULE DbPow
BANK 0
PUBLIC Start
Start:
    DB POW(2,3), 1+1, "AB"
    RET
`
	asm := NewAssembler()
	obj, err := asm.Assemble(src)
	if err != nil {
		t.Fatalf("Assemble failed: %v", err)
	}
	expected := []byte{8, 2, 'A', 'B', 0xC9}
	if !bytes.Equal(obj.Segments[0].Data, expected) {
		t.Fatalf("Byte mismatch:\nGot:      % X\nExpected: % X", obj.Segments[0].Data, expected)
	}
}

// TestDwWithExpressionAndSymbol: DW aceita uma expressão numérica pura
// (nova nesta leva) e continua aceitando, sem quebrar, um símbolo/rótulo
// comum (comportamento pré-existente, resolvido só na linkagem via
// relocation -- por isso o segundo operando fica reservado como 00 00 até
// o MUSUBI resolver, não testável só com Assemble()).
func TestDwWithExpressionAndSymbol(t *testing.T) {
	src := `
MODULE DwExpr
BANK 0
PUBLIC Start, Alvo
Start:
    DW 1+2, Alvo
Alvo:
    RET
`
	asm := NewAssembler()
	obj, err := asm.Assemble(src)
	if err != nil {
		t.Fatalf("Assemble failed: %v", err)
	}
	data := obj.Segments[0].Data
	// 1+2=3 (03 00 little-endian) resolvido em tempo de montagem; os 2
	// bytes seguintes (endereço de Alvo) ficam reservados como 00 00,
	// preenchidos só na linkagem via relocation.
	expected := []byte{0x03, 0x00, 0x00, 0x00, 0xC9}
	if !bytes.Equal(data, expected) {
		t.Fatalf("Byte mismatch:\nGot:      % X\nExpected: % X", data, expected)
	}
	if len(obj.Relocations) != 1 {
		t.Fatalf("esperada 1 relocation (pro operando 'Alvo'), obtida %d", len(obj.Relocations))
	}
}

// TestLocalLabelMatchesHandWrittenMangledName: ".loop:" dentro de "Start:"
// deve virar exatamente "Start_loop" -- compara byte a byte contra a
// versão escrita à mão com o nome já mesclado, que é o comportamento
// esperado documentado no plano.
func TestLocalLabelMatchesHandWrittenMangledName(t *testing.T) {
	srcLocal := `
MODULE LocalLbl
BANK 0
PUBLIC Start
Start:
.loop:
    NOP
    JR .loop
    RET
`
	srcMangled := `
MODULE LocalLblMangled
BANK 0
PUBLIC Start
Start:
Start_loop:
    NOP
    JR Start_loop
    RET
`
	asmLocal := NewAssembler()
	objLocal, err := asmLocal.Assemble(srcLocal)
	if err != nil {
		t.Fatalf("Assemble (rótulo local) falhou: %v", err)
	}
	asmMangled := NewAssembler()
	objMangled, err := asmMangled.Assemble(srcMangled)
	if err != nil {
		t.Fatalf("Assemble (nome já mesclado à mão) falhou: %v", err)
	}
	if !bytes.Equal(objLocal.Segments[0].Data, objMangled.Segments[0].Data) {
		t.Fatalf("Byte mismatch:\nRótulo local:   % X\nNome mesclado:  % X", objLocal.Segments[0].Data, objMangled.Segments[0].Data)
	}
}

// TestLocalLabelNoCollisionAcrossScopes: o mesmo nome ".loop" em duas
// funções diferentes não deve colidir -- cada uma resolve dentro do seu
// próprio escopo (Function1_loop / Function2_loop).
func TestLocalLabelNoCollisionAcrossScopes(t *testing.T) {
	src := `
MODULE TwoScopes
BANK 0
PUBLIC Function1, Function2
Function1:
.loop:
    NOP
    JR .loop
Function2:
.loop:
    NOP
    JR .loop
    RET
`
	asm := NewAssembler()
	obj, err := asm.Assemble(src)
	if err != nil {
		t.Fatalf("Assemble failed: %v", err)
	}
	// NOP(1) + JR -3(2) repetido duas vezes (cada bloco salta pro seu
	// próprio .loop, não pro do outro escopo) + RET.
	expected := []byte{0x00, 0x18, 0xFD, 0x00, 0x18, 0xFD, 0xC9}
	if !bytes.Equal(obj.Segments[0].Data, expected) {
		t.Fatalf("Byte mismatch:\nGot:      % X\nExpected: % X", obj.Segments[0].Data, expected)
	}
}

// TestLocalLabelBeforeAnyGlobalErrors: um rótulo local usado antes de
// qualquer rótulo global no arquivo é um erro de compilação claro, não um
// símbolo fantasma silencioso.
func TestLocalLabelBeforeAnyGlobalErrors(t *testing.T) {
	src := `
MODULE BadLocal
BANK 0
.loop:
    NOP
`
	asm := NewAssembler()
	if _, err := asm.Assemble(src); err == nil {
		t.Fatal("esperado erro de montagem para rótulo local antes de qualquer rótulo global, mas montou com sucesso")
	}
}

// TestLocalLabelPreservesCase: o mangle preserva maiúsculas/minúsculas
// exatamente como escrito, sem forçar minúsculo.
func TestLocalLabelPreservesCase(t *testing.T) {
	src := `
MODULE CaseLocal
BANK 0
PUBLIC Rotina
Rotina:
.Loop:
    NOP
    JR .Loop
    RET
`
	asm := NewAssembler()
	obj, err := asm.Assemble(src)
	if err != nil {
		t.Fatalf("Assemble failed: %v", err)
	}
	// JR local resolve em tempo de montagem (REL8, não gera relocation) --
	// se a definição ".Loop:" e a referência "JR .Loop" não tivessem
	// mesclado pro MESMO nome com o MESMO case ("Rotina_Loop" nos dois),
	// o alvo não seria encontrado como rótulo local e o assembler geraria
	// uma relocation em vez de resolver o deslocamento relativo aqui.
	expected := []byte{0x00, 0x18, 0xFD, 0xC9} // NOP, JR -3, RET
	if !bytes.Equal(obj.Segments[0].Data, expected) {
		t.Fatalf("Byte mismatch (indica que o case não foi preservado igual entre definição e referência):\nGot:      % X\nExpected: % X", obj.Segments[0].Data, expected)
	}
	if len(obj.Relocations) != 0 {
		t.Fatalf("esperada 0 relocations (JR local deveria resolver em tempo de montagem), obtida %d", len(obj.Relocations))
	}
}

// TestIfTrueKeepsCode: condição não-zero mantém o bloco IF, sem ELSE.
func TestIfTrueKeepsCode(t *testing.T) {
	src := `
MODULE IfTrue
BANK 0
PUBLIC Start
Start:
IF 1
    NOP
ENDIF
    RET
`
	asm := NewAssembler()
	obj, err := asm.Assemble(src)
	if err != nil {
		t.Fatalf("Assemble failed: %v", err)
	}
	expected := []byte{0x00, 0xC9}
	if !bytes.Equal(obj.Segments[0].Data, expected) {
		t.Fatalf("Byte mismatch:\nGot:      % X\nExpected: % X", obj.Segments[0].Data, expected)
	}
}

// TestIfFalseElseBranch: condição zero descarta o bloco IF e mantém o ELSE.
func TestIfFalseElseBranch(t *testing.T) {
	src := `
MODULE IfElse
BANK 0
PUBLIC Start
Start:
IF 0
    NOP
    NOP
ELSE
    HALT
ENDIF
    RET
`
	asm := NewAssembler()
	obj, err := asm.Assemble(src)
	if err != nil {
		t.Fatalf("Assemble failed: %v", err)
	}
	expected := []byte{0x76, 0xC9} // HALT, RET
	if !bytes.Equal(obj.Segments[0].Data, expected) {
		t.Fatalf("Byte mismatch:\nGot:      % X\nExpected: % X", obj.Segments[0].Data, expected)
	}
}

// TestIfReferencesEarlierConstant: exemplo canônico -- "FORMATO=1 / IF
// FORMATO==1 ..." -- condição IF referenciando uma variável definida mais
// acima no mesmo arquivo precisa enxergar o valor certo mesmo rodando
// antes do Pass 1 de verdade.
func TestIfReferencesEarlierConstant(t *testing.T) {
	src := `
MODULE IfConst
BANK 0
PUBLIC Start
FORMATO = 1
Start:
IF FORMATO==1
    NOP
ELSE
    HALT
ENDIF
    RET
`
	asm := NewAssembler()
	obj, err := asm.Assemble(src)
	if err != nil {
		t.Fatalf("Assemble failed: %v", err)
	}
	expected := []byte{0x00, 0xC9}
	if !bytes.Equal(obj.Segments[0].Data, expected) {
		t.Fatalf("Byte mismatch:\nGot:      % X\nExpected: % X", obj.Segments[0].Data, expected)
	}
}

// TestIfNested: aninhamento de IF dentro de IF, sem limite artificial.
func TestIfNested(t *testing.T) {
	src := `
MODULE IfNested
BANK 0
PUBLIC Start
Start:
IF 1
    IF 0
        HALT
    ELSE
        NOP
    ENDIF
ENDIF
    RET
`
	asm := NewAssembler()
	obj, err := asm.Assemble(src)
	if err != nil {
		t.Fatalf("Assemble failed: %v", err)
	}
	expected := []byte{0x00, 0xC9}
	if !bytes.Equal(obj.Segments[0].Data, expected) {
		t.Fatalf("Byte mismatch:\nGot:      % X\nExpected: % X", obj.Segments[0].Data, expected)
	}
}

// TestIfDeadBranchSkipsEvaluation: um IF aninhado dentro de um ramo já
// morto não deve tentar avaliar a condição (que pode referenciar algo
// nunca definido) -- só precisa casar corretamente com seu ENDIF.
func TestIfDeadBranchSkipsEvaluation(t *testing.T) {
	src := `
MODULE IfDeadBranch
BANK 0
PUBLIC Start
Start:
IF 0
    IF NuncaDefinido
        HALT
    ENDIF
ENDIF
    RET
`
	asm := NewAssembler()
	obj, err := asm.Assemble(src)
	if err != nil {
		t.Fatalf("Assemble failed (não deveria tentar avaliar condição em ramo morto): %v", err)
	}
	expected := []byte{0xC9}
	if !bytes.Equal(obj.Segments[0].Data, expected) {
		t.Fatalf("Byte mismatch:\nGot:      % X\nExpected: % X", obj.Segments[0].Data, expected)
	}
}

// TestIfElseWithoutIfErrors / TestEndifWithoutIfErrors / TestIfWithoutEndifErrors:
// erros de estrutura claros, não montagem silenciosamente errada.
func TestIfElseWithoutIfErrors(t *testing.T) {
	src := "MODULE Bad\nBANK 0\nPUBLIC Start\nStart:\nELSE\n    RET\n"
	asm := NewAssembler()
	if _, err := asm.Assemble(src); err == nil {
		t.Fatal("esperado erro para ELSE sem IF correspondente")
	}
}

func TestEndifWithoutIfErrors(t *testing.T) {
	src := "MODULE Bad\nBANK 0\nPUBLIC Start\nStart:\nENDIF\n    RET\n"
	asm := NewAssembler()
	if _, err := asm.Assemble(src); err == nil {
		t.Fatal("esperado erro para ENDIF sem IF correspondente")
	}
}

func TestIfWithoutEndifErrors(t *testing.T) {
	src := "MODULE Bad\nBANK 0\nPUBLIC Start\nStart:\nIF 1\n    RET\n"
	asm := NewAssembler()
	if _, err := asm.Assemble(src); err == nil {
		t.Fatal("esperado erro para IF sem ENDIF correspondente")
	}
}

// TestDeftAlias: DT/DEFT são aliases de DB pra literais de texto, pedidos
// explicitamente por Wilson (compatibilidade com outros assemblers Z80).
func TestDeftAlias(t *testing.T) {
	src := `
MODULE DeftAlias
BANK 0
PUBLIC Start
Start:
    DEFT "OK"
    RET
`
	asm := NewAssembler()
	obj, err := asm.Assemble(src)
	if err != nil {
		t.Fatalf("Assemble failed: %v", err)
	}
	expected := []byte{'O', 'K', 0xC9}
	if !bytes.Equal(obj.Segments[0].Data, expected) {
		t.Fatalf("Byte mismatch:\nGot:      % X\nExpected: % X", obj.Segments[0].Data, expected)
	}
}

