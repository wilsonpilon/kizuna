package hako

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"

	"github.com/wilsonpilon/kizuna/pkg/mob"
)

// Open carrega e valida um arquivo de biblioteca .HLIB
func Open(filename string) (*Archive, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("falha ao ler biblioteca %s: %w", filename, err)
	}
	return FromBytes(data)
}

// FromBytes carrega um Archive a partir de bytes em memória
func FromBytes(data []byte) (*Archive, error) {
	if len(data) < HeaderSize {
		return nil, fmt.Errorf("arquivo muito pequeno para ser um .HLIB (%d bytes)", len(data))
	}

	// 1. Validar Header
	magic := string(data[0:4])
	if magic != MagicHLIB {
		return nil, fmt.Errorf("identificador mágico inválido: esperado %q, obtido %q", MagicHLIB, magic)
	}

	ver := data[4]
	if ver != CurrentVersion {
		return nil, fmt.Errorf("versão não suportada do formato .HLIB: %d (esperada: %d)", ver, CurrentVersion)
	}

	numModules := binary.LittleEndian.Uint16(data[6:8])
	dictOffset := binary.LittleEndian.Uint32(data[8:12])
	numSymbols := binary.LittleEndian.Uint16(data[12:14])

	if int(dictOffset) > len(data) {
		return nil, fmt.Errorf("offset do dicionário fora dos limites do arquivo (offset %d, tamanho %d)", dictOffset, len(data))
	}

	arc := &Archive{
		Version:      ver,
		Modules:      make([]*ModuleEntry, 0, numModules),
		Symbols:      make(map[string]*SymbolEntry),
		rawData:      data,
		moduleLookup: make(map[string]*ModuleEntry),
	}

	// 2. Ler Tabela de Módulos
	cursor := HeaderSize
	for i := 0; i < int(numModules); i++ {
		// Ler nome terminado em \0
		nullPos := bytes.IndexByte(data[cursor:], 0)
		if nullPos == -1 {
			return nil, fmt.Errorf("nome do módulo %d não terminado por nulo", i)
		}
		modName := string(data[cursor : cursor+nullPos])
		cursor += nullPos + 1

		if cursor+10 > len(data) {
			return nil, fmt.Errorf("dados truncados na tabela de módulos para '%s'", modName)
		}

		size := binary.LittleEndian.Uint32(data[cursor : cursor+4])
		cursor += 4
		offset := binary.LittleEndian.Uint32(data[cursor : cursor+4])
		cursor += 4
		symCount := binary.LittleEndian.Uint16(data[cursor : cursor+2])
		cursor += 2

		if int(offset+size) > len(data) {
			return nil, fmt.Errorf("payload do módulo '%s' excede os limites do arquivo (offset: %d, tam: %d, total: %d)",
				modName, offset, size, len(data))
		}

		entry := &ModuleEntry{
			Index:       i,
			Name:        modName,
			Size:        size,
			Offset:      offset,
			SymbolCount: symCount,
		}
		arc.Modules = append(arc.Modules, entry)
		arc.moduleLookup[modName] = entry
	}

	// 3. Ler Dicionário Global de Símbolos
	dictCursor := int(dictOffset)
	for i := 0; i < int(numSymbols); i++ {
		if dictCursor >= len(data) {
			return nil, fmt.Errorf("dicionário de símbolos truncado no símbolo %d", i)
		}
		nullPos := bytes.IndexByte(data[dictCursor:], 0)
		if nullPos == -1 {
			return nil, fmt.Errorf("nome do símbolo %d não terminado por nulo no dicionário", i)
		}
		symName := string(data[dictCursor : dictCursor+nullPos])
		dictCursor += nullPos + 1

		if dictCursor+3 > len(data) {
			return nil, fmt.Errorf("registro truncado para o símbolo '%s'", symName)
		}

		modIdx := binary.LittleEndian.Uint16(data[dictCursor : dictCursor+2])
		dictCursor += 2
		kind := mob.SymbolKind(data[dictCursor])
		dictCursor += 1

		if int(modIdx) >= len(arc.Modules) {
			return nil, fmt.Errorf("símbolo '%s' aponta para índice de módulo inválido %d", symName, modIdx)
		}

		modName := arc.Modules[modIdx].Name
		arc.Symbols[symName] = &SymbolEntry{
			Name:        symName,
			ModuleIndex: modIdx,
			ModuleName:  modName,
			Kind:        kind,
		}
	}

	return arc, nil
}

// List retorna a lista de módulos contidos na biblioteca
func (a *Archive) List() []*ModuleEntry {
	return a.Modules
}

// GetModuleInfo retorna metadados de um módulo pelo nome
func (a *Archive) GetModuleInfo(name string) (*ModuleEntry, bool) {
	m, ok := a.moduleLookup[name]
	return m, ok
}

// FindModuleForSymbol busca qual módulo da biblioteca exporta um dado símbolo
func (a *Archive) FindModuleForSymbol(symName string) (*SymbolEntry, bool) {
	s, ok := a.Symbols[symName]
	return s, ok
}

// ExtractRaw extrai os bytes brutos do arquivo .MOB de um módulo
func (a *Archive) ExtractRaw(moduleName string) ([]byte, error) {
	m, ok := a.moduleLookup[moduleName]
	if !ok {
		return nil, fmt.Errorf("módulo '%s' não encontrado na biblioteca", moduleName)
	}
	start := int(m.Offset)
	end := start + int(m.Size)
	res := make([]byte, m.Size)
	copy(res, a.rawData[start:end])
	return res, nil
}

// ExtractObject extrai e deserializa o ObjectFile de um módulo
func (a *Archive) ExtractObject(moduleName string) (*mob.ObjectFile, error) {
	raw, err := a.ExtractRaw(moduleName)
	if err != nil {
		return nil, err
	}
	reader := mob.NewReader()
	return reader.Read(bytes.NewReader(raw))
}
