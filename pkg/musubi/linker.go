package musubi

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/wilsonpilon/kizuna/pkg/hako"
	"github.com/wilsonpilon/kizuna/pkg/mob"
)

// Porta padrão da Memory Mapper para a Página 2 (0x8000..0xBFFF) no hardware MSX
const DefaultMapperPortPage2 = 0xFE

// Endereço base da janela comutável (Página 2)
const BankWindowBaseAddress = 0x8000

// Tamanho máximo de uma página de 16KB da Memory Mapper
const BankWindowSize = 0x4000 // 16.384 bytes (16KB)

// LinkerConfig contém os parâmetros de configuração da linkagem.
type LinkerConfig struct {
	BaseAddress    uint16 // Endereço de início da área comum (padrão MSX-DOS .COM: 0x0100)
	EntryPoint     string // Nome do símbolo de entrada (ex: "Start", "MAIN")
	MapFile        string // Caminho opcional para exportar relatório de mapa de memória
	MapperPortPage2 uint8  // Porta I/O da Memory Mapper para a Página 2 (padrão: 0xFE)
	Verbose        bool   // Modo detalhado
}

// DefaultConfig retorna a configuração padrão para executáveis MSX-DOS 2 .COM
func DefaultConfig() LinkerConfig {
	return LinkerConfig{
		BaseAddress:    0x0100, // TPA do MSX-DOS
		EntryPoint:     "Start",
		MapperPortPage2: DefaultMapperPortPage2,
		Verbose:        false,
	}
}

// PlacedSegment representa um segmento posicionado na memória com endereço final.
type PlacedSegment struct {
	ModuleIndex  int
	SegmentIndex int
	Type         mob.SegmentType
	Bank         uint8
	BaseAddr     uint16
	Size         uint16
	Data         []byte
}

// ResolvedSymbol representa um símbolo com seu endereço absoluto de memória e seu banco.
type ResolvedSymbol struct {
	Name     string
	Address  uint16
	Bank     uint8
	Kind     mob.SymbolKind
	Segment  *PlacedSegment
	IsPublic bool
}

// Trampoline representa um trampolim de troca de banco gerado na Área Comum.
type Trampoline struct {
	SymbolName    string
	TargetBank    uint8
	TargetAddress uint16
	Address       uint16 // Endereço do trampolim na Área Comum
	Code          []byte // Opcodes Z80 do trampolim
}

// BankPayload armazena os dados brutos de um banco paginável (1..N).
type BankPayload struct {
	Bank uint8
	Data []byte
}

// LinkResult contém as informações do executável gerado.
type LinkResult struct {
	Binary       []byte
	EntryPoint   uint16
	BaseAddress  uint16
	EndAddress   uint16
	TotalSize    int
	IsMultiBank  bool
	BanksUsed    []uint8
	Symbols      map[string]*ResolvedSymbol
	Segments     []*PlacedSegment
	Trampolines  map[string]*Trampoline
	BankPayloads []*BankPayload
}

type segRef struct {
	modIdx int
	segIdx int
	seg    *mob.Segment
}

func orderSegments(segs []segRef) []segRef {
	var code, data, bss []segRef
	for _, s := range segs {
		switch s.seg.Type {
		case mob.SegmentCode:
			code = append(code, s)
		case mob.SegmentData:
			data = append(data, s)
		case mob.SegmentBSS:
			bss = append(bss, s)
		default:
			code = append(code, s)
		}
	}
	return append(append(code, data...), bss...)
}

// Linker é o responsável por ligar objetos .MOB e gerar o binário .COM
type Linker struct {
	config LinkerConfig
}

// NewLinker cria uma nova instância do Linker.
func NewLinker(cfg LinkerConfig) *Linker {
	if cfg.BaseAddress == 0 {
		cfg.BaseAddress = 0x0100
	}
	if cfg.MapperPortPage2 == 0 {
		cfg.MapperPortPage2 = DefaultMapperPortPage2
	}
	return &Linker{config: cfg}
}

// Link realiza a linkagem completa (monobanco ou multi-banco com trampolins).
func (l *Linker) Link(objects ...*mob.ObjectFile) (*LinkResult, error) {
	return l.LinkWithLibraries(objects, nil)
}

