package main

import (
	"fmt"
	"os"

	"github.com/wilsonpilon/kizuna/pkg/mob"
	"github.com/wilsonpilon/kizuna/pkg/version"
)

func main() {
	if len(os.Args) < 2 || os.Args[1] == "-h" || os.Args[1] == "--help" {
		fmt.Println(version.Banner("MOBDUMP"))
		fmt.Fprintf(os.Stderr, "Uso: mobdump <arquivo.mob> | mobdump --version\n")
		os.Exit(1)
	}

	if os.Args[1] == "--version" || os.Args[1] == "-v" {
		fmt.Println(version.Banner("MOBDUMP"))
		os.Exit(0)
	}

	filename := os.Args[1]
	obj, err := mob.LoadFromFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao carregar %s: %v\n", filename, err)
		os.Exit(1)
	}

	fmt.Printf("=== KIZUNA MOB DUMP: %s ===\n", filename)
	fmt.Printf("Formato Versão: %d\n", obj.Version)
	fmt.Printf("Total de Segmentos:    %d\n", len(obj.Segments))
	fmt.Printf("Total de Símbolos:     %d\n", len(obj.Symbols))
	fmt.Printf("Total de Relocações:   %d\n\n", len(obj.Relocations))

	fmt.Println("--- SEGMENTOS ---")
	for i, seg := range obj.Segments {
		fmt.Printf("[%2d] Tipo: %-4s | Banco: %2d | Tamanho: %5d bytes\n",
			i, seg.Type, seg.Bank, seg.Size)
	}
	fmt.Println()

	fmt.Println("--- SÍMBOLOS ---")
	for i, sym := range obj.Symbols {
		if sym.Class == mob.SymbolPublic {
			fmt.Printf("[%2d] %-6s %-4s '%s' -> Seg: %d, Offset: 0x%04X\n",
				i, sym.Class, sym.Kind, sym.Name, sym.SegmentIndex, sym.Offset)
		} else {
			fmt.Printf("[%2d] %-6s %-4s '%s'\n",
				i, sym.Class, sym.Kind, sym.Name)
		}
	}
	fmt.Println()

	fmt.Println("--- RELOCAÇÕES ---")
	for i, rel := range obj.Relocations {
		symName := fmt.Sprintf("sym[%d]", rel.SymbolIndex)
		if int(rel.SymbolIndex) < len(obj.Symbols) {
			symName = fmt.Sprintf("'%s'", obj.Symbols[rel.SymbolIndex].Name)
		}
		fmt.Printf("[%2d] Seg: %d, Offset: 0x%04X | Tipo: %-7s | Alvo: %s\n",
			i, rel.SegmentIndex, rel.Offset, rel.Type, symName)
	}
}
