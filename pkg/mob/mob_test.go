package mob

import (
	"bytes"
	"path/filepath"
	"reflect"
	"testing"
)

func TestMOBRoundTrip(t *testing.T) {
	// Criar um ObjectFile hipotético simulando um módulo Z80 real
	obj := NewObjectFile()

	// Segmento 0: CODE na área comum (Banco 0)
	// Código: LD A, 2; CALL Setup; RET
	codeCommon := []byte{0x3E, 0x02, 0xCD, 0x00, 0x00, 0xC9}
	seg0 := obj.AddSegment(SegmentCode, 0, codeCommon, 0)

	// Segmento 1: CODE no Banco 1 (Janela comutável)
	// Código: PUSH AF; LD A, 15; OUT (0x99), A; POP AF; RET
	codeBank1 := []byte{0xF5, 0x3E, 0x0F, 0xD3, 0x99, 0xF1, 0xC9}
	seg1 := obj.AddSegment(SegmentCode, 1, codeBank1, 0)

	// Segmento 2: DATA na área comum (Banco 0)
	// String com length prefix: 12, "Hello, Kizuna"
	dataMsg := append([]byte{13}, []byte("Hello, Kizuna")...)
	seg2 := obj.AddSegment(SegmentData, 0, dataMsg, 0)

	// Segmento 3: BSS no Banco 2 (Reserva 256 bytes)
	seg3 := obj.AddSegment(SegmentBSS, 2, nil, 256)

	// Símbolos:
	// Public PROC no seg 0, offset 0 ("_start")
	sym0 := obj.AddSymbol("_start", SymbolPublic, SymbolProc, seg0, 0x0000)
	// Public PROC no seg 1, offset 0 ("SetupScreen")
	sym1 := obj.AddSymbol("SetupScreen", SymbolPublic, SymbolProc, seg1, 0x0000)
	// Public DATA no seg 2, offset 0 ("MsgText")
	sym2 := obj.AddSymbol("MsgText", SymbolPublic, SymbolData, seg2, 0x0000)
	// Extern PROC ("BDOS_PRINT")
	sym3 := obj.AddSymbol("BDOS_PRINT", SymbolExtern, SymbolProc, 0, 0)
	// Public BSS DATA no seg 3, offset 0 ("Buffer")
	sym4 := obj.AddSymbol("Buffer", SymbolPublic, SymbolData, seg3, 0x0000)

	// Relocações:
	// No seg0, offset 3 (argumento do CALL): precisa resolver SetupScreen (sym1) como ABS16
	obj.AddRelocation(seg0, 3, sym1, RelocAbs16)
	// No seg0, offset 0 (exemplo): precisa do número do banco de SetupScreen
	obj.AddRelocation(seg0, 1, sym1, RelocBankNum)

	_ = sym0
	_ = sym2
	_ = sym3
	_ = sym4

	// Serializar em memória
	encoded, err := Encode(obj)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	// Validar tamanho mínimo e cabeçalho bruto
	if len(encoded) < HeaderSize {
		t.Fatalf("Encoded buffer too small: %d bytes", len(encoded))
	}
	if string(encoded[0:4]) != "MOB1" {
		t.Fatalf("Expected magic MOB1, got %s", string(encoded[0:4]))
	}
	if encoded[4] != CurrentVersion {
		t.Fatalf("Expected version %d, got %d", CurrentVersion, encoded[4])
	}

	// Deserializar
	decoded, err := Decode(encoded)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	// Validar campos
	if decoded.Version != obj.Version {
		t.Errorf("Version mismatch: got %d, want %d", decoded.Version, obj.Version)
	}

	// Validar segmentos
	if len(decoded.Segments) != len(obj.Segments) {
		t.Fatalf("Segment count mismatch: got %d, want %d", len(decoded.Segments), len(obj.Segments))
	}
	for i := range obj.Segments {
		orig := obj.Segments[i]
		got := decoded.Segments[i]
		if got.Type != orig.Type {
			t.Errorf("Segment %d Type mismatch: got %v, want %v", i, got.Type, orig.Type)
		}
		if got.Bank != orig.Bank {
			t.Errorf("Segment %d Bank mismatch: got %d, want %d", i, got.Bank, orig.Bank)
		}
		if got.Size != orig.Size {
			t.Errorf("Segment %d Size mismatch: got %d, want %d", i, got.Size, orig.Size)
		}
		if !bytes.Equal(got.Data, orig.Data) {
			t.Errorf("Segment %d Data mismatch: got %x, want %x", i, got.Data, orig.Data)
		}
	}

	// Validar símbolos
	if len(decoded.Symbols) != len(obj.Symbols) {
		t.Fatalf("Symbols count mismatch: got %d, want %d", len(decoded.Symbols), len(obj.Symbols))
	}
	for i := range obj.Symbols {
		orig := obj.Symbols[i]
		got := decoded.Symbols[i]
		if !reflect.DeepEqual(orig, got) {
			t.Errorf("Symbol %d mismatch: got %+v, want %+v", i, got, orig)
		}
	}

	// Validar relocações
	if len(decoded.Relocations) != len(obj.Relocations) {
		t.Fatalf("Relocations count mismatch: got %d, want %d", len(decoded.Relocations), len(obj.Relocations))
	}
	for i := range obj.Relocations {
		orig := obj.Relocations[i]
		got := decoded.Relocations[i]
		if !reflect.DeepEqual(orig, got) {
			t.Errorf("Relocation %d mismatch: got %+v, want %+v", i, got, orig)
		}
	}
}