// LinkWithLibraries realiza a linkagem resolvendo dependências a partir de bibliotecas .HLIB.
// Módulos não utilizados nas bibliotecas são ignorados (Smart-Linking / Dead-Code Elimination).
func (l *Linker) LinkWithLibraries(objects []*mob.ObjectFile, archives []*hako.Archive) (*LinkResult, error) {
	activeObjects := make([]*mob.ObjectFile, len(objects))
	copy(activeObjects, objects)

	loadedModules := make(map[string]bool)

	// Resolução transitiva a partir dos arquivos .HLIB
	for len(archives) > 0 {
		defined := make(map[string]bool)
		for _, obj := range activeObjects {
			for _, sym := range obj.Symbols {
				if sym.Class == mob.SymbolPublic {
					defined[sym.Name] = true
				}
			}
		}

		needed := make(map[string]bool)
		if l.config.EntryPoint != "" && !defined[l.config.EntryPoint] {
			needed[l.config.EntryPoint] = true
		}

		for _, obj := range activeObjects {
			for _, reloc := range obj.Relocations {
				if int(reloc.SymbolIndex) < len(obj.Symbols) {
					sym := obj.Symbols[reloc.SymbolIndex]
					if !defined[sym.Name] {
						needed[sym.Name] = true
					}
				}
			}
		}

		if len(needed) == 0 {
			break
		}

		addedAny := false
		for symName := range needed {
			for arcIdx, arc := range archives {
				if symEntry, found := arc.FindModuleForSymbol(symName); found {
					key := fmt.Sprintf("%d:%s", arcIdx, symEntry.ModuleName)
					if !loadedModules[key] {
						loadedModules[key] = true
						obj, err := arc.ExtractObject(symEntry.ModuleName)
						if err != nil {
							return nil, fmt.Errorf("falha ao extrair módulo '%s' da biblioteca para '%s': %w", symEntry.ModuleName, symName, err)
						}
						activeObjects = append(activeObjects, obj)
						addedAny = true
						break
					}
				}
			}
			if addedAny {
				break
			}
		}

		if !addedAny {
			break // Nenhuma biblioteca pôde resolver os símbolos restantes
		}
	}

	if len(activeObjects) == 0 {
		return nil, fmt.Errorf("nenhum objeto .mob fornecido para linkagem")
	}

	return l.linkObjects(activeObjects)
}

