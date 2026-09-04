package musubi

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

	"github.com/wilsonpilon/kizuna/pkg/hako"
	"github.com/wilsonpilon/kizuna/pkg/mob"
)

func TestLinkTwoModules(t *testing.T) {
	// Módulo 1: Main
	// Código: CALL SubFunc; RET
	// 0xCD 0x00 0x00, 0xC9 (4 bytes)
	mod1 := mob.NewObjectFile()
	seg1 := mod1.AddSegment(mob.SegmentCode, 0, []byte{0xCD, 0x00, 0x00, 0xC9}, 0)
	symStart := mod1.AddSymbol("Start", mob.SymbolPublic, mob.SymbolProc, seg1, 0x0000)
	symSub := mod1.AddSymbol("SubFunc", mob.SymbolExtern, mob.SymbolProc, 0, 0)
	mod1.AddRelocation(seg1, 1, symSub, mob.RelocAbs16) // Offset 1 é o endereço de 16-bit do CALL

	_ = symStart

	// Módulo 2: Util
	// Código: LD A, 10; RET
	// 0x3E 0x0A, 0xC9 (3 bytes)
	mod2 := mob.NewObjectFile()
	seg2 := mod2.AddSegment(mob.SegmentCode, 0, []byte{0x3E, 0x0A, 0xC9}, 0)
	mod2.AddSymbol("SubFunc", mob.SymbolPublic, mob.SymbolProc, seg2, 0x0000)

	cfg := LinkerConfig{
		BaseAddress: 0x0100,
		EntryPoint:  "Start",
	}
	linker := NewLinker(cfg)
	res, err := linker.Link(mod1, mod2)
	if err != nil {
		t.Fatalf("Link failed: %v", err)
	}

	// Tamanho esperado: 4 bytes (mod1) + 3 bytes (mod2) = 7 bytes
	if res.TotalSize != 7 {
		t.Fatalf("Expected 7 bytes, got %d", res.TotalSize)
	}

	// SubFunc deve estar posicionado logo após o mod1: 0x0100 + 4 = 0x0104
	subSym := res.Symbols["SubFunc"]
	if subSym == nil || subSym.Address != 0x0104 {
		t.Fatalf("Expected SubFunc at 0x0104, got %v", subSym)
	}

	// O CALL no mod1 (offset 1 e 2) deve ter sido realocado para 0x0104
	callTarget := binary.LittleEndian.Uint16(res.Binary[1:3])
	if callTarget != 0x0104 {
		t.Errorf("Expected CALL target 0x0104, got 0x%04X", callTarget)
	}
}

func TestLinkSampleHello(t *testing.T) {
	helloMobPath := filepath.Join("..", "..", "sample", "hello.mob")
	if _, err := os.Stat(helloMobPath); os.IsNotExist(err) {
		t.Skip("sample/hello.mob não encontrado, ignorando teste")
	}

	obj, err := mob.LoadFromFile(helloMobPath)
	if err != nil {
		t.Fatalf("Failed to load hello.mob: %v", err)
	}

	tempDir := t.TempDir()
	mapFile := filepath.Join(tempDir, "hello.map")

	cfg := LinkerConfig{
		BaseAddress: 0x0100,
		EntryPoint:  "Start",
		MapFile:     mapFile,
	}

	linker := NewLinker(cfg)
	res, err := linker.Link(obj)
	if err != nil {
		t.Fatalf("Link failed for hello.mob: %v", err)
	}

	// Start deve ser 0x0100
	startSym := res.Symbols["Start"]
	if startSym.Address != 0x0100 {
		t.Errorf("Expected Start at 0x0100, got 0x%04X", startSym.Address)
	}

	// MsgHello deve ser 0x0100 + 9 = 0x0109
	msgSym := res.Symbols["MsgHello"]
	if msgSym.Address != 0x0109 {
		t.Errorf("Expected MsgHello at 0x0109, got 0x%04X", msgSym.Address)
	}

	// A instrução 'LD DE, MsgHello' (0x11 nnL nnH) deve apontar para 0x0109
	ldDeTarget := binary.LittleEndian.Uint16(res.Binary[1:3])
	if ldDeTarget != 0x0109 {
		t.Errorf("Expected LD DE to point to 0x0109, got 0x%04X", ldDeTarget)
	}

	// Verificar se o arquivo .map foi gerado
	mapContent, err := os.ReadFile(mapFile)
	if err != nil {
		t.Fatalf("Failed to read map file: %v", err)
	}
	if len(mapContent) == 0 {
		t.Errorf("Map file is empty")
	}
}

