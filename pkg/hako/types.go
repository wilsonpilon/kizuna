package hako

import "github.com/wilsonpilon/kizuna/pkg/mob"

const (
	// MagicHLIB identifica o arquivo de biblioteca KIZUNA
	MagicHLIB = "HLIB"

	// CurrentVersion é a versão atual do formato .HLIB
	CurrentVersion uint8 = 1

	// HeaderSize é o tamanho fixo do cabeçalho .HLIB (14 bytes)
	// 0..3:   Magic "HLIB"
	// 4:      Versão (1)
	// 5:      Reservado (0)
	// 6..7:   Total de módulos (uint16)
	// 8..11:  Offset do Dicionário Global de Símbolos (uint32)
	// 12..13: Total de símbolos no dicionário (uint16)
	HeaderSize = 14
)

// ModuleEntry descreve um módulo .MOB contido no arquivo .HLIB
type ModuleEntry struct {
	Index       int
	Name        string
	Size        uint32
	Offset      uint32
	SymbolCount uint16
}

// SymbolEntry mapeia um símbolo exportado (PUBLIC) para o módulo que o contém
type SymbolEntry struct {
	Name        string
	ModuleIndex uint16
	ModuleName  string
	Kind        mob.SymbolKind // mob.SymbolProc ou mob.SymbolData
}

// Archive representa um arquivo .HLIB carregado em memória
type Archive struct {
	Version      uint8
	Modules      []*ModuleEntry
	Symbols      map[string]*SymbolEntry
	rawData      []byte
	moduleLookup map[string]*ModuleEntry
}