func (l *Linker) linkObjects(objects []*mob.ObjectFile) (*LinkResult, error) {
	// 1. Descobrir todos os bancos utilizados
	banksUsedMap := make(map[uint8]bool)
	bankSegments := make(map[uint8][]segRef)

	for mIdx, obj := range objects {
		for sIdx, seg := range obj.Segments {
			banksUsedMap[seg.Bank] = true
			ref := segRef{modIdx: mIdx, segIdx: sIdx, seg: seg}
			bankSegments[seg.Bank] = append(bankSegments[seg.Bank], ref)
		}
	}

	var banksList []int
	for b := range banksUsedMap {
		banksList = append(banksList, int(b))
	}
	sort.Ints(banksList)

	isMultiBank := len(banksList) > 1 || (len(banksList) == 1 && banksList[0] != 0)

	// Calcular tarefas de cópia de bancos se for multi-banco
	type copyTask struct {
		bank uint8
		size uint16
	}
	var copyTasks []copyTask
	var bootstrapSize uint16
	if isMultiBank {
		for _, bInt := range banksList {
			b := uint8(bInt)
			if b == 0 {
				continue
			}
			var bSize uint16
			for _, ref := range bankSegments[b] {
				if ref.seg.Type != mob.SegmentBSS {
					bSize += ref.seg.Size
				}
			}
			if bSize > 0 {
				copyTasks = append(copyTasks, copyTask{bank: b, size: bSize})
			}
		}
		// Alinhamento Slot 0xA8 (18 bytes) + EXTBIO mapper detection (32 bytes) + Print "[L]" (21 bytes) + 16 bytes por banco + 8 bytes epílogo
		bootstrapSize = uint16(18 + 32 + 21 + len(copyTasks)*16 + 8)
	}

	// 2. Posicionar segmentos na memória
	// Banco 0: começa em BaseAddress + bootstrapSize (se multi-banco)
	// Bancos 1..N: começam na janela da Página 2 (0x8000)
	placedSegments := make([]*PlacedSegment, 0)
	segLookup := make(map[string]*PlacedSegment)

	// Se for multi-banco, alocamos o Bootstrap Loader em BaseAddress (0x0100)
	var bootstrapSeg *PlacedSegment
	if isMultiBank {
		bootstrapSeg = &PlacedSegment{
			ModuleIndex:  -2, // Linker Bootstrap
			SegmentIndex: -2,
			Type:         mob.SegmentCode,
			Bank:         0,
			BaseAddr:     l.config.BaseAddress,
			Size:         bootstrapSize,
			Data:         make([]byte, bootstrapSize),
		}
		placedSegments = append(placedSegments, bootstrapSeg)
	}

	// A) Posicionar Banco 0 (Área Comum)
	commonCurrentAddr := l.config.BaseAddress + bootstrapSize
	if segs, ok := bankSegments[0]; ok {
		// Ordenar CODE, DATA, BSS
		orderedCommon := orderSegments(segs)
		for _, ref := range orderedCommon {
			segData := copyData(ref.seg)
			size := ref.seg.Size
			if uint32(commonCurrentAddr)+uint32(size) > 0x8000 {
				return nil, fmt.Errorf("área comum (Banco 0) invadiu a janela comutável da Página 2 (0x8000)")
			}

			ps := &PlacedSegment{
				ModuleIndex:  ref.modIdx,
				SegmentIndex: ref.segIdx,
				Type:         ref.seg.Type,
				Bank:         0,
				BaseAddr:     commonCurrentAddr,
				Size:         size,
				Data:         segData,
			}
			key := fmt.Sprintf("%d:%d", ref.modIdx, ref.segIdx)
			segLookup[key] = ps
			placedSegments = append(placedSegments, ps)
			commonCurrentAddr += size
		}
	}

	// B) Posicionar Bancos 1..N (Pagináveis)
	for _, bInt := range banksList {
		b := uint8(bInt)
		if b == 0 {
			continue
		}
		segs := bankSegments[b]
		orderedBank := orderSegments(segs)
		bankCurrentAddr := uint16(BankWindowBaseAddress)

		for _, ref := range orderedBank {
			segData := copyData(ref.seg)
			size := ref.seg.Size
			if uint32(bankCurrentAddr)+uint32(size) > uint32(BankWindowBaseAddress+BankWindowSize) {
				return nil, fmt.Errorf("banco paginável %d excedeu o limite de 16KB (tamanho acumulado: %d bytes)",
					b, (bankCurrentAddr+size)-BankWindowBaseAddress)
			}

			ps := &PlacedSegment{
				ModuleIndex:  ref.modIdx,
				SegmentIndex: ref.segIdx,
				Type:         ref.seg.Type,
				Bank:         b,
				BaseAddr:     bankCurrentAddr,
				Size:         size,
				Data:         segData,
			}
			key := fmt.Sprintf("%d:%d", ref.modIdx, ref.segIdx)
			segLookup[key] = ps
			placedSegments = append(placedSegments, ps)
			bankCurrentAddr += size
		}
	}

	// 3. Resolver Tabela Global de Símbolos
	globalSymbols := make(map[string]*ResolvedSymbol)
	for mIdx, obj := range objects {
		for _, sym := range obj.Symbols {
			if sym.Class == mob.SymbolPublic {
				key := fmt.Sprintf("%d:%d", mIdx, sym.SegmentIndex)
				ps, ok := segLookup[key]
				if !ok {
					return nil, fmt.Errorf("módulo %d: segmento %d não encontrado para símbolo '%s'", mIdx, sym.SegmentIndex, sym.Name)
				}

				absAddr := ps.BaseAddr + sym.Offset
				if existing, duplicate := globalSymbols[sym.Name]; duplicate {
					return nil, fmt.Errorf("símbolo duplicado '%s' (já definido no Banco %d em 0x%04X, redefinido no Banco %d em 0x%04X)",
						sym.Name, existing.Bank, existing.Address, ps.Bank, absAddr)
				}

				globalSymbols[sym.Name] = &ResolvedSymbol{
					Name:     sym.Name,
					Address:  absAddr,
					Bank:     ps.Bank,
					Kind:     sym.Kind,
					Segment:  ps,
					IsPublic: true,
				}
			}
		}
	}

	// 4. Identificar necessidades de trampolim e alocar dispatcher + trampolins na Área Comum
	trampolines := make(map[string]*Trampoline)
	var putP2Addr uint16
	var getP2Addr uint16

	if isMultiBank {
		dispatcherAddr := commonCurrentAddr
		putP2Addr = dispatcherAddr
		getP2Addr = dispatcherAddr + 3
		dispatcherCode := l.buildDispatcherCode(dispatcherAddr)
		dispSize := uint16(len(dispatcherCode))

		if uint32(commonCurrentAddr)+uint32(dispSize) > 0x8000 {
			return nil, fmt.Errorf("área comum estourou limite de 0x8000 ao alocar dispatcher")
		}

		trampolineAddrCursor := commonCurrentAddr + dispSize

		for mIdx, obj := range objects {
			for _, reloc := range obj.Relocations {
				key := fmt.Sprintf("%d:%d", mIdx, reloc.SegmentIndex)
				ps := segLookup[key]
				sym := obj.Symbols[reloc.SymbolIndex]
				resolved, found := globalSymbols[sym.Name]
				if !found {
					return nil, fmt.Errorf("referência indefinida: símbolo '%s' não encontrado", sym.Name)
				}

				// Se a chamada for cross-bank e for um procedimento (PROC)
				if reloc.Type == mob.RelocAbs16 && ps.Bank != resolved.Bank && resolved.Kind == mob.SymbolProc {
					if _, exists := trampolines[resolved.Name]; !exists {
						// Criar novo trampolim na Área Comum
						tCode := l.buildTrampolineCode(resolved.Bank, resolved.Address, putP2Addr, getP2Addr)
						tSize := uint16(len(tCode))

						if uint32(trampolineAddrCursor)+uint32(tSize) > 0x8000 {
							return nil, fmt.Errorf("área comum estourou limite de 0x8000 ao alocar trampolim para '%s'", resolved.Name)
						}

						trampolines[resolved.Name] = &Trampoline{
							SymbolName:    resolved.Name,
							TargetBank:    resolved.Bank,
							TargetAddress: resolved.Address,
							Address:       trampolineAddrCursor,
							Code:          tCode,
						}
						trampolineAddrCursor += tSize
					}
				}
			}
		}

		// Segmento na Área Comum contendo Dispatcher + Trampolins
		var trampBytes []byte
		trampBytes = append(trampBytes, dispatcherCode...)

		// Ordenar trampolins por endereço para determinismo
		var sortedTramps []*Trampoline
		for _, t := range trampolines {
			sortedTramps = append(sortedTramps, t)
		}
		sort.Slice(sortedTramps, func(i, j int) bool {
			return sortedTramps[i].Address < sortedTramps[j].Address
		})

		for _, t := range sortedTramps {
			trampBytes = append(trampBytes, t.Code...)
		}

		trampolineSegment := &PlacedSegment{
			ModuleIndex:  -1, // Gerado pelo linker
			SegmentIndex: -1,
			Type:         mob.SegmentCode,
			Bank:         0,
			BaseAddr:     commonCurrentAddr,
			Size:         uint16(len(trampBytes)),
			Data:         trampBytes,
		}
		placedSegments = append(placedSegments, trampolineSegment)
		commonCurrentAddr = trampolineAddrCursor
	}

	// 5. Aplicar Relocações
	for mIdx, obj := range objects {
		for _, reloc := range obj.Relocations {
			key := fmt.Sprintf("%d:%d", mIdx, reloc.SegmentIndex)
			ps := segLookup[key]
			sym := obj.Symbols[reloc.SymbolIndex]
			resolved, found := globalSymbols[sym.Name]
			if !found {
				return nil, fmt.Errorf("referência indefinida: símbolo '%s' não encontrado", sym.Name)
			}
			offset := int(reloc.Offset)

			switch reloc.Type {
			case mob.RelocAbs16:
				if offset+2 > len(ps.Data) {
					return nil, fmt.Errorf("relocação ABS16 fora dos limites no segmento %d", reloc.SegmentIndex)
				}
				addend := binary.LittleEndian.Uint16(ps.Data[offset : offset+2])

				// Decisão de destino: se cross-bank para PROC, usar endereço do Trampolim!
				targetAddr := resolved.Address
				if ps.Bank != resolved.Bank && resolved.Kind == mob.SymbolProc {
					if tr, ok := trampolines[resolved.Name]; ok {
						targetAddr = tr.Address
					}
				}

				finalAddr := targetAddr + addend
				binary.LittleEndian.PutUint16(ps.Data[offset:offset+2], finalAddr)

			case mob.RelocRel8:
				if offset+1 > len(ps.Data) {
					return nil, fmt.Errorf("relocação REL8 fora dos limites no segmento %d", reloc.SegmentIndex)
				}
				if ps.Bank != resolved.Bank {
					return nil, fmt.Errorf("salto relativo REL8 (JR/DJNZ) não permitido entre bancos diferentes (de Banco %d para Banco %d no símbolo '%s')",
						ps.Bank, resolved.Bank, sym.Name)
				}
				pc := int(ps.BaseAddr) + offset + 1
				disp := int(resolved.Address) - pc
				if disp < -128 || disp > 127 {
					return nil, fmt.Errorf("salto relativo REL8 fora de alcance para '%s' (deslocamento: %d bytes)", sym.Name, disp)
				}
				ps.Data[offset] = uint8(int8(disp))

			case mob.RelocBankNum:
				if offset+1 > len(ps.Data) {
					return nil, fmt.Errorf("relocação BANKNUM fora dos limites no segmento %d", reloc.SegmentIndex)
				}
				ps.Data[offset] = resolved.Bank
			}
		}
	}

	// 6. Montagem dos Payloads de Banco
	bankPayloadsMap := make(map[uint8][]byte)
	for _, bInt := range banksList {
		b := uint8(bInt)
		if b == 0 {
			continue
		}
		var bData []byte
		for _, ps := range placedSegments {
			if ps.Bank == b && ps.Type != mob.SegmentBSS && len(ps.Data) > 0 {
				bData = append(bData, ps.Data...)
			}
		}
		bankPayloadsMap[b] = bData
	}


	var bankPayloads []*BankPayload
	for _, bInt := range banksList {
		b := uint8(bInt)
		if b == 0 {
			continue
		}
		bankPayloads = append(bankPayloads, &BankPayload{
			Bank: b,
			Data: bankPayloadsMap[b],
		})
	}

	// Se for multi-banco, preencher o bootstrapSeg com o código do loader
	entryAddress := l.config.BaseAddress
	if l.config.EntryPoint != "" {
		if entrySym, ok := globalSymbols[l.config.EntryPoint]; ok {
			entryAddress = entrySym.Address
		}
	}

	if isMultiBank && bootstrapSeg != nil {
		firstPayloadAddr := commonCurrentAddr
		bootCode := l.buildBootstrapCode(bankPayloads, firstPayloadAddr, entryAddress, putP2Addr, getP2Addr)
		copy(bootstrapSeg.Data, bootCode)
	}

	// Coletar dados brutos da Área Comum (Banco 0, incluindo bootstrap se houver)
	var commonBinary []byte
	for _, ps := range placedSegments {
		if ps.Bank == 0 && ps.Type != mob.SegmentBSS && len(ps.Data) > 0 {
			commonBinary = append(commonBinary, ps.Data...)
		}
	}

	// Montar binário final: Área Comum + Payloads de Bancos Pagináveis
	var finalBinary []byte
	finalBinary = append(finalBinary, commonBinary...)
	if isMultiBank {
		for _, b := range bankPayloads {
			finalBinary = append(finalBinary, b.Data...)
		}
	}

	var byteBanksUsed []uint8
	for _, b := range banksList {
		byteBanksUsed = append(byteBanksUsed, uint8(b))
	}

	res := &LinkResult{
		Binary:       finalBinary,
		EntryPoint:   entryAddress,
		BaseAddress:  l.config.BaseAddress,
		EndAddress:   commonCurrentAddr,
		TotalSize:    len(finalBinary),
		IsMultiBank:  isMultiBank,
		BanksUsed:    byteBanksUsed,
		Symbols:      globalSymbols,
		Segments:     placedSegments,
		Trampolines:  trampolines,
		BankPayloads: bankPayloads,
	}

	// 7. Gerar Mapa de Memória (.map)
	if l.config.MapFile != "" {
		if err := l.writeMapFile(l.config.MapFile, res); err != nil {
			return nil, fmt.Errorf("erro ao gerar mapa de memória: %w", err)
		}
	}

	return res, nil
}

