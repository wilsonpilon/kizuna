package hako

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/wilsonpilon/kizuna/pkg/mob"
)

// Pack empacota múltiplos arquivos .MOB em um único arquivo de biblioteca .HLIB
func Pack(outputFile string, mobFiles ...string) error {
	if len(mobFiles) == 0 {
		return fmt.Errorf("nenhum arquivo .mob fornecido para empacotamento")
	}

	type modData struct {
		name    string
		raw     []byte
		obj     *mob.ObjectFile
		pubSyms []*mob.Symbol
	}

	var modules []modData
	seenSymbols := make(map[string]string) // symName -> moduleName

	for _, f := range mobFiles {
		raw, err := os.ReadFile(f)
		if err != nil {
			return fmt.Errorf("falha ao ler %s: %w", f, err)
		}

		obj, err := mob.LoadFromFile(f)
		if err != nil {
			return fmt.Errorf("falha ao interpretar %s como .mob: %w", f, err)
		}

		modName := strings.TrimSuffix(filepath.Base(f), filepath.Ext(f))

		var pubSyms []*mob.Symbol
		for _, s := range obj.Symbols {
			if s.Class == mob.SymbolPublic {
				if existingMod, exists := seenSymbols[s.Name]; exists {
					return fmt.Errorf("símbolo público duplicado '%s' encontrado em '%s' e '%s'",
						s.Name, existingMod, modName)
				}
				seenSymbols[s.Name] = modName
				pubSyms = append(pubSyms, s)
			}
		}

		modules = append(modules, modData{
			name:    modName,
			raw:     raw,
			obj:     obj,
			pubSyms: pubSyms,
		})
	}

	var buf bytes.Buffer

	// 1. Escrever Cabeçalho provisório (14 bytes)
	buf.WriteString(MagicHLIB)                   // 4 bytes
	buf.WriteByte(CurrentVersion)               // 1 byte
	buf.WriteByte(0)                            // 1 byte (reservado)
	binary.Write(&buf, binary.LittleEndian, uint16(len(modules))) // 2 bytes
	binary.Write(&buf, binary.LittleEndian, uint32(0))           // 4 bytes (offset do dicionário - patch depois)
	binary.Write(&buf, binary.LittleEndian, uint16(len(seenSymbols))) // 2 bytes

	// 2. Escrever Tabela de Módulos
	type modOffsetPatch struct {
		offsetPos int
	}
	var patches []modOffsetPatch

	for _, m := range modules {
		// Nome do módulo com terminador \0
		buf.WriteString(m.name)
		buf.WriteByte(0)

		// Tamanho dos dados
		binary.Write(&buf, binary.LittleEndian, uint32(len(m.raw)))

		// Placeholder para o offset dos dados (será patcheado)
		patchPos := buf.Len()
		binary.Write(&buf, binary.LittleEndian, uint32(0))
		patches = append(patches, modOffsetPatch{offsetPos: patchPos})

		// Total de símbolos públicos
		binary.Write(&buf, binary.LittleEndian, uint16(len(m.pubSyms)))
	}

	// 3. Escrever os dados brutos de cada módulo .MOB e patchear os offsets
	outBytes := buf.Bytes()
	for i, m := range modules {
		dataOffset := uint32(len(outBytes))
		binary.LittleEndian.PutUint32(outBytes[patches[i].offsetPos:patches[i].offsetPos+4], dataOffset)
		outBytes = append(outBytes, m.raw...)
	}

	// 4. Escrever o Dicionário Global de Símbolos
	dictOffset := uint32(len(outBytes))
	var dictBuf bytes.Buffer

	for mIdx, m := range modules {
		for _, s := range m.pubSyms {
			dictBuf.WriteString(s.Name)
			dictBuf.WriteByte(0)
			binary.Write(&dictBuf, binary.LittleEndian, uint16(mIdx))
			dictBuf.WriteByte(uint8(s.Kind))
		}
	}

	outBytes = append(outBytes, dictBuf.Bytes()...)

	// 5. Patchear o Offset do Dicionário no Cabeçalho (bytes 8..11)
	binary.LittleEndian.PutUint32(outBytes[8:12], dictOffset)

	// Gravar arquivo final
	if err := os.WriteFile(outputFile, outBytes, 0644); err != nil {
		return fmt.Errorf("falha ao gravar biblioteca %s: %w", outputFile, err)
	}

	return nil
}
