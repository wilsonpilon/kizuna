package mob

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

// Reader é responsável por deserializar um arquivo .MOB para um ObjectFile em memória.
type Reader struct{}

// NewReader cria uma nova instância de Reader.
func NewReader() *Reader {
	return &Reader{}
}

// Read faz a leitura de um ObjectFile a partir de um io.Reader.
func (r *Reader) Read(in io.Reader) (*ObjectFile, error) {
	// Lemos todo o conteúdo em memória para facilitar indexação e validações
	data, err := io.ReadAll(in)
	if err != nil {
		return nil, fmt.Errorf("failed to read data: %w", err)
	}

	if len(data) < HeaderSize {
		return nil, fmt.Errorf("file too small (%d bytes), minimum header is %d bytes", len(data), HeaderSize)
	}

	// 1. Validar Header
	magic := string(data[0:4])
	if magic != MagicMOB1 {
		return nil, fmt.Errorf("invalid magic: expected %q, got %q", MagicMOB1, magic)
	}

	version := data[4]
	if version != CurrentVersion {
		return nil, fmt.Errorf("unsupported format version: %d (expected %d)", version, CurrentVersion)
	}

	numSegments := binary.LittleEndian.Uint16(data[5:7])
	numSymbols := binary.LittleEndian.Uint16(data[7:9])
	numRelocs := binary.LittleEndian.Uint16(data[9:11])
	stringTableOffset := binary.LittleEndian.Uint16(data[11:13])

	if int(stringTableOffset) > len(data) {
		return nil, fmt.Errorf("invalid string table offset %d (file size %d)", stringTableOffset, len(data))
	}

	stringTable := data[stringTableOffset:]

	// Função auxiliar para obter strings terminadas em \0 a partir de um offset
	getString := func(offset uint16) (string, error) {
		if int(offset) >= len(stringTable) {
			return "", fmt.Errorf("string offset %d out of bounds (string table size %d)", offset, len(stringTable))
		}
		nullPos := bytes.IndexByte(stringTable[offset:], 0)
		if nullPos == -1 {
			// Não encontrou \0 antes do fim da tabela
			return string(stringTable[offset:]), nil
		}
		return string(stringTable[offset : int(offset)+nullPos]), nil
	}

	// 2. Ler Segmentos
	cursor := HeaderSize
	segments := make([]*Segment, 0, numSegments)

	for i := 0; i < int(numSegments); i++ {
		if cursor+SegmentHeaderSize > int(stringTableOffset) {
			return nil, fmt.Errorf("truncated segment header at index %d", i)
		}

		segType := SegmentType(data[cursor])
		bank := data[cursor+1]
		segSize := binary.LittleEndian.Uint16(data[cursor+2 : cursor+4])
		cursor += SegmentHeaderSize

		var segData []byte
		if segType != SegmentBSS {
			if cursor+int(segSize) > int(stringTableOffset) {
				return nil, fmt.Errorf("truncated data for segment %d (expected %d bytes)", i, segSize)
			}
			segData = make([]byte, segSize)
			copy(segData, data[cursor:cursor+int(segSize)])
			cursor += int(segSize)
		}

		segments = append(segments, &Segment{
			Type: segType,
			Bank: bank,
			Data: segData,
			Size: segSize,
		})
	}

	// 3. Ler Símbolos
	symbols := make([]*Symbol, 0, numSymbols)
	for i := 0; i < int(numSymbols); i++ {
		if cursor+SymbolRecordSize > int(stringTableOffset) {
			return nil, fmt.Errorf("truncated symbol record at index %d", i)
		}

		nameOffset := binary.LittleEndian.Uint16(data[cursor : cursor+2])
		class := SymbolClass(data[cursor+2])
		kind := SymbolKind(data[cursor+3])
		segIdx := binary.LittleEndian.Uint16(data[cursor+4 : cursor+6])
		offset := binary.LittleEndian.Uint16(data[cursor+6 : cursor+8])
		cursor += SymbolRecordSize

		symName, err := getString(nameOffset)
		if err != nil {
			return nil, fmt.Errorf("invalid symbol name at index %d: %w", i, err)
		}

		symbols = append(symbols, &Symbol{
			Name:         symName,
			Class:        class,
			Kind:         kind,
			SegmentIndex: segIdx,
			Offset:       offset,
		})
	}

	// 4. Ler Relocações
	relocs := make([]*Relocation, 0, numRelocs)
	for i := 0; i < int(numRelocs); i++ {
		if cursor+RelocRecordSize > int(stringTableOffset) {
			return nil, fmt.Errorf("truncated relocation record at index %d", i)
		}

		segIdx := binary.LittleEndian.Uint16(data[cursor : cursor+2])
		offset := binary.LittleEndian.Uint16(data[cursor+2 : cursor+4])
		symIdx := binary.LittleEndian.Uint16(data[cursor+4 : cursor+6])
		relocType := RelocType(data[cursor+6])
		cursor += RelocRecordSize

		relocs = append(relocs, &Relocation{
			SegmentIndex: segIdx,
			Offset:       offset,
			SymbolIndex:  symIdx,
			Type:         relocType,
		})
	}

	return &ObjectFile{
		Version:     version,
		Segments:    segments,
		Symbols:     symbols,
		Relocations: relocs,
	}, nil
}

// Decode deserializa bytes para um ObjectFile.
func Decode(data []byte) (*ObjectFile, error) {
	r := NewReader()
	return r.Read(bytes.NewReader(data))
}

// LoadFromFile carrega um arquivo .MOB do disco.
func LoadFromFile(filename string) (*ObjectFile, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := NewReader()
	return r.Read(f)
}