// buildDispatcherCode gera o despachante central na Área Comum com suporte dual:
// se o MSX-DOS 2 com EXTBIO estiver ativo, ele saltará diretamente para a tabela do kernel;
// se não, executará o chaveamento direto via porta I/O (com tracking em variável para leitura segura).
func (l *Linker) buildDispatcherCode(baseAddr uint16) []byte {
	fallbackPutAddr := baseAddr + 7
	fallbackGetAddr := baseAddr + 13
	curBankAddr := baseAddr + 6

	fbPutLo := uint8(fallbackPutAddr & 0xFF)
	fbPutHi := uint8(fallbackPutAddr >> 8)
	fbGetLo := uint8(fallbackGetAddr & 0xFF)
	fbGetHi := uint8(fallbackGetAddr >> 8)
	curLo := uint8(curBankAddr & 0xFF)
	curHi := uint8(curBankAddr >> 8)

	return []byte{
		// +00: Musubi_PutP2 (JP Fallback_PutP2 - patcheado dinamicamente para o DOS 2 pelo loader)
		0xC3, fbPutLo, fbPutHi,
		// +03: Musubi_GetP2 (JP Fallback_GetP2 - patcheado dinamicamente para o DOS 2 pelo loader)
		0xC3, fbGetLo, fbGetHi,
		// +06: Musubi_CurP2 (armazena banco ativo se em modo fallback)
		0x01,
		// +07: Fallback_PutP2: OUT (port), A; LD (curBankAddr), A; RET
		0xD3, l.config.MapperPortPage2,
		0x32, curLo, curHi,
		0xC9,
		// +13: Fallback_GetP2: LD A, (curBankAddr); RET
		0x3A, curLo, curHi,
		0xC9,
	}
}

