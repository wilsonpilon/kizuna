package mob

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

const (
	HeaderSize        = 13
	SymbolRecordSize  = 8
	RelocRecordSize   = 7
	SegmentHeaderSize = 4
)

// Writer serializa um ObjectFile para o formato binário .MOB
type Writer struct{}

// NewWriter cria um novo Writer para o formato .MOB
func NewWriter() *Writer {
	return &Writer{}
}

// Write serializa o ObjectFile em w.
func (w *Writer) Write(out io.Writer, obj *ObjectFile) error {
	if obj == nil {
		return fmt.Errorf("object file cannot be nil")
	}

	// 1. Montar a tabela de strings e mapear offsets
	stringTable := make([]byte, 0)
	// Byte 0 sempre será \0 para representar strings vazias
	stringTable = append(stringTable, 0)

	symbolOffsets := make([]uint16, len(obj.Symbols))
	stringCache := make(map[string]uint16)
	stringCache[""] = 0

	for i, sym := range obj.Symbols {
		if offset, ok := stringCache[sym.Name]; ok {
			symbolOffsets[i] = offset
		} else {
			offset := uint16(len(stringTable))
			symbolOffsets[i] = offset
			stringCache[sym.Name] = offset
			stringTable = append(stringTable, []byte(sym.Name)...)
			stringTable = append(stringTable, 0) // terminador nulo
		}
	}

	// 2. Calcular o offset onde a tabela de strings começará
	// Header (13 bytes)
	offset := uint32(HeaderSize)

	// Segmentos: para cada segmento, 4 bytes de header + len(Data)
	for _, seg := range obj.Segments {
		offset += uint32(SegmentHeaderSize)
		if seg.Type != SegmentBSS {
			offset += uint32(len(seg.Data))
		}
	}

	// Símbolos: len(Symbols) * 8 bytes
	offset += uint32(len(obj.Symbols) * SymbolRecordSize)

	// Relocações: len(Relocations) * 7 bytes
	offset += uint32(len(obj.Relocations) * RelocRecordSize)

	if offset > 0xFFFF {
		return fmt.Errorf("object file size exceeds 64KB limit (%d bytes)", offset)
	}
	stringTableOffset := uint16(offset)

	// 3. Escrever Header (13 bytes)
	headerBuf := make([]byte, HeaderSize)
	copy(headerBuf[0:4], []byte(MagicMOB1))
	headerBuf[4] = obj.Version
	binary.LittleEndian.PutUint16(headerBuf[5:7], uint16(len(obj.Segments)))
	binary.LittleEndian.PutUint16(headerBuf[7:9], uint16(len(obj.Symbols)))
	binary.LittleEndian.PutUint16(headerBuf[9:11], uint16(len(obj.Relocations)))
	binary.LittleEndian.PutUint16(headerBuf[11:13], stringTableOffset)

	if _, err := out.Write(headerBuf); err != nil {
		return fmt.Errorf("error writing header: %w", err)
	}

	// 4. Escrever Segmentos
	for i, seg := range obj.Segments {
		segHeader := make([]byte, SegmentHeaderSize)
		segHeader[0] = uint8(seg.Type)
		segHeader[1] = seg.Bank
		binary.LittleEndian.PutUint16(segHeader[2:4], seg.Size)

		if _, err := out.Write(segHeader); err != nil {
			return fmt.Errorf("error writing segment %d header: %w", i, err)
		}

		if seg.Type != SegmentBSS && len(seg.Data) > 0 {
			if _, err := out.Write(seg.Data); err != nil {
				return fmt.Errorf("error writing segment %d data: %w", i, err)
			}
		}
	}

	// 5. Escrever Símbolos (8 bytes cada)
	symBuf := make([]byte, SymbolRecordSize)
	for i, sym := range obj.Symbols {
		binary.LittleEndian.PutUint16(symBuf[0:2], symbolOffsets[i])
		symBuf[2] = uint8(sym.Class)
		symBuf[3] = uint8(sym.Kind)
		binary.LittleEndian.PutUint16(symBuf[4:6], sym.SegmentIndex)
		binary.LittleEndian.PutUint16(symBuf[6:8], sym.Offset)

		if _, err := out.Write(symBuf); err != nil {
			return fmt.Errorf("error writing symbol %d: %w", i, err)
		}
	}

	// 6. Escrever Relocações (7 bytes cada)
	relocBuf := make([]byte, RelocRecordSize)
	for i, reloc := range obj.Relocations {
		binary.LittleEndian.PutUint16(relocBuf[0:2], reloc.SegmentIndex)
		binary.LittleEndian.PutUint16(relocBuf[2:4], reloc.Offset)
		binary.LittleEndian.PutUint16(relocBuf[4:6], reloc.SymbolIndex)
		relocBuf[6] = uint8(reloc.Type)

		if _, err := out.Write(relocBuf); err != nil {
			return fmt.Errorf("error writing relocation %d: %w", i, err)
		}
	}

	// 7. Escrever Tabela de Strings
	if _, err := out.Write(stringTable); err != nil {
		return fmt.Errorf("error writing string table: %w", err)
	}

	return nil
}

// Encode retorna os bytes do ObjectFile serializado.
func Encode(obj *ObjectFile) ([]byte, error) {
	var buf bytes.Buffer
	w := NewWriter()
	if err := w.Write(&buf, obj); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// SaveToFile salva o arquivo .MOB no disco.
func SaveToFile(filename string, obj *ObjectFile) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	w := NewWriter()
	return w.Write(f, obj)
}