func TestMultiBankTrampolineGeneration(t *testing.T) {
	// Módulo 1: Área Comum (Banco 0)
	// CALL Bank1Func; RET
	mod0 := mob.NewObjectFile()
	seg0 := mod0.AddSegment(mob.SegmentCode, 0, []byte{0xCD, 0x00, 0x00, 0xC9}, 0)
	mod0.AddSymbol("Start", mob.SymbolPublic, mob.SymbolProc, seg0, 0x0000)
	symB1 := mod0.AddSymbol("Bank1Func", mob.SymbolExtern, mob.SymbolProc, 0, 0)
	mod0.AddRelocation(seg0, 1, symB1, mob.RelocAbs16)

	// Módulo 2: Banco Paginável 1 (Página 2, 0x8000)
	// LD A, 10; RET
	mod1 := mob.NewObjectFile()
	seg1 := mod1.AddSegment(mob.SegmentCode, 1, []byte{0x3E, 0x0A, 0xC9}, 0)
	mod1.AddSymbol("Bank1Func", mob.SymbolPublic, mob.SymbolProc, seg1, 0x0000)

	cfg := LinkerConfig{
		BaseAddress: 0x0100,
		EntryPoint:  "Start",
	}
	linker := NewLinker(cfg)
	res, err := linker.Link(mod0, mod1)
	if err != nil {
		t.Fatalf("Link failed: %v", err)
	}

	if !res.IsMultiBank {
		t.Errorf("Expected IsMultiBank to be true")
	}

	// Símbolo Bank1Func deve estar em 0x8000 no Banco 1
	sym := res.Symbols["Bank1Func"]
	if sym.Address != 0x8000 || sym.Bank != 1 {
		t.Fatalf("Expected Bank1Func at 0x8000 on Bank 1, got 0x%04X on Bank %d", sym.Address, sym.Bank)
	}

	// Deve ter gerado 1 trampolim para Bank1Func
	if len(res.Trampolines) != 1 {
		t.Fatalf("Expected 1 trampoline, got %d", len(res.Trampolines))
	}

	tramp, ok := res.Trampolines["Bank1Func"]
	if !ok {
		t.Fatalf("Trampoline for Bank1Func not found")
	}

	// O trampolim deve morar na Área Comum
	if tramp.Address < 0x0100 || tramp.Address >= 0x8000 {
		t.Errorf("Trampoline address 0x%04X outside common area", tramp.Address)
	}

	// O CALL do mod0 (alocado no Segmento 1, após o Bootstrap) no offset 1 deve apontar para o trampolim
	callTarget := binary.LittleEndian.Uint16(res.Segments[1].Data[1:3])
	if callTarget != tramp.Address {
		t.Errorf("Expected CALL in mod0 to point to trampoline 0x%04X, got 0x%04X", tramp.Address, callTarget)
	}

	// O trampolim deve ter 19 bytes e chavear para o Banco 1 chamando Musubi_PutP2 e a rotina alvo (0x8000):
	if len(tramp.Code) != 19 {
		t.Fatalf("Expected trampoline length 19, got %d", len(tramp.Code))
	}
	if tramp.Code[3] != 0xF5 || tramp.Code[4] != 0x3E || tramp.Code[5] != 0x01 {
		t.Errorf("Expected PUSH AF (0xF5) and LD A, 1 (0x3E 0x01), got %02X %02X %02X", tramp.Code[3], tramp.Code[4], tramp.Code[5])
	}
	if tramp.Code[9] != 0xCD || tramp.Code[10] != 0x00 || tramp.Code[11] != 0x80 {
		t.Errorf("Expected CALL 0x8000 (0xCD 0x00 0x80), got %02X %02X %02X", tramp.Code[9], tramp.Code[10], tramp.Code[11])
	}
}