func TestFileSaveAndLoad(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.mob")

	obj := NewObjectFile()
	seg := obj.AddSegment(SegmentCode, 0, []byte{0xC9}, 0) // RET
	obj.AddSymbol("Return", SymbolPublic, SymbolProc, seg, 0)

	if err := SaveToFile(filePath, obj); err != nil {
		t.Fatalf("SaveToFile failed: %v", err)
	}

	loaded, err := LoadFromFile(filePath)
	if err != nil {
		t.Fatalf("LoadFromFile failed: %v", err)
	}

	if len(loaded.Segments) != 1 || len(loaded.Symbols) != 1 {
		t.Fatalf("Loaded file structure unexpected: segments=%d, symbols=%d", len(loaded.Segments), len(loaded.Symbols))
	}
	if loaded.Symbols[0].Name != "Return" {
		t.Errorf("Expected symbol name 'Return', got %s", loaded.Symbols[0].Name)
	}
}

func TestCorruptedInput(t *testing.T) {
	// Teste com buffer muito curto
	_, err := Decode([]byte{1, 2, 3})
	if err == nil {
		t.Error("Expected error on truncated buffer, got nil")
	}

	// Teste com Magic errado
	badMagic := make([]byte, HeaderSize)
	copy(badMagic[0:4], "XYZ1")
	badMagic[4] = CurrentVersion
	_, err = Decode(badMagic)
	if err == nil {
		t.Error("Expected error on invalid magic, got nil")
	}

	// Teste com Versão errada
	badVer := make([]byte, HeaderSize)
	copy(badVer[0:4], "MOB1")
	badVer[4] = 99 // versão desconhecida
	_, err = Decode(badVer)
	if err == nil {
		t.Error("Expected error on unsupported version, got nil")
	}
}

func TestStringPoolDeduplication(t *testing.T) {
	obj := NewObjectFile()
	seg := obj.AddSegment(SegmentCode, 0, []byte{0x00}, 0)
	// Dois símbolos com o mesmo nome
	obj.AddSymbol("DuplicatedName", SymbolPublic, SymbolProc, seg, 0)
	obj.AddSymbol("DuplicatedName", SymbolExtern, SymbolProc, 0, 0)

	data, err := Encode(obj)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	// Como há deduplicação no pool de strings, a string só deve aparecer uma vez na tabela de strings
	occurrences := bytes.Count(data, []byte("DuplicatedName"))
	if occurrences != 1 {
		t.Errorf("Expected exactly 1 instance of 'DuplicatedName' in binary pool, got %d", occurrences)
	}

	decoded, err := Decode(data)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	if decoded.Symbols[0].Name != "DuplicatedName" || decoded.Symbols[1].Name != "DuplicatedName" {
		t.Errorf("Symbols did not preserve name after string pool deduplication")
	}
}