// buildTrampolineCode gera os opcodes Z80 do trampolim de troca de banco para Página 2
func (l *Linker) buildTrampolineCode(targetBank uint8, targetAddr uint16, putP2Addr uint16, getP2Addr uint16) []byte {
	getLo := uint8(getP2Addr & 0xFF)
	getHi := uint8((getP2Addr >> 8) & 0xFF)
	putLo := uint8(putP2Addr & 0xFF)
	putHi := uint8((putP2Addr >> 8) & 0xFF)
	addrLo := uint8(targetAddr & 0xFF)
	addrHi := uint8((targetAddr >> 8) & 0xFF)

	// Trampolim com preservação integral de registradores e retorno em A/HL:
	// CALL Musubi_GetP2 ; CD lo hi  (3 bytes) - obtém banco atual da Página 2
	// PUSH AF           ; F5        (1 byte)  - salva banco anterior na pilha
	// LD A, targetBank  ; 3E bank   (2 bytes) - carrega banco de destino
	// CALL Musubi_PutP2 ; CD lo hi  (3 bytes) - ativa banco de destino na Página 2
	// CALL targetAddr   ; CD lo hi  (3 bytes) - executa a rotina
	// EX AF, AF'        ; 08        (1 byte)  - protege retorno em A e flags
	// POP AF            ; F1        (1 byte)  - recupera o banco anterior
	// CALL Musubi_PutP2 ; CD lo hi  (3 bytes) - restaura o banco original da Página 2
	// EX AF, AF'        ; 08        (1 byte)  - restaura retorno em A e flags
	// RET               ; C9        (1 byte)  - retorna ao chamador original
	return []byte{
		0xCD, getLo, getHi,
		0xF5,
		0x3E, targetBank,
		0xCD, putLo, putHi,
		0xCD, addrLo, addrHi,
		0x08,
		0xF1,
		0xCD, putLo, putHi,
		0x08,
		0xC9,
	}
}