func TestIntraBankCallNoTrampoline(t *testing.T) {
	// Módulo no Banco 1: FuncA chama FuncB (mesmo banco)
	mod := mob.NewObjectFile()
	seg := mod.AddSegment(mob.SegmentCode, 1, []byte{
		0xCD, 0x00, 0x00, // CALL FuncB (offset 0)
		0xC9,             // RET (offset 3)
		0x3E, 0x2A,       // FuncB: LD A, 42 (offset 4)
		0xC9,             // RET (offset 6)
	}, 0)

	symA := mod.AddSymbol("FuncA", mob.SymbolPublic, mob.SymbolProc, seg, 0x0000)
	symB := mod.AddSymbol("FuncB", mob.SymbolPublic, mob.SymbolProc, seg, 0x0004)
	mod.AddRelocation(seg, 1, symB, mob.RelocAbs16)
	_ = symA

	linker := NewLinker(DefaultConfig())
	res, err := linker.Link(mod)
	if err != nil {
		t.Fatalf("Link failed: %v", err)
	}

	// NÃO deve haver trampolins gerados
	if len(res.Trampolines) != 0 {
		t.Errorf("Expected 0 trampolines for intra-bank call, got %d", len(res.Trampolines))
	}

	// O CALL deve apontar direto para FuncB em 0x8004 (no Segmento 1 do Banco 1)
	callTarget := binary.LittleEndian.Uint16(res.Segments[1].Data[1:3])
	if callTarget != 0x8004 {
		t.Errorf("Expected intra-bank CALL target 0x8004, got 0x%04X", callTarget)
	}
}

func TestUnresolvedSymbolError(t *testing.T) {
	mod := mob.NewObjectFile()
	seg := mod.AddSegment(mob.SegmentCode, 0, []byte{0xCD, 0x00, 0x00}, 0)
	sym := mod.AddSymbol("MissingFunc", mob.SymbolExtern, mob.SymbolProc, 0, 0)
	mod.AddRelocation(seg, 1, sym, mob.RelocAbs16)

	linker := NewLinker(DefaultConfig())
	_, err := linker.Link(mod)
	if err == nil {
		t.Error("Expected error for unresolved symbol, got nil")
	}
}

func TestDuplicateSymbolError(t *testing.T) {
	mod1 := mob.NewObjectFile()
	seg1 := mod1.AddSegment(mob.SegmentCode, 0, []byte{0xC9}, 0)
	mod1.AddSymbol("SameName", mob.SymbolPublic, mob.SymbolProc, seg1, 0)

	mod2 := mob.NewObjectFile()
	seg2 := mod2.AddSegment(mob.SegmentCode, 0, []byte{0xC9}, 0)
	mod2.AddSymbol("SameName", mob.SymbolPublic, mob.SymbolProc, seg2, 0)

	linker := NewLinker(DefaultConfig())
	_, err := linker.Link(mod1, mod2)
	if err == nil {
		t.Error("Expected error for duplicate symbol, got nil")
	}
}

