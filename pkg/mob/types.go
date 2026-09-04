package mob

import "fmt"

// SegmentType define o tipo de segmento de memória.
type SegmentType uint8

const (
	SegmentCode SegmentType = 1
	SegmentData SegmentType = 2
	SegmentBSS  SegmentType = 3
)

func (t SegmentType) String() string {
	switch t {
	case SegmentCode:
		return "CODE"
	case SegmentData:
		return "DATA"
	case SegmentBSS:
		return "BSS"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", t)
	}
}

// SymbolClass define a visibilidade do símbolo (exportado ou importado).
type SymbolClass uint8

const (
	SymbolPublic SymbolClass = 1
	SymbolExtern SymbolClass = 2
)

func (c SymbolClass) String() string {
	switch c {
	case SymbolPublic:
		return "PUBLIC"
	case SymbolExtern:
		return "EXTERN"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", c)
	}
}

// SymbolKind define se o símbolo é um procedimento executável ou dados.
// É usado pelo linker para saber se chamadas cross-bank precisam de trampolim.
type SymbolKind uint8

const (
	SymbolProc SymbolKind = 1
	SymbolData SymbolKind = 2
)

func (k SymbolKind) String() string {
	switch k {
	case SymbolProc:
		return "PROC"
	case SymbolData:
		return "DATA"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", k)
	}
}

// RelocType define o tipo de relocalização a ser aplicada pelo linker.
type RelocType uint8

const (
	RelocAbs16   RelocType = 1 // Endereço absoluto de 16 bits (ex: CALL, JP, LD HL, addr)
	RelocRel8    RelocType = 2 // Deslocamento relativo de 8 bits (ex: JR, DJNZ)
	RelocBankNum RelocType = 3 // Número do banco de memória (1 byte)
)

func (r RelocType) String() string {
	switch r {
	case RelocAbs16:
		return "ABS16"
	case RelocRel8:
		return "REL8"
	case RelocBankNum:
		return "BANKNUM"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", r)
	}
}

// Magic e Versão atual da especificação .MOB
const (
	MagicMOB1      = "MOB1"
	CurrentVersion = 1
)

// Segment representa um segmento no arquivo objeto.
type Segment struct {
	Type SegmentType // CODE, DATA, BSS
	Bank uint8       // 0 = área comum (fixa), 1..N = banco paginável
	Data []byte      // Bytes brutos (CODE/DATA); vazio para BSS
	Size uint16      // Tamanho total do segmento (para BSS determina o espaço reservado)
}

// Symbol representa uma entrada na tabela de símbolos.
type Symbol struct {
	Name         string
	Class        SymbolClass // PUBLIC ou EXTERN
	Kind         SymbolKind  // PROC ou DATA
	SegmentIndex uint16      // Índice do segmento onde o símbolo está definido (apenas se PUBLIC)
	Offset       uint16      // Deslocamento dentro do segmento (apenas se PUBLIC)
}

// Relocation representa um ponto onde um endereço/banco precisa ser resolvido na linkagem.
type Relocation struct {
	SegmentIndex uint16    // Segmento onde a relocalização deve ser aplicada
	Offset       uint16    // Deslocamento dentro do segmento onde o valor será escrito
	SymbolIndex  uint16    // Índice do símbolo alvo na tabela de símbolos
	Type         RelocType // ABS16, REL8 ou BANKNUM
}

// ObjectFile é a representação em memória de um arquivo .MOB completo.
type ObjectFile struct {
	Version     uint8
	Segments    []*Segment
	Symbols     []*Symbol
	Relocations []*Relocation
}

// NewObjectFile cria uma nova instância vazia de ObjectFile.
func NewObjectFile() *ObjectFile {
	return &ObjectFile{
		Version:     CurrentVersion,
		Segments:    make([]*Segment, 0),
		Symbols:     make([]*Symbol, 0),
		Relocations: make([]*Relocation, 0),
	}
}

// AddSegment adiciona um segmento ao arquivo objeto.
func (o *ObjectFile) AddSegment(segType SegmentType, bank uint8, data []byte, bssSize uint16) uint16 {
	size := uint16(len(data))
	if segType == SegmentBSS {
		size = bssSize
		data = nil
	}
	seg := &Segment{
		Type: segType,
		Bank: bank,
		Data: data,
		Size: size,
	}
	o.Segments = append(o.Segments, seg)
	return uint16(len(o.Segments) - 1)
}

// AddSymbol adiciona um símbolo à tabela de símbolos.
func (o *ObjectFile) AddSymbol(name string, class SymbolClass, kind SymbolKind, segIndex uint16, offset uint16) uint16 {
	sym := &Symbol{
		Name:         name,
		Class:        class,
		Kind:         kind,
		SegmentIndex: segIndex,
		Offset:       offset,
	}
	o.Symbols = append(o.Symbols, sym)
	return uint16(len(o.Symbols) - 1)
}

// AddRelocation adiciona uma relocalização.
func (o *ObjectFile) AddRelocation(segIndex uint16, offset uint16, symIndex uint16, relocType RelocType) {
	reloc := &Relocation{
		SegmentIndex: segIndex,
		Offset:       offset,
		SymbolIndex:  symIndex,
		Type:         relocType,
	}
	o.Relocations = append(o.Relocations, reloc)
}
