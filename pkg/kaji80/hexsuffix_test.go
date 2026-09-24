package kaji80

import "testing"

// Regressão: um símbolo que TERMINA em "h" e COMEÇA com letras hexadecimais
// (CHAR_IsGraph, Fetch, Data_Path, Bench...) era lido como número hexadecimal
// -- parseHex usava Sscanf("%x"), que aceita só o prefixo ("C" de "CHAR_..."),
// então o operando virava o valor 0Ch, sem erro e sem relocation.
func TestSymbolEndingInHIsNotHex(t *testing.T) {
	names := []string{"CHAR_IsGraph", "Fetch", "Data_Path", "Bench", "Dash", "Each", "Face", "Beach"}
	for _, n := range names {
		src := "MODULE X\nBANK 0\nPUBLIC S\nEXTERN " + n + "\nS:\n" +
			"    CALL " + n + "\n    JP " + n + "\n    LD HL, " + n + "\n    LD DE, " + n + "\n    DW " + n + "\n"
		obj, err := NewAssembler().Assemble(src)
		if err != nil {
			t.Fatalf("%s: %v", n, err)
		}
		if got := len(obj.Relocations); got != 5 {
			t.Errorf("%s: %d relocations, quer 5 (o símbolo virou um número?)", n, got)
		}
		for _, r := range obj.Relocations {
			if obj.Symbols[r.SymbolIndex].Name != n {
				t.Errorf("%s: relocation aponta para %q", n, obj.Symbols[r.SymbolIndex].Name)
			}
		}
	}
	// números com sufixo h continuam sendo números
	obj, err := NewAssembler().Assemble("MODULE X\nBANK 0\nPUBLIC S\nS:\n    LD HL, 0CAFEh\n    DW 0Bh\n    LD DE, 1234h\n")
	if err != nil {
		t.Fatal(err)
	}
	if len(obj.Relocations) != 0 {
		t.Errorf("literais hexadecimais geraram relocation")
	}
	want := []byte{0x21, 0xFE, 0xCA, 0x0B, 0x00, 0x11, 0x34, 0x12}
	if got := obj.Segments[0].Data; string(got) != string(want) {
		t.Errorf("bytes = % X, quer % X", got, want)
	}
}