func TestSmartLinkingFromHlib(t *testing.T) {
	tmpDir := t.TempDir()

	// 1. Módulo math_add (exporta Add16 e depende transitivamente de HelperFunc)
	// CALL HelperFunc; ADD HL, DE; RET (0xCD 0x00 0x00, 0x19, 0xC9)
	modAdd := mob.NewObjectFile()
	segAdd := modAdd.AddSegment(mob.SegmentCode, 0, []byte{0xCD, 0x00, 0x00, 0x19, 0xC9}, 0)
	modAdd.AddSymbol("Add16", mob.SymbolPublic, mob.SymbolProc, segAdd, 0)
	symHelperExt := modAdd.AddSymbol("HelperFunc", mob.SymbolExtern, mob.SymbolProc, 0, 0)
	modAdd.AddRelocation(segAdd, 1, symHelperExt, mob.RelocAbs16)
	addPath := filepath.Join(tmpDir, "math_add.mob")
	if err := mob.SaveToFile(addPath, modAdd); err != nil {
		t.Fatalf("Failed to save math_add.mob: %v", err)
	}

	// 2. Módulo math_helper (exporta HelperFunc)
	// NOP; RET (0x00, 0xC9)
	modHelper := mob.NewObjectFile()
	segHelper := modHelper.AddSegment(mob.SegmentCode, 0, []byte{0x00, 0xC9}, 0)
	modHelper.AddSymbol("HelperFunc", mob.SymbolPublic, mob.SymbolProc, segHelper, 0)
	helperPath := filepath.Join(tmpDir, "math_helper.mob")
	if err := mob.SaveToFile(helperPath, modHelper); err != nil {
		t.Fatalf("Failed to save math_helper.mob: %v", err)
	}

	// 3. Módulo math_sub (exporta Sub16 - NÃO deve ser incluído na linkagem!)
	// OR A; SBC HL, DE; RET (0xB7, 0xED, 0x52, 0xC9)
	modSub := mob.NewObjectFile()
	segSub := modSub.AddSegment(mob.SegmentCode, 0, []byte{0xB7, 0xED, 0x52, 0xC9}, 0)
	modSub.AddSymbol("Sub16", mob.SymbolPublic, mob.SymbolProc, segSub, 0)
	subPath := filepath.Join(tmpDir, "math_sub.mob")
	if err := mob.SaveToFile(subPath, modSub); err != nil {
		t.Fatalf("Failed to save math_sub.mob: %v", err)
	}

	// 4. Empacotar math.hlib contendo math_add, math_helper e math_sub
	hlibPath := filepath.Join(tmpDir, "math.hlib")
	if err := hako.Pack(hlibPath, addPath, helperPath, subPath); err != nil {
		t.Fatalf("hako.Pack failed: %v", err)
	}

	// 5. Módulo principal: chama apenas Add16
	// CALL Add16; RET (0xCD 0x00 0x00, 0xC9)
	modMain := mob.NewObjectFile()
	segMain := modMain.AddSegment(mob.SegmentCode, 0, []byte{0xCD, 0x00, 0x00, 0xC9}, 0)
	modMain.AddSymbol("Start", mob.SymbolPublic, mob.SymbolProc, segMain, 0)
	symAddExt := modMain.AddSymbol("Add16", mob.SymbolExtern, mob.SymbolProc, 0, 0)
	modMain.AddRelocation(segMain, 1, symAddExt, mob.RelocAbs16)
	mainPath := filepath.Join(tmpDir, "main.mob")
	if err := mob.SaveToFile(mainPath, modMain); err != nil {
		t.Fatalf("Failed to save main.mob: %v", err)
	}

	// 6. Linkar main.mob com math.hlib
	outCom := filepath.Join(tmpDir, "app.com")
	cfg := DefaultConfig()
	res, err := LinkToFile(outCom, cfg, mainPath, hlibPath)
	if err != nil {
		t.Fatalf("LinkToFile with .hlib failed: %v", err)
	}

	// 7. Validações
	// Deve ter resolvido Add16 e HelperFunc transitivamente
	if res.Symbols["Add16"] == nil {
		t.Error("Expected Add16 to be resolved from .hlib")
	}
	if res.Symbols["HelperFunc"] == nil {
		t.Error("Expected HelperFunc to be resolved transitively from .hlib")
	}

	// Sub16 NÃO deve estar presente (Dead-Code Elimination / Smart-Linking)
	if res.Symbols["Sub16"] != nil {
		t.Errorf("Sub16 was included, but should have been eliminated as dead code! Symbol: %v", res.Symbols["Sub16"])
	}

	// O arquivo executável deve existir e ter o tamanho exato de:
	// main (4 bytes) + math_add (5 bytes) + math_helper (2 bytes) = 11 bytes
	expectedSize := 4 + 5 + 2
	if res.TotalSize != expectedSize {
		t.Errorf("Expected TotalSize %d bytes, got %d", expectedSize, res.TotalSize)
	}

	// Verificar se a chamada para Add16 aponta para o endereço correto (0x0100 + 4 = 0x0104)
	callAddTarget := binary.LittleEndian.Uint16(res.Binary[1:3])
	if callAddTarget != 0x0104 {
		t.Errorf("Expected CALL Add16 target 0x0104, got 0x%04X", callAddTarget)
	}
}