// buildBootstrapCode gera o código do Bootstrap Loader que copia os bancos para a Página 2
func (l *Linker) buildBootstrapCode(banks []*BankPayload, firstPayloadAddr uint16, userEntry uint16, putP2Addr uint16, getP2Addr uint16) []byte {
	var code []byte

	// 1. Alinhamento de Slot da Página 2 (0x8000..0xBFFF) com a Página 1 (RAM Mapper) via Porta 0xA8 (18 bytes)
	// Garante que a Página 2 aponte para o slot do Memory Mapper mesmo com cartuchos de DOS externos
	code = append(code,
		0xDB, 0xA8, // IN A, (0xA8)
		0x47,       // LD B, A
		0x0F,       // RRCA
		0x0F,       // RRCA
		0xE6, 0x03, // AND 0x03 (isola slot da Página 1 - RAM)
		0x07,       // RLCA
		0x07,       // RLCA
		0x07,       // RLCA
		0x07,       // RLCA (posiciona bits no slot da Página 2)
		0x4F,       // LD C, A
		0x78,       // LD A, B
		0xE6, 0xCF, // AND 0xCF (limpa slot atual da Página 2)
		0xB1,       // OR C    (aplica slot da RAM na Página 2)
		0xD3, 0xA8, // OUT (0xA8), A
	)

	// 2. Consulta da Tabela de Suporte da Memory Mapper via EXTBIO (0xFFCA) (32 bytes)
	putPatchAddr := putP2Addr + 1
	getPatchAddr := getP2Addr + 1
	putPatchLo := uint8(putPatchAddr & 0xFF)
	putPatchHi := uint8(putPatchAddr >> 8)
	getPatchLo := uint8(getPatchAddr & 0xFF)
	getPatchHi := uint8(getPatchAddr >> 8)

	code = append(code,
		0x3A, 0x20, 0xFB, // LD A, (0xFB20) - verifica flag HOKVLD
		0x0F,             // RRCA
		0x30, 0x1A,       // JR NC, +26 (salta se EXTBIO indisponível)
		0xAF,             // XOR A
		0x11, 0x02, 0x04, // LD DE, 0x0402 (D=4: Mapper, E=2: Obter Tabela)
		0xCD, 0xCA, 0xFF, // CALL 0xFFCA (EXTBIO -> HL = Tabela, A = total segmentos)
		0xB7,             // OR A
		0x28, 0x11,       // JR Z, +17 (salta se nenhum segmento)
		0xE5,             // PUSH HL
		0x11, 0x24, 0x00, // LD DE, 0x0024 (+24h = PUT_P2)
		0x19,             // ADD HL, DE
		0x22, putPatchLo, putPatchHi, // LD (Musubi_PutP2 + 1), HL
		0xE1,             // POP HL
		0x11, 0x27, 0x00, // LD DE, 0x0027 (+27h = GET_P2)
		0x19,             // ADD HL, DE
		0x22, getPatchLo, getPatchHi, // LD (Musubi_GetP2 + 1), HL
	)

	// 3. Imprimir "[L]" para indicar execução do loader via BDOS função 02h (21 bytes)
	code = append(code, 0x1E, '[', 0x0E, 0x02, 0xCD, 0x05, 0x00)
	code = append(code, 0x1E, 'L', 0x0E, 0x02, 0xCD, 0x05, 0x00)
	code = append(code, 0x1E, ']', 0x0E, 0x02, 0xCD, 0x05, 0x00)

	// 4. Copiar os payloads de cada banco para a Página 2 (0x8000) usando Musubi_PutP2 (16 bytes por banco)
	putLo := uint8(putP2Addr & 0xFF)
	putHi := uint8(putP2Addr >> 8)

	currSource := firstPayloadAddr
	for _, b := range banks {
		if len(b.Data) == 0 {
			continue
		}
		size := uint16(len(b.Data))
		// LD A, b.Bank
		code = append(code, 0x3E, b.Bank)
		// CALL Musubi_PutP2
		code = append(code, 0xCD, putLo, putHi)
		// LD HL, currSource
		code = append(code, 0x21, uint8(currSource&0xFF), uint8(currSource>>8))
		// LD DE, 0x8000
		code = append(code, 0x11, 0x00, 0x80)
		// LD BC, size
		code = append(code, 0x01, uint8(size&0xFF), uint8(size>>8))
		// LDIR
		code = append(code, 0xED, 0xB0)

		currSource += size
	}

	// 5. Chavear Página 2 para o primeiro banco paginável (ex: Banco 1) via Musubi_PutP2 (5 bytes)
	firstBank := uint8(1)
	if len(banks) > 0 {
		firstBank = banks[0].Bank
	}
	code = append(code, 0x3E, firstBank)
	code = append(code, 0xCD, putLo, putHi)

	// 6. Saltar para o ponto de entrada do usuário (3 bytes)
	code = append(code, 0xC3, uint8(userEntry&0xFF), uint8(userEntry>>8))

	return code
}

