package hako

import (
	"path/filepath"
	"testing"

	"github.com/wilsonpilon/kizuna/pkg/mob"
)

func TestHakoPackAndRead(t *testing.T) {
	tempDir := t.TempDir()

	// 1. Criar Módulo 1: math_add
	mod1 := mob.NewObjectFile()
	seg1 := mod1.AddSegment(mob.SegmentCode, 0, []byte{0x3E, 0x05, 0xC9}, 0)
	mod1.AddSymbol("Add16", mob.SymbolPublic, mob.SymbolProc, seg1, 0)
	mob1Path := filepath.Join(tempDir, "math_add.mob")
	if err := mob.SaveToFile(mob1Path, mod1); err != nil {
		t.Fatalf("failed to save mob1: %v", err)
	}

	// 2. Criar Módulo 2: math_sub
	mod2 := mob.NewObjectFile()
	seg2 := mod2.AddSegment(mob.SegmentCode, 0, []byte{0x3E, 0x03, 0xC9}, 0)
	mod2.AddSymbol("Sub16", mob.SymbolPublic, mob.SymbolProc, seg2, 0)
	mod2.AddSymbol("MathData", mob.SymbolPublic, mob.SymbolData, seg2, 2)
	mob2Path := filepath.Join(tempDir, "math_sub.mob")
	if err := mob.SaveToFile(mob2Path, mod2); err != nil {
		t.Fatalf("failed to save mob2: %v", err)
	}

	// 3. Empacotar em math.hlib
	libPath := filepath.Join(tempDir, "math.hlib")
	if err := Pack(libPath, mob1Path, mob2Path); err != nil {
		t.Fatalf("Pack failed: %v", err)
	}

	// 4. Abrir biblioteca e verificar conteúdo
	arc, err := Open(libPath)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}

	if len(arc.Modules) != 2 {
		t.Fatalf("expected 2 modules, got %d", len(arc.Modules))
	}

	if arc.Modules[0].Name != "math_add" || arc.Modules[1].Name != "math_sub" {
		t.Errorf("module names mismatch: got [%s, %s]", arc.Modules[0].Name, arc.Modules[1].Name)
	}

	// 5. Testar Dicionário Global de Símbolos
	symAdd, found := arc.FindModuleForSymbol("Add16")
	if !found {
		t.Errorf("symbol Add16 not found in dictionary")
	} else if symAdd.ModuleName != "math_add" || symAdd.Kind != mob.SymbolProc {
		t.Errorf("symbol Add16 info mismatch: %+v", symAdd)
	}

	symSub, found := arc.FindModuleForSymbol("Sub16")
	if !found {
		t.Errorf("symbol Sub16 not found in dictionary")
	} else if symSub.ModuleName != "math_sub" || symSub.Kind != mob.SymbolProc {
		t.Errorf("symbol Sub16 info mismatch: %+v", symSub)
	}

	symData, found := arc.FindModuleForSymbol("MathData")
	if !found {
		t.Errorf("symbol MathData not found in dictionary")
	} else if symData.ModuleName != "math_sub" || symData.Kind != mob.SymbolData {
		t.Errorf("symbol MathData info mismatch: %+v", symData)
	}

	_, foundNonExistent := arc.FindModuleForSymbol("UnknownFunc")
	if foundNonExistent {
		t.Errorf("expected UnknownFunc not to be found")
	}

	// 6. Extrair objeto e comparar
	extObj1, err := arc.ExtractObject("math_add")
	if err != nil {
		t.Fatalf("ExtractObject math_add failed: %v", err)
	}
	if len(extObj1.Symbols) != 1 || extObj1.Symbols[0].Name != "Add16" {
		t.Errorf("extracted object 1 symbols mismatch: %+v", extObj1.Symbols)
	}

	extObj2, err := arc.ExtractObject("math_sub")
	if err != nil {
		t.Fatalf("ExtractObject math_sub failed: %v", err)
	}
	if len(extObj2.Symbols) != 2 {
		t.Errorf("extracted object 2 symbols count mismatch: %+v", extObj2.Symbols)
	}
}

func TestHakoDuplicateSymbolRejection(t *testing.T) {
	tempDir := t.TempDir()

	mod1 := mob.NewObjectFile()
	seg1 := mod1.AddSegment(mob.SegmentCode, 0, []byte{0xC9}, 0)
	mod1.AddSymbol("SameFunc", mob.SymbolPublic, mob.SymbolProc, seg1, 0)
	mob1Path := filepath.Join(tempDir, "mod1.mob")
	mob.SaveToFile(mob1Path, mod1)

	mod2 := mob.NewObjectFile()
	seg2 := mod2.AddSegment(mob.SegmentCode, 0, []byte{0xC9}, 0)
	mod2.AddSymbol("SameFunc", mob.SymbolPublic, mob.SymbolProc, seg2, 0)
	mob2Path := filepath.Join(tempDir, "mod2.mob")
	mob.SaveToFile(mob2Path, mod2)

	libPath := filepath.Join(tempDir, "conflict.hlib")
	err := Pack(libPath, mob1Path, mob2Path)
	if err == nil {
		t.Errorf("expected error when packing modules with duplicate public symbols, got nil")
	}
}