type bytesBuilder struct {
	buf []byte
}

func (b *bytesBuilder) write(data []byte) {
	b.buf = append(b.buf, data...)
}

func (b *bytesBuilder) bytes() []byte {
	return b.buf
}

func copyData(seg *mob.Segment) []byte {
	if seg.Type == mob.SegmentBSS || len(seg.Data) == 0 {
		return nil
	}
	d := make([]byte, len(seg.Data))
	copy(d, seg.Data)
	return d
}

func (l *Linker) writeMapFile(filename string, res *LinkResult) error {
	var sb strings.Builder

	sb.WriteString("===============================================================================\n")
	sb.WriteString("  KIZUNA MUSUBI LINKER — RELATÓRIO DE MAPA DE MEMÓRIA (MULTI-BANCO)\n")
	sb.WriteString("===============================================================================\n\n")

	sb.WriteString(fmt.Sprintf("Modo de Memória: %s\n", map[bool]string{false: "Monobanco", true: "Multi-Banco (Memory Mapper)"}[res.IsMultiBank]))
	sb.WriteString(fmt.Sprintf("Bancos Utilizados: %v\n", res.BanksUsed))
	sb.WriteString(fmt.Sprintf("Endereço Base:     0x%04X\n", res.BaseAddress))
	sb.WriteString(fmt.Sprintf("Ponto de Entrada:  0x%04X\n", res.EntryPoint))
	sb.WriteString(fmt.Sprintf("Tamanho do .COM:   %d bytes (0x%04X)\n\n", res.TotalSize, res.TotalSize))

	sb.WriteString("--- SEGMENTOS ALOCADOS ---\n")
	for i, seg := range res.Segments {
		modDesc := fmt.Sprintf("Mod: %2d", seg.ModuleIndex)
		if seg.ModuleIndex == -1 {
			modDesc = "LINKER-TRAMP"
		}
		sb.WriteString(fmt.Sprintf("[%2d] 0x%04X - 0x%04X | Banco: %2d | Tipo: %-4s | %s | Tam: %5d bytes\n",
			i, seg.BaseAddr, seg.BaseAddr+seg.Size, seg.Bank, seg.Type, modDesc, seg.Size))
	}
	sb.WriteString("\n")

	if len(res.Trampolines) > 0 {
		sb.WriteString("--- TRAMPOLINS DE BANK SWITCHING (ÁREA COMUM -> PÁGINA 2) ---\n")
		var sortedTramps []*Trampoline
		for _, t := range res.Trampolines {
			sortedTramps = append(sortedTramps, t)
		}
		sort.Slice(sortedTramps, func(i, j int) bool {
			return sortedTramps[i].Address < sortedTramps[j].Address
		})

		for _, t := range sortedTramps {
			sb.WriteString(fmt.Sprintf("0x%04X -> Alvo: '%s' no Banco %2d (0x%04X) | Tamanho: %d bytes\n",
				t.Address, t.SymbolName, t.TargetBank, t.TargetAddress, len(t.Code)))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("--- TABELA GLOBAL DE SÍMBOLOS ---\n")
	type symEntry struct {
		name string
		addr uint16
		bank uint8
		kind string
	}
	var symList []symEntry
	for name, s := range res.Symbols {
		symList = append(symList, symEntry{
			name: name,
			addr: s.Address,
			bank: s.Bank,
			kind: s.Kind.String(),
		})
	}
	sort.Slice(symList, func(i, j int) bool {
		if symList[i].bank == symList[j].bank {
			return symList[i].addr < symList[j].addr
		}
		return symList[i].bank < symList[j].bank
	})

	for _, s := range symList {
		sb.WriteString(fmt.Sprintf("0x%04X [Banco %2d] %-4s %s\n", s.addr, s.bank, s.kind, s.name))
	}

	return os.WriteFile(filename, []byte(sb.String()), 0644)
}

// LinkToFile é um helper conveniente para linkar arquivos .MOB e .HLIB e gravar diretamente o .COM
func LinkToFile(outputCom string, cfg LinkerConfig, files ...string) (*LinkResult, error) {
	var objects []*mob.ObjectFile
	var archives []*hako.Archive

	for _, f := range files {
		ext := strings.ToLower(filepath.Ext(f))
		if ext == ".hlib" {
			arc, err := hako.Open(f)
			if err != nil {
				return nil, fmt.Errorf("falha ao carregar biblioteca %s: %w", f, err)
			}
			archives = append(archives, arc)
		} else {
			obj, err := mob.LoadFromFile(f)
			if err != nil {
				return nil, fmt.Errorf("falha ao carregar objeto %s: %w", f, err)
			}
			objects = append(objects, obj)
		}
	}

	linker := NewLinker(cfg)
	res, err := linker.LinkWithLibraries(objects, archives)
	if err != nil {
		return nil, err
	}

	if err := os.WriteFile(outputCom, res.Binary, 0644); err != nil {
		return nil, fmt.Errorf("falha ao gravar executável %s: %w", outputCom, err)
	}

	return res, nil
}
