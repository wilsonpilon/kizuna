package dignac

import (
	"encoding/binary"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/wilsonpilon/kizuna/pkg/kaji80"
	"github.com/wilsonpilon/kizuna/pkg/mob"
)

type varKind int

const (
	varLocal varKind = iota
	varParam
	varGlobal
)

type varSymbol struct {
	Name   string
	Kind   varKind
	Type   string // "INTEGER", "STRING", "BOOLEAN", "SINGLE" ou "DOUBLE"
	Offset int    // offset a partir de IX (positivo para param, positivo absoluto para local onde local está em IX - Offset)
	Label  string // label para globais
}

// typeSize devolve o tamanho em bytes do armazenamento de uma variável pelo
// seu tipo declarado -- usado tanto pra alocar o frame de locais (que hoje
// soma 2 bytes fixos por variável, sem olhar pro tipo) quanto pra emitir a
// diretiva de dado de uma global.
func typeSize(t string) int {
	switch strings.ToUpper(t) {
	case "BOOLEAN":
		return 1
	case "STRING":
		return 256 // 1 byte de tamanho + até 255 bytes de dados (SPEC.md §7)
	case "SINGLE":
		return 4 // IEEE754 binary32
	case "DOUBLE":
		return 8 // IEEE754 binary64
	default: // INTEGER
		return 2
	}
}

// storageDirective devolve a diretiva Assembly que reserva/inicializa o
// armazenamento de uma variável GLOBAL do tipo dado.
func storageDirective(t string) string {
	switch strings.ToUpper(t) {
	case "BOOLEAN":
		return "DB 00h"
	case "STRING":
		return "DS 256"
	case "SINGLE":
		return "DS 4"
	case "DOUBLE":
		return "DS 8"
	default: // INTEGER
		return "DW 0000h"
	}
}

// CodeGenerator traduz a AST de MSX-BASIC Dignified em Assembly Z80 compatível com KAJI80
type CodeGenerator struct {
	module         *ModuleNode
	asm            strings.Builder
	labelCounter   int
	stringLits     map[string]string // texto -> label (ex: "StrLit_1")
	globals        map[string]*varSymbol
	currentLocals  map[string]*varSymbol
	currentParams  map[string]*varSymbol
	localFrameSize int

	// Flags de rastreamento de dependências da MSXLIB
	needsMul16    bool
	needsDiv16    bool
	needsPrintDec bool
	needsPrintStr bool
	needsPrintChr bool
	needsPset     bool
	needsLine     bool
	needsBoxFill  bool
	needsCls      bool
	needsBeep     bool
	needsChgMod   bool

	needsSpriteSet     bool
	needsSpriteDefine  bool
	needsSpriteHideAll bool
	needsPlaySequence  bool
	needsMuteAll       bool
	needsFileOpen      bool
	needsFileCreate    bool
	needsFileClose     bool
	needsFileWrite     bool
	needsFileSeek      bool
	needsPrintDecBuf   bool
	needsStrCopyLen    bool
	needsPrintLenStr   bool
	needsForStepSign   bool

	extraData []string     // blocos DB pré-renderizados (padrões de sprite, sequências MML, caminhos de arquivo, literais de PRINT #n)
	fileNums  map[int]bool // números de arquivo (#n) realmente usados no módulo
}

// NewCodeGenerator cria um novo gerador de código
func NewCodeGenerator(module *ModuleNode) *CodeGenerator {
	return &CodeGenerator{
		module:     module,
		stringLits: make(map[string]string),
		globals:    make(map[string]*varSymbol),
		fileNums:   make(map[int]bool),
	}
}

func (cg *CodeGenerator) newLabel(prefix string) string {
	cg.labelCounter++
	return fmt.Sprintf("%s_%d", prefix, cg.labelCounter)
}

func (cg *CodeGenerator) getStringLabel(str string) string {
	if lbl, ok := cg.stringLits[str]; ok {
		return lbl
	}
	lbl := cg.newLabel("StrLit")
	cg.stringLits[str] = lbl
	return lbl
}

// GenerateAsm emite o código assembly Z80 completo formatado para o KAJI80
func (cg *CodeGenerator) GenerateAsm() (string, error) {
	cg.asm.Reset()

	modName := cg.module.Name
	if modName == "" {
		modName = "DignifiedModule"
	}

	cg.asm.WriteString("; =============================================================================\n")
	cg.asm.WriteString("; Código gerado pelo compilador MSX-BASIC Dignified DIGNAC - Kizuna Toolchain\n")
	cg.asm.WriteString(fmt.Sprintf("; Módulo: %s | Banco: %d\n", modName, cg.module.Bank))
	cg.asm.WriteString("; =============================================================================\n\n")

	cg.asm.WriteString(fmt.Sprintf("MODULE %s\n", modName))
	cg.asm.WriteString(fmt.Sprintf("BANK %d\n\n", cg.module.Bank))

	// Coleta variáveis globais
	for _, g := range cg.module.Globals {
		for _, v := range g.Decls {
			lower := strings.ToLower(v.Name)
			cg.globals[lower] = &varSymbol{
				Name:  v.Name,
				Kind:  varGlobal,
				Type:  v.Type,
				Label: fmt.Sprintf("Global_%s", sanitizeIdent(v.Name)),
			}
		}
	}

	// Buffer temporário para o corpo do código
	var codeBody strings.Builder

	// Gera procedimentos
	for _, proc := range cg.module.Procedures {
		if err := cg.generateProcedure(&codeBody, proc); err != nil {
			return "", err
		}
	}

	// Exportações (PUBLIC)
	publics := make([]string, 0)
	for _, p := range cg.module.Publics {
		publics = append(publics, p)
	}
	hasStart := false
	hasMain := false
	var mainProcName string
	for _, proc := range cg.module.Procedures {
		if strings.EqualFold(proc.Name, "Start") {
			hasStart = true
		}
		if strings.EqualFold(proc.Name, "Main") {
			hasMain = true
			mainProcName = proc.Name
		}
	}

	if hasMain && !hasStart {
		if !containsString(publics, "Start") {
			publics = append(publics, "Start")
		}
	}

	if len(publics) > 0 {
		cg.asm.WriteString(fmt.Sprintf("PUBLIC %s\n", strings.Join(publics, ", ")))
	}

	// Importações (EXTERN)
	externs := make([]string, 0)
	for _, ext := range cg.module.Externs {
		externs = append(externs, ext)
	}
	if hasMain && !hasStart && !containsString(externs, "BDOS_Exit") {
		externs = append(externs, "BDOS_Exit")
	}
	if cg.needsMul16 && !containsString(externs, "Mul16") {
		externs = append(externs, "Mul16")
	}
	if cg.needsDiv16 && !containsString(externs, "Div16") {
		externs = append(externs, "Div16")
	}
	if cg.needsPrintStr && !containsString(externs, "BDOS_PrintString") {
		externs = append(externs, "BDOS_PrintString")
	}
	if cg.needsPrintChr && !containsString(externs, "BDOS_PrintChar") {
		externs = append(externs, "BDOS_PrintChar")
	}
	if cg.needsPrintDec && !containsString(externs, "PrintDec16") {
		externs = append(externs, "PrintDec16")
	}
	if cg.needsPset && !containsString(externs, "VDP_PSet") {
		externs = append(externs, "VDP_PSet")
	}
	if cg.needsLine && !containsString(externs, "VDP_Line") {
		externs = append(externs, "VDP_Line")
	}
	if cg.needsBoxFill && !containsString(externs, "VDP_BoxFill") {
		externs = append(externs, "VDP_BoxFill")
	}
	if cg.needsCls && !containsString(externs, "BIOS_CLS") {
		externs = append(externs, "BIOS_CLS")
	}
	if cg.needsBeep && !containsString(externs, "BIOS_BEEP") {
		externs = append(externs, "BIOS_BEEP")
	}
	if cg.needsChgMod && !containsString(externs, "BIOS_CHGMOD") {
		externs = append(externs, "BIOS_CHGMOD")
	}
	if cg.needsSpriteSet && !containsString(externs, "VDP_SpriteSet") {
		externs = append(externs, "VDP_SpriteSet")
	}
	if cg.needsSpriteDefine && !containsString(externs, "VDP_SpriteDefine") {
		externs = append(externs, "VDP_SpriteDefine")
	}
	if cg.needsSpriteHideAll && !containsString(externs, "VDP_SpriteHideAll") {
		externs = append(externs, "VDP_SpriteHideAll")
	}
	if cg.needsPlaySequence && !containsString(externs, "PSG_PlaySequence") {
		externs = append(externs, "PSG_PlaySequence")
	}
	if cg.needsMuteAll && !containsString(externs, "PSG_MuteAll") {
		externs = append(externs, "PSG_MuteAll")
	}
	if cg.needsFileOpen && !containsString(externs, "BDOS_FileOpen") {
		externs = append(externs, "BDOS_FileOpen")
	}
	if cg.needsFileCreate && !containsString(externs, "BDOS_FileCreate") {
		externs = append(externs, "BDOS_FileCreate")
	}
	if cg.needsFileClose && !containsString(externs, "BDOS_FileClose") {
		externs = append(externs, "BDOS_FileClose")
	}
	if cg.needsFileWrite && !containsString(externs, "BDOS_FileWrite") {
		externs = append(externs, "BDOS_FileWrite")
	}
	if cg.needsFileSeek && !containsString(externs, "BDOS_FileSeek") {
		externs = append(externs, "BDOS_FileSeek")
	}
	if cg.needsPrintDecBuf && !containsString(externs, "PrintDec16ToBuffer") {
		externs = append(externs, "PrintDec16ToBuffer")
	}
	if cg.needsStrCopyLen && !containsString(externs, "StrCopyLen") {
		externs = append(externs, "StrCopyLen")
	}
	if cg.needsPrintLenStr && !containsString(externs, "BDOS_PrintLenStr") {
		externs = append(externs, "BDOS_PrintLenStr")
	}

	if len(externs) > 0 {
		cg.asm.WriteString(fmt.Sprintf("EXTERN %s\n\n", strings.Join(externs, ", ")))
	} else {
		cg.asm.WriteString("\n")
	}

	// Ponto de entrada padrão para MSX-DOS 2 caso exista PROCEDURE Main
	if hasMain && !hasStart {
		cg.asm.WriteString("; --- Ponto de Entrada para Executável MSX-DOS 2 ---\n")
		cg.asm.WriteString("Start:\n")
		cg.asm.WriteString(fmt.Sprintf("    CALL %s\n", mainProcName))
		cg.asm.WriteString("    CALL BDOS_Exit\n\n")
	}

	// Escreve o código dos procedimentos
	cg.asm.WriteString(codeBody.String())

	// Seção de Literais de Texto
	if len(cg.stringLits) > 0 {
		cg.asm.WriteString("; --- Literais de String ---\n")
		for str, lbl := range cg.stringLits {
			cg.asm.WriteString(fmt.Sprintf("%s:\n", lbl))
			cg.asm.WriteString(fmt.Sprintf("    DB \"%s$\"\n", escapeString(str)))
		}
		cg.asm.WriteString("\n")
	}

	// Seção de Dados Extras: padrões de sprite (SPRITE PATTERN), sequências
	// de música traduzidas do MML (PLAY), caminhos de arquivo e literais de
	// PRINT #n (ASCIIZ/sem terminador -- não usam o '$' dos literais acima,
	// que é específico de BDOS_PrintString). Ordem de inserção já é
	// determinística (um slice, não um map) -- sem risco do mesmo
	// não-determinismo já corrigido na tabela de símbolos do KAJI80.
	if len(cg.extraData) > 0 {
		cg.asm.WriteString("; --- Sprites / Música / Arquivos ---\n")
		for _, block := range cg.extraData {
			cg.asm.WriteString(block)
		}
		cg.asm.WriteString("\n")
	}

	// Células de rascunho de PUT SPRITE (ver o comentário no case
	// *PutSpriteStmt em generateStmt)
	if cg.needsSpriteSet {
		cg.asm.WriteString("; --- Rascunho de PUT SPRITE ---\n")
		cg.asm.WriteString("DGN_Sprite_Idx: DB 00h\n")
		cg.asm.WriteString("DGN_Sprite_X: DB 00h\n")
		cg.asm.WriteString("DGN_Sprite_Y: DB 00h\n")
		cg.asm.WriteString("DGN_Sprite_Pattern: DB 00h\n")
		cg.asm.WriteString("DGN_Sprite_Color: DB 00h\n\n")
	}

	// Sinal do STEP de um FOR com STEP explícito, guardado 1x por iteração
	// antes do teste de término decidir se o laço é ascendente ou
	// descendente (ver o comentário no case *ForStmt em generateStmt).
	// Compartilhada entre todos os FOR...STEP do módulo -- segura mesmo com
	// laços aninhados porque é sempre escrita e lida dentro do mesmo bloco
	// de teste, nunca através da execução do corpo.
	if cg.needsForStepSign {
		cg.asm.WriteString("; --- Sinal do STEP (FOR...STEP) ---\n")
		cg.asm.WriteString("DGN_ForStepSign: DB 00h\n\n")
	}

	// Handles de arquivo (#n) -- uma variável global de 1 byte por número de
	// arquivo realmente usado no módulo, preenchida por OPEN e lida por
	// CLOSE/PRINT #n. Números ordenados para saída determinística (mapa do
	// Go embaralha a ordem de iteração).
	if len(cg.fileNums) > 0 {
		nums := make([]int, 0, len(cg.fileNums))
		for n := range cg.fileNums {
			nums = append(nums, n)
		}
		sort.Ints(nums)
		cg.asm.WriteString("; --- Handles de Arquivo (#n) ---\n")
		for _, n := range nums {
			cg.asm.WriteString(fmt.Sprintf("DGN_FileHandle_%d:\n    DB 00h\n", n))
		}
		cg.asm.WriteString("DGN_FileNumBuf: DS 6\n")
		cg.asm.WriteString("DGN_CRLF: DB 0Dh, 0Ah\n\n")
	}

	// Seção de Variáveis Globais. Nomes ordenados antes de emitir -- iterar
	// cg.globals (um map do Go) direto embaralharia a ordem a cada
	// remontagem, mesma classe de não-determinismo já corrigida na tabela de
	// símbolos do KAJI80 (ver histórico do projeto).
	if len(cg.globals) > 0 {
		names := make([]string, 0, len(cg.globals))
		for name := range cg.globals {
			names = append(names, name)
		}
		sort.Strings(names)

		cg.asm.WriteString("; --- Variáveis Globais ---\n")
		for _, name := range names {
			sym := cg.globals[name]
			cg.asm.WriteString(fmt.Sprintf("%s:\n    %s\n", sym.Label, storageDirective(sym.Type)))
		}
		cg.asm.WriteString("\n")
	}

	cg.asm.WriteString("ENDMOD\n")
	return cg.asm.String(), nil
}

// Compile compila diretamente para um arquivo de objeto .MOB
func (cg *CodeGenerator) Compile() (*mob.ObjectFile, string, error) {
	asmSource, err := cg.GenerateAsm()
	if err != nil {
		return nil, "", err
	}

	asm := kaji80.NewAssembler()
	obj, err := asm.Assemble(asmSource)
	if err != nil {
		return nil, asmSource, fmt.Errorf("erro na montagem do assembly Z80 gerado pelo DIGNAC: %w\nCódigo gerado:\n%s", err, asmSource)
	}

	return obj, asmSource, nil
}

func (cg *CodeGenerator) generateProcedure(sb *strings.Builder, proc *ProcedureNode) error {
	cg.currentLocals = make(map[string]*varSymbol)
	cg.currentParams = make(map[string]*varSymbol)

	numParams := len(proc.Params)
	// Convenção de chamada Kizuna (especificação §7):
	// Parâmetros empilhados da esquerda para a direita (primeiro fica mais fundo na pilha)
	// PUSH P1, PUSH P2, CALL -> (IX+4) = P2, (IX+6) = P1
	for i, p := range proc.Params {
		offset := 4 + 2*(numParams-1-i)
		cg.currentParams[strings.ToLower(p.Name)] = &varSymbol{
			Name:   p.Name,
			Kind:   varParam,
			Type:   p.Type,
			Offset: offset,
		}
	}

	// Alocação de variáveis locais no frame de pilha (IX - Offset). Cada
	// variável reserva o tamanho do seu próprio tipo (typeSize) em vez dos
	// 2 bytes fixos de antes -- uma STRING local, por exemplo, ocupa 256
	// bytes do frame, não 2.
	localOffset := 0
	for _, locDecl := range proc.Locals {
		for _, v := range locDecl.Decls {
			localOffset += typeSize(v.Type)
			cg.currentLocals[strings.ToLower(v.Name)] = &varSymbol{
				Name:   v.Name,
				Kind:   varLocal,
				Type:   v.Type,
				Offset: localOffset,
			}
		}
	}
	cg.localFrameSize = localOffset

	sb.WriteString(fmt.Sprintf("; -----------------------------------------------------------------------------\n"))
	sb.WriteString(fmt.Sprintf("; Procedimento: %s (Locais: %d bytes, Parâmetros: %d)\n", proc.Name, cg.localFrameSize, numParams))
	sb.WriteString(fmt.Sprintf("; -----------------------------------------------------------------------------\n"))
	sb.WriteString(fmt.Sprintf("%s:\n", proc.Name))

	// Prólogo da ABI Kizuna com frame pointer IX
	sb.WriteString("    PUSH IX\n")
	sb.WriteString("    LD IX, 0000h\n")
	sb.WriteString("    ADD IX, SP\n")

	if cg.localFrameSize > 0 {
		sb.WriteString(fmt.Sprintf("    LD HL, -%d\n", cg.localFrameSize))
		sb.WriteString("    ADD HL, SP\n")
		sb.WriteString("    LD SP, HL\n")
	}

	// Gera corpo do procedimento
	for _, stmt := range proc.Body {
		if err := cg.generateStmt(sb, stmt); err != nil {
			return err
		}
	}

	// Epílogo
	sb.WriteString(fmt.Sprintf(".Exit_%s:\n", proc.Name))
	sb.WriteString("    LD SP, IX\n")
	sb.WriteString("    POP IX\n")
	sb.WriteString("    RET\n\n")

	return nil
}

func (cg *CodeGenerator) generateStmt(sb *strings.Builder, stmt Stmt) error {
	switch s := stmt.(type) {
	case *AssignStmt:
		switch cg.varType(s.VarName) {
		case "STRING":
			return cg.generateStringAssign(sb, s)
		case "SINGLE", "DOUBLE":
			return cg.generateFloatAssign(sb, s, cg.varType(s.VarName))
		default:
			// Avalia expressão -> HL
			if err := cg.generateExpr(sb, s.Value); err != nil {
				return err
			}
			// Armazena HL na variável
			return cg.storeVar(sb, s.VarName)
		}

	case *ForStmt:
		// FOR var = start TO end [STEP step] ... NEXT [var]
		loopStart := cg.newLabel("For_Start")
		loopEnd := cg.newLabel("For_End")
		hasStep := s.Step != nil

		// 1. Inicializa variável com Start
		if err := cg.generateExpr(sb, s.Start); err != nil {
			return err
		}
		if err := cg.storeVar(sb, s.VarName); err != nil {
			return err
		}

		// 1b. Se houver STEP explícito, o SINAL é calculado uma única vez
		// (assumindo que não muda durante o laço -- mesma suposição de
		// qualquer BASIC clássico) e guardado numa célula compartilhada do
		// módulo, lida a cada iteração pelo teste de término (passo 2)
		// pra decidir se o laço é ascendente ou descendente.
		if hasStep {
			cg.needsForStepSign = true
			if err := cg.generateExpr(sb, s.Step); err != nil {
				return err
			}
			sb.WriteString("    LD A, H\n") // bit 7 do byte alto = sinal do valor de 16 bits
			sb.WriteString("    LD (DGN_ForStepSign), A\n")
		}

		sb.WriteString(fmt.Sprintf("%s:\n", loopStart))

		// 2. Condição de término. SBC HL,DE (Var - End) dá Z (Var==End,
		// sempre continua, nas duas direções) e C (Var<End, sem sinal). Sem
		// STEP explícito (incremento sempre +1) ou com STEP positivo,
		// continua enquanto Var<=End; com STEP negativo, continua enquanto
		// Var>=End -- por isso o "Var>End" e o "Var<End" cada um decide
		// parar ou continuar consultando DGN_ForStepSign, em vez de uma
		// comparação fixa que só cobria o caso ascendente (bug real: um
		// STEP negativo nunca detectava o fim do laço corretamente).
		if err := cg.loadVar(sb, s.VarName); err != nil {
			return err
		}
		sb.WriteString("    PUSH HL\n")
		if err := cg.generateExpr(sb, s.End); err != nil {
			return err
		}
		sb.WriteString("    EX DE, HL\n") // DE = End
		sb.WriteString("    POP HL\n")    // HL = Var
		sb.WriteString("    OR A\n")
		sb.WriteString("    SBC HL, DE\n")

		bodyLbl := cg.newLabel("For_Body")
		gtLbl := cg.newLabel("For_GT") // Var > End (NC e NZ)
		ltLbl := cg.newLabel("For_LT") // Var < End (C)

		sb.WriteString(fmt.Sprintf("    JP Z, %s\n", bodyLbl))  // Var == End -> sempre continua
		sb.WriteString(fmt.Sprintf("    JP C, %s\n", ltLbl))    // Var < End
		sb.WriteString(fmt.Sprintf("%s:\n", gtLbl))             // Var > End
		if hasStep {
			sb.WriteString("    LD A, (DGN_ForStepSign)\n")
			sb.WriteString("    AND 80h\n")
			sb.WriteString(fmt.Sprintf("    JP NZ, %s\n", bodyLbl)) // descendente: Var>End ainda continua
		}
		sb.WriteString(fmt.Sprintf("    JP %s\n", loopEnd)) // ascendente (ou sem STEP): Var>End encerra
		sb.WriteString(fmt.Sprintf("%s:\n", ltLbl))         // Var < End
		if hasStep {
			sb.WriteString("    LD A, (DGN_ForStepSign)\n")
			sb.WriteString("    AND 80h\n")
			sb.WriteString(fmt.Sprintf("    JP NZ, %s\n", loopEnd)) // descendente: Var<End encerra
		}
		sb.WriteString(fmt.Sprintf("%s:\n", bodyLbl))

		// 3. Executa corpo
		for _, child := range s.Body {
			if err := cg.generateStmt(sb, child); err != nil {
				return err
			}
		}

		// 4. Incremento: var = var + step
		if err := cg.loadVar(sb, s.VarName); err != nil {
			return err
		}
		if s.Step != nil {
			sb.WriteString("    PUSH HL\n")
			if err := cg.generateExpr(sb, s.Step); err != nil {
				return err
			}
			sb.WriteString("    EX DE, HL\n")
			sb.WriteString("    POP HL\n")
			sb.WriteString("    ADD HL, DE\n")
		} else {
			sb.WriteString("    INC HL\n")
		}
		if err := cg.storeVar(sb, s.VarName); err != nil {
			return err
		}

		sb.WriteString(fmt.Sprintf("    JP %s\n", loopStart))
		sb.WriteString(fmt.Sprintf("%s:\n", loopEnd))
		return nil

	case *IfStmt:
		elseLbl := cg.newLabel("If_Else")
		endLbl := cg.newLabel("If_End")

		if err := cg.generateExpr(sb, s.Condition); err != nil {
			return err
		}
		// HL contém a condição (0 = falso)
		sb.WriteString("    LD A, H\n")
		sb.WriteString("    OR L\n")
		if len(s.ElseBody) > 0 {
			sb.WriteString(fmt.Sprintf("    JP Z, %s\n", elseLbl))
		} else {
			sb.WriteString(fmt.Sprintf("    JP Z, %s\n", endLbl))
		}

		for _, child := range s.ThenBody {
			if err := cg.generateStmt(sb, child); err != nil {
				return err
			}
		}

		if len(s.ElseBody) > 0 {
			sb.WriteString(fmt.Sprintf("    JP %s\n", endLbl))
			sb.WriteString(fmt.Sprintf("%s:\n", elseLbl))
			for _, child := range s.ElseBody {
				if err := cg.generateStmt(sb, child); err != nil {
					return err
				}
			}
		}

		sb.WriteString(fmt.Sprintf("%s:\n", endLbl))
		return nil

	case *WhileStmt:
		startLbl := cg.newLabel("While_Start")
		endLbl := cg.newLabel("While_End")

		sb.WriteString(fmt.Sprintf("%s:\n", startLbl))
		if err := cg.generateExpr(sb, s.Condition); err != nil {
			return err
		}
		sb.WriteString("    LD A, H\n")
		sb.WriteString("    OR L\n")
		sb.WriteString(fmt.Sprintf("    JP Z, %s\n", endLbl))

		for _, child := range s.Body {
			if err := cg.generateStmt(sb, child); err != nil {
				return err
			}
		}

		sb.WriteString(fmt.Sprintf("    JP %s\n", startLbl))
		sb.WriteString(fmt.Sprintf("%s:\n", endLbl))
		return nil

	case *DoLoopStmt:
		startLbl := cg.newLabel("Do_Start")
		endLbl := cg.newLabel("Do_End")

		sb.WriteString(fmt.Sprintf("%s:\n", startLbl))
		if s.Condition != nil && s.IsWhile {
			if err := cg.generateExpr(sb, s.Condition); err != nil {
				return err
			}
			sb.WriteString("    LD A, H\n")
			sb.WriteString("    OR L\n")
			sb.WriteString(fmt.Sprintf("    JP Z, %s\n", endLbl))
		}

		for _, child := range s.Body {
			if err := cg.generateStmt(sb, child); err != nil {
				return err
			}
		}

		sb.WriteString(fmt.Sprintf("    JP %s\n", startLbl))
		sb.WriteString(fmt.Sprintf("%s:\n", endLbl))
		return nil

	case *CallStmt:
		// Empilha argumentos da esquerda para a direita (Kizuna ABI)
		for _, arg := range s.Args {
			if err := cg.generateExpr(sb, arg); err != nil {
				return err
			}
			sb.WriteString("    PUSH HL\n")
		}
		sb.WriteString(fmt.Sprintf("    CALL %s\n", s.Name))
		if len(s.Args) > 0 {
			sb.WriteString(fmt.Sprintf("    LD HL, %d\n", len(s.Args)*2))
			sb.WriteString("    ADD HL, SP\n")
			sb.WriteString("    LD SP, HL\n")
		}
		return nil

	case *PsetStmt:
		// PSET (x, y)[, color]
		// VDP_PSet espera: BC = X, DE = Y, A = Color
		cg.needsPset = true
		if err := cg.generateExpr(sb, s.X); err != nil {
			return err
		}
		sb.WriteString("    PUSH HL\n") // salva X

		if err := cg.generateExpr(sb, s.Y); err != nil {
			return err
		}
		sb.WriteString("    PUSH HL\n") // salva Y

		if s.Color != nil {
			if err := cg.generateExpr(sb, s.Color); err != nil {
				return err
			}
		} else {
			sb.WriteString("    LD HL, 000Fh\n") // Cor 15 padrão
		}
		sb.WriteString("    LD A, L\n") // A = Color
		sb.WriteString("    POP DE\n")  // DE = Y
		sb.WriteString("    POP BC\n")  // BC = X
		sb.WriteString("    CALL VDP_PSet\n")
		return nil

	case *LineStmt:
		// LINE (x1,y1)-(x2,y2)[, color][, B | BF]
		if s.BoxFill {
			cg.needsBoxFill = true
		} else if s.Box {
			cg.needsLine = true
		} else {
			cg.needsLine = true
		}

		// Empilha coordenadas e cor: X1, Y1, X2, Y2, Color
		if err := cg.generateExpr(sb, s.X1); err != nil {
			return err
		}
		sb.WriteString("    PUSH HL\n")
		if err := cg.generateExpr(sb, s.Y1); err != nil {
			return err
		}
		sb.WriteString("    PUSH HL\n")
		if err := cg.generateExpr(sb, s.X2); err != nil {
			return err
		}
		sb.WriteString("    PUSH HL\n")
		if err := cg.generateExpr(sb, s.Y2); err != nil {
			return err
		}
		sb.WriteString("    PUSH HL\n")

		if s.Color != nil {
			if err := cg.generateExpr(sb, s.Color); err != nil {
				return err
			}
		} else {
			sb.WriteString("    LD HL, 000Fh\n")
		}
		sb.WriteString("    PUSH HL\n")

		if s.BoxFill {
			sb.WriteString("    CALL VDP_BoxFill\n")
		} else {
			sb.WriteString("    CALL VDP_Line\n")
		}
		// Limpeza da pilha (5 argumentos de 16-bit = 10 bytes)
		sb.WriteString("    LD HL, 10\n")
		sb.WriteString("    ADD HL, SP\n")
		sb.WriteString("    LD SP, HL\n")
		return nil

	case *PrintStmt:
		if s.FileNum != nil {
			return cg.generatePrintFileStmt(sb, s)
		}
		for _, arg := range s.Args {
			switch a := arg.(type) {
			case *StringExpr:
				cg.needsPrintStr = true
				lbl := cg.getStringLabel(a.Value)
				sb.WriteString(fmt.Sprintf("    LD DE, %s\n", lbl))
				sb.WriteString("    CALL BDOS_PrintString\n")
			case *VarExpr:
				switch cg.varType(a.Name) {
				case "STRING":
					cg.needsPrintLenStr = true
					if err := cg.loadVarAddress(sb, a.Name); err != nil {
						return err
					}
					sb.WriteString("    CALL BDOS_PrintLenStr\n")
				case "SINGLE", "DOUBLE":
					return fmt.Errorf("PRINT de '%s': impressão de ponto flutuante (SINGLE/DOUBLE) ainda não implementada", a.Name)
				default:
					cg.needsPrintDec = true
					if err := cg.generateExpr(sb, a); err != nil {
						return err
					}
					sb.WriteString("    CALL PrintDec16\n")
				}
			default:
				cg.needsPrintDec = true
				if err := cg.generateExpr(sb, a); err != nil {
					return err
				}
				sb.WriteString("    CALL PrintDec16\n")
			}
		}
		if !s.TrailingSemicolon {
			cg.needsPrintChr = true
			sb.WriteString("    LD E, 0Dh\n")
			sb.WriteString("    CALL BDOS_PrintChar\n")
			sb.WriteString("    LD E, 0Ah\n")
			sb.WriteString("    CALL BDOS_PrintChar\n")
		}
		return nil

	case *PutSpriteStmt:
		// VDP_SpriteSet espera: A=índice, H=Y, L=X, D=padrão, E=cor.
		// Avalia os 5 valores primeiro para células de rascunho (mesmo
		// idioma de VDP_PSet_ColorArg) -- evita um valor sobrescrever outro
		// no meio do caminho, já que só A tem endereçamento absoluto direto
		// (LD A,(nn)); os demais só recebem via LD r,A.
		cg.needsSpriteSet = true
		if err := cg.generateExpr(sb, s.Index); err != nil {
			return err
		}
		sb.WriteString("    LD A, L\n    LD (DGN_Sprite_Idx), A\n")
		if err := cg.generateExpr(sb, s.X); err != nil {
			return err
		}
		sb.WriteString("    LD A, L\n    LD (DGN_Sprite_X), A\n")
		if err := cg.generateExpr(sb, s.Y); err != nil {
			return err
		}
		sb.WriteString("    LD A, L\n    LD (DGN_Sprite_Y), A\n")
		if err := cg.generateExpr(sb, s.Color); err != nil {
			return err
		}
		sb.WriteString("    LD A, L\n    LD (DGN_Sprite_Color), A\n")
		if err := cg.generateExpr(sb, s.Pattern); err != nil {
			return err
		}
		sb.WriteString("    LD A, L\n    LD (DGN_Sprite_Pattern), A\n")

		sb.WriteString("    LD A, (DGN_Sprite_Y)\n    LD H, A\n")
		sb.WriteString("    LD A, (DGN_Sprite_X)\n    LD L, A\n")
		sb.WriteString("    LD A, (DGN_Sprite_Pattern)\n    LD D, A\n")
		sb.WriteString("    LD A, (DGN_Sprite_Color)\n    LD E, A\n")
		sb.WriteString("    LD A, (DGN_Sprite_Idx)\n")
		sb.WriteString("    CALL VDP_SpriteSet\n")
		return nil

	case *SpritePatternStmt:
		cg.needsSpriteDefine = true
		parts := make([]string, len(s.Bytes))
		for i, bExpr := range s.Bytes {
			numExpr, ok := bExpr.(*NumberExpr)
			if !ok {
				return fmt.Errorf("SPRITE PATTERN: os bytes do padrão precisam ser constantes numéricas literais")
			}
			if numExpr.Value < 0 || numExpr.Value > 255 {
				return fmt.Errorf("SPRITE PATTERN: byte de padrão %d fora do intervalo 0..255", numExpr.Value)
			}
			parts[i] = fmt.Sprintf("%02Xh", numExpr.Value)
		}
		label := cg.newLabel("SpritePattern")
		cg.extraData = append(cg.extraData, fmt.Sprintf("%s:\n    DB %s\n", label, strings.Join(parts, ", ")))

		if err := cg.generateExpr(sb, s.Pattern); err != nil {
			return err
		}
		sb.WriteString("    LD A, L\n")
		sb.WriteString(fmt.Sprintf("    LD HL, %s\n", label))
		sb.WriteString(fmt.Sprintf("    LD BC, %d\n", len(s.Bytes)))
		sb.WriteString("    CALL VDP_SpriteDefine\n")
		return nil

	case *SpriteOffStmt:
		cg.needsSpriteHideAll = true
		sb.WriteString("    CALL VDP_SpriteHideAll\n")
		return nil

	case *PlayStmt:
		cg.needsPlaySequence = true
		cg.needsMuteAll = true
		events, err := parseMML(s.MML)
		if err != nil {
			return fmt.Errorf("PLAY: %w", err)
		}
		label := cg.newLabel("PlaySeq")
		cg.extraData = append(cg.extraData, fmt.Sprintf("%s:\n%s", label, encodeEvents(events)))
		sb.WriteString(fmt.Sprintf("    LD HL, %s\n", label))
		sb.WriteString("    CALL PSG_PlaySequence\n")
		// O PSG é um chip com estado: sem isto, a última nota tocada
		// continua soando indefinidamente (mesmo depois do programa sair
		// de volta ao MSX-DOS) até algo mais reprogramar o canal --
		// diferente de uma instrução que "termina" e não deixa efeito
		// colateral pendente, o que seria a expectativa razoável de PLAY
		// como statement autocontido.
		sb.WriteString("    CALL PSG_MuteAll\n")
		return nil

	case *OpenStmt:
		pathStr, ok := s.Path.(*StringExpr)
		if !ok {
			return fmt.Errorf("OPEN exige um caminho literal (ex: OPEN \"TEST.TXT\" FOR OUTPUT AS #1)")
		}
		cg.fileNums[s.FileNum] = true
		pathLabel := cg.newLabel("FilePath")
		cg.extraData = append(cg.extraData, fmt.Sprintf("%s:\n    DB \"%s\", 00h\n", pathLabel, escapeString(pathStr.Value)))

		handleLabel := fmt.Sprintf("DGN_FileHandle_%d", s.FileNum)

		switch s.Mode {
		case "INPUT":
			cg.needsFileOpen = true
			sb.WriteString(fmt.Sprintf("    LD DE, %s\n", pathLabel))
			sb.WriteString("    LD A, 01h\n") // somente leitura
			sb.WriteString("    CALL BDOS_FileOpen\n")
			sb.WriteString("    LD A, B\n") // handle retornado em B
			sb.WriteString(fmt.Sprintf("    LD (%s), A\n", handleLabel))

		case "APPEND":
			// FOR APPEND precisa abrir o arquivo EXISTENTE (sem truncar) e
			// só então posicionar o ponteiro no fim antes de qualquer
			// escrita -- diferente de OUTPUT, que sempre cria do zero.
			// BDOS_FileCreate/ATTR_NORMAL trunca um arquivo já existente,
			// então cai nele só como fallback se o arquivo ainda não
			// existir (BDOS_FileOpen retornando erro em A).
			cg.needsFileOpen = true
			cg.needsFileCreate = true
			cg.needsFileSeek = true
			openedLbl := cg.newLabel("Append_Opened")

			sb.WriteString(fmt.Sprintf("    LD DE, %s\n", pathLabel))
			sb.WriteString("    LD A, 00h\n") // leitura+escrita
			sb.WriteString("    CALL BDOS_FileOpen\n")
			sb.WriteString("    OR A\n")
			sb.WriteString(fmt.Sprintf("    JP Z, %s\n", openedLbl))
			// Arquivo não existia -- cria do zero (DE recarregado: uma
			// chamada BDOS anterior não deixa garantia sobre seu valor).
			sb.WriteString(fmt.Sprintf("    LD DE, %s\n", pathLabel))
			sb.WriteString("    LD A, 00h\n")
			sb.WriteString("    LD B, 00h\n") // atributo normal
			sb.WriteString("    CALL BDOS_FileCreate\n")
			sb.WriteString(fmt.Sprintf("%s:\n", openedLbl))
			sb.WriteString("    LD A, B\n") // handle retornado em B
			sb.WriteString(fmt.Sprintf("    LD (%s), A\n", handleLabel))
			// Posiciona o ponteiro no fim do arquivo (método 2, offset 0)
			// antes de qualquer PRINT #n escrever nele.
			sb.WriteString("    LD B, A\n")
			sb.WriteString("    LD A, 02h\n") // 2 = a partir do fim
			sb.WriteString("    LD DE, 0000h\n")
			sb.WriteString("    LD HL, 0000h\n")
			sb.WriteString("    CALL BDOS_FileSeek\n")

		default: // OUTPUT
			cg.needsFileCreate = true
			sb.WriteString(fmt.Sprintf("    LD DE, %s\n", pathLabel))
			sb.WriteString("    LD A, 00h\n") // leitura+escrita
			sb.WriteString("    LD B, 00h\n") // atributo normal (cria ou trunca)
			sb.WriteString("    CALL BDOS_FileCreate\n")
			sb.WriteString("    LD A, B\n") // handle retornado em B
			sb.WriteString(fmt.Sprintf("    LD (%s), A\n", handleLabel))
		}
		return nil

	case *CloseStmt:
		cg.needsFileClose = true
		cg.fileNums[s.FileNum] = true
		handleLabel := fmt.Sprintf("DGN_FileHandle_%d", s.FileNum)
		sb.WriteString(fmt.Sprintf("    LD A, (%s)\n", handleLabel))
		sb.WriteString("    LD B, A\n")
		sb.WriteString("    CALL BDOS_FileClose\n")
		return nil

	case *ClsStmt:
		cg.needsCls = true
		sb.WriteString("    CALL BIOS_CLS\n")
		return nil

	case *BeepStmt:
		cg.needsBeep = true
		sb.WriteString("    CALL BIOS_BEEP\n")
		return nil

	case *ScreenStmt:
		cg.needsChgMod = true
		if err := cg.generateExpr(sb, s.Mode); err != nil {
			return err
		}
		sb.WriteString("    LD A, L\n")
		sb.WriteString("    CALL BIOS_CHGMOD\n")
		return nil

	case *ReturnStmt:
		if s.Value != nil {
			if err := cg.generateExpr(sb, s.Value); err != nil {
				return err
			}
		}
		sb.WriteString("    LD SP, IX\n")
		sb.WriteString("    POP IX\n")
		sb.WriteString("    RET\n")
		return nil

	default:
		return fmt.Errorf("instrução não suportada pelo gerador de código: %T", stmt)
	}
}

// generatePrintFileStmt gera PRINT #n, expr[, expr...] -- mesma gramática de
// argumentos do PRINT de console, mas escrevendo em arquivo via
// BDOS_FileWrite em vez de BDOS_PrintString/PrintDec16 (que só sabem
// escrever no console). Literais de string vão direto (sem o terminador '$'
// que só o console precisa); expressões numéricas passam por
// PrintDec16ToBuffer antes de ir pro arquivo.
func (cg *CodeGenerator) generatePrintFileStmt(sb *strings.Builder, s *PrintStmt) error {
	n := *s.FileNum
	cg.fileNums[n] = true
	handleLabel := fmt.Sprintf("DGN_FileHandle_%d", n)

	writeBuf := func(bufExpr string, sizeExpr string) {
		sb.WriteString(fmt.Sprintf("    LD A, (%s)\n", handleLabel))
		sb.WriteString("    LD B, A\n")
		sb.WriteString(fmt.Sprintf("    LD DE, %s\n", bufExpr))
		sb.WriteString(fmt.Sprintf("    LD HL, %s\n", sizeExpr))
		sb.WriteString("    CALL BDOS_FileWrite\n")
	}

	cg.needsFileWrite = true
	for _, arg := range s.Args {
		switch a := arg.(type) {
		case *StringExpr:
			lbl := cg.newLabel("FileStr")
			cg.extraData = append(cg.extraData, fmt.Sprintf("%s:\n    DB \"%s\"\n", lbl, escapeString(a.Value)))
			writeBuf(lbl, fmt.Sprintf("%d", len(a.Value)))
		default:
			cg.needsPrintDecBuf = true
			if err := cg.generateExpr(sb, a); err != nil {
				return err
			}
			sb.WriteString("    LD DE, DGN_FileNumBuf\n")
			sb.WriteString("    CALL PrintDec16ToBuffer\n")
			// A = tamanho (retorno de PrintDec16ToBuffer); HL precisa desse
			// mesmo valor pra BDOS_FileWrite, então monta HL a partir de A
			// em vez de usar o helper writeBuf (que só aceita um tamanho
			// fixo conhecido em tempo de compilação, não um valor de retorno).
			sb.WriteString("    LD L, A\n    LD H, 0\n")
			sb.WriteString(fmt.Sprintf("    LD A, (%s)\n    LD B, A\n", handleLabel))
			sb.WriteString("    LD DE, DGN_FileNumBuf\n")
			sb.WriteString("    CALL BDOS_FileWrite\n")
		}
	}

	if !s.TrailingSemicolon {
		writeBuf("DGN_CRLF", "2")
	}

	return nil
}

func (cg *CodeGenerator) generateExpr(sb *strings.Builder, expr Expr) error {
	switch e := expr.(type) {
	case *NumberExpr:
		sb.WriteString(fmt.Sprintf("    LD HL, %04Xh\n", uint16(e.Value)))
		return nil

	case *StringExpr:
		lbl := cg.getStringLabel(e.Value)
		sb.WriteString(fmt.Sprintf("    LD HL, %s\n", lbl))
		return nil

	case *FloatExpr:
		return fmt.Errorf("aritmética de ponto flutuante (SINGLE/DOUBLE) ainda não implementada -- literal %v", e.Value)

	case *VarExpr:
		switch cg.varType(e.Name) {
		case "STRING":
			return fmt.Errorf("variável STRING '%s' só pode ser usada em atribuição ou PRINT por enquanto", e.Name)
		case "SINGLE", "DOUBLE":
			return fmt.Errorf("aritmética de ponto flutuante (SINGLE/DOUBLE) ainda não implementada -- variável '%s'", e.Name)
		default:
			return cg.loadVar(sb, e.Name)
		}

	case *UnaryExpr:
		if err := cg.generateExpr(sb, e.Expr); err != nil {
			return err
		}
		switch e.Op {
		case TokenMinus:
			// HL = -HL
			sb.WriteString("    EX DE, HL\n")
			sb.WriteString("    LD HL, 0000h\n")
			sb.WriteString("    OR A\n")
			sb.WriteString("    SBC HL, DE\n")
		case TokenNot:
			// Inverte boolean / bitwise NOT
			sb.WriteString("    LD A, H\n")
			sb.WriteString("    CPL\n")
			sb.WriteString("    LD H, A\n")
			sb.WriteString("    LD A, L\n")
			sb.WriteString("    CPL\n")
			sb.WriteString("    LD L, A\n")
		}
		return nil

	case *CallExpr:
		// Chamada de função como expressão
		for _, arg := range e.Args {
			if err := cg.generateExpr(sb, arg); err != nil {
				return err
			}
			sb.WriteString("    PUSH HL\n")
		}
		sb.WriteString(fmt.Sprintf("    CALL %s\n", e.Name))
		if len(e.Args) > 0 {
			sb.WriteString(fmt.Sprintf("    LD DE, %d\n", len(e.Args)*2))
			sb.WriteString("    ADD HL, DE\n") // preserva HL?
			// Melhor restaurar SP com EXX ou registrador auxiliar
			sb.WriteString("    EX DE, HL\n")
			sb.WriteString("    ADD HL, SP\n")
			sb.WriteString("    LD SP, HL\n")
			sb.WriteString("    EX DE, HL\n") // HL tem o retorno
		}
		return nil

	case *BinaryExpr:
		// Avalia lado esquerdo
		if err := cg.generateExpr(sb, e.Left); err != nil {
			return err
		}
		sb.WriteString("    PUSH HL\n")

		// Avalia lado direito
		if err := cg.generateExpr(sb, e.Right); err != nil {
			return err
		}
		sb.WriteString("    EX DE, HL\n") // DE = Right
		sb.WriteString("    POP HL\n")    // HL = Left

		switch e.Op {
		case TokenPlus:
			sb.WriteString("    ADD HL, DE\n")
		case TokenMinus:
			sb.WriteString("    OR A\n")
			sb.WriteString("    SBC HL, DE\n")
		case TokenMul:
			cg.needsMul16 = true
			sb.WriteString("    CALL Mul16\n")
		case TokenDiv:
			cg.needsDiv16 = true
			sb.WriteString("    CALL Div16\n") // HL = quociente
		case TokenMod:
			cg.needsDiv16 = true
			sb.WriteString("    CALL Div16\n") // HL = quociente, DE = resto
			sb.WriteString("    EX DE, HL\n")  // HL = resto
		case TokenAnd:
			sb.WriteString("    LD A, H\n    AND D\n    LD H, A\n")
			sb.WriteString("    LD A, L\n    AND E\n    LD L, A\n")
		case TokenOr:
			sb.WriteString("    LD A, H\n    OR D\n     LD H, A\n")
			sb.WriteString("    LD A, L\n    OR E\n     LD L, A\n")
		case TokenXor:
			sb.WriteString("    LD A, H\n    XOR D\n    LD H, A\n")
			sb.WriteString("    LD A, L\n    XOR E\n    LD L, A\n")

		case TokenEqual, TokenNotEqual, TokenLess, TokenLessEq, TokenGreater, TokenGreaterEq:
			trueLbl := cg.newLabel("Rel_True")
			endLbl := cg.newLabel("Rel_End")

			sb.WriteString("    OR A\n")
			sb.WriteString("    SBC HL, DE\n")

			switch e.Op {
			case TokenEqual:
				sb.WriteString(fmt.Sprintf("    JP Z, %s\n", trueLbl))
			case TokenNotEqual:
				sb.WriteString(fmt.Sprintf("    JP NZ, %s\n", trueLbl))
			case TokenLess:
				sb.WriteString(fmt.Sprintf("    JP C, %s\n", trueLbl))
			case TokenLessEq:
				sb.WriteString(fmt.Sprintf("    JP C, %s\n", trueLbl))
				sb.WriteString(fmt.Sprintf("    JP Z, %s\n", trueLbl))
			case TokenGreater:
				sb.WriteString(fmt.Sprintf("    JP Z, %s\n", endLbl))
				sb.WriteString(fmt.Sprintf("    JP NC, %s\n", trueLbl))
			case TokenGreaterEq:
				sb.WriteString(fmt.Sprintf("    JP NC, %s\n", trueLbl))
			}

			// Falso: HL = 0000h
			sb.WriteString("    LD HL, 0000h\n")
			sb.WriteString(fmt.Sprintf("    JP %s\n", endLbl))
			// Verdadeiro: HL = 0001h
			sb.WriteString(fmt.Sprintf("%s:\n", trueLbl))
			sb.WriteString("    LD HL, 0001h\n")
			sb.WriteString(fmt.Sprintf("%s:\n", endLbl))

		default:
			return fmt.Errorf("operador binário não suportado: %v", e.Op)
		}
		return nil

	default:
		return fmt.Errorf("expressão não suportada: %T", expr)
	}
}

func (cg *CodeGenerator) loadVar(sb *strings.Builder, name string) error {
	lower := strings.ToLower(name)

	// 1. Variável local: (IX - Offset)
	if sym, ok := cg.currentLocals[lower]; ok {
		sb.WriteString(fmt.Sprintf("    LD L, (IX - %d)\n", sym.Offset))
		sb.WriteString(fmt.Sprintf("    LD H, (IX - %d)\n", sym.Offset-1))
		return nil
	}

	// 2. Parâmetro: (IX + Offset)
	if sym, ok := cg.currentParams[lower]; ok {
		sb.WriteString(fmt.Sprintf("    LD L, (IX + %d)\n", sym.Offset))
		sb.WriteString(fmt.Sprintf("    LD H, (IX + %d)\n", sym.Offset+1))
		return nil
	}

	// 3. Global
	if sym, ok := cg.globals[lower]; ok {
		sb.WriteString(fmt.Sprintf("    LD HL, (%s)\n", sym.Label))
		return nil
	}

	return fmt.Errorf("variável '%s' não declarada", name)
}

func (cg *CodeGenerator) storeVar(sb *strings.Builder, name string) error {
	lower := strings.ToLower(name)

	// 1. Variável local: (IX - Offset)
	if sym, ok := cg.currentLocals[lower]; ok {
		sb.WriteString(fmt.Sprintf("    LD (IX - %d), L\n", sym.Offset))
		sb.WriteString(fmt.Sprintf("    LD (IX - %d), H\n", sym.Offset-1))
		return nil
	}

	// 2. Parâmetro: (IX + Offset)
	if sym, ok := cg.currentParams[lower]; ok {
		sb.WriteString(fmt.Sprintf("    LD (IX + %d), L\n", sym.Offset))
		sb.WriteString(fmt.Sprintf("    LD (IX + %d), H\n", sym.Offset+1))
		return nil
	}

	// 3. Global
	if sym, ok := cg.globals[lower]; ok {
		sb.WriteString(fmt.Sprintf("    LD (%s), HL\n", sym.Label))
		return nil
	}

	return fmt.Errorf("variável '%s' não declarada", name)
}

// loadVarAddress carrega em HL o ENDEREÇO de uma variável (não o valor) --
// necessário para tipos que não cabem num par de registradores de 16 bits
// (STRING: 256 bytes; SINGLE: 4; DOUBLE: 8), diferente de loadVar/storeVar,
// que sempre movem exatamente 2 bytes.
//
// Para uma global, o endereço já É o label (LD HL,label). Para uma local ou
// parâmetro, o endereço é relativo a IX (IX-Offset ou IX+Offset) -- não
// existe "ADD HL,IX" no Z80 (só ADD HL com BC/DE/HL/SP), então o caminho é
// copiar IX pra HL via PUSH IX/POP HL e somar o deslocamento com ADD HL,DE.
// O deslocamento negativo (caso local) é calculado em Go como um uint16 em
// complemento de dois e emitido já pronto em hexadecimal -- mesmo padrão que
// generateExpr já usa pra NumberExpr, em vez de depender do parser do KAJI80
// reconhecer um literal decimal negativo (que não foi verificado).
func (cg *CodeGenerator) loadVarAddress(sb *strings.Builder, name string) error {
	lower := strings.ToLower(name)

	if sym, ok := cg.currentLocals[lower]; ok {
		sb.WriteString("    PUSH IX\n    POP HL\n")
		sb.WriteString(fmt.Sprintf("    LD DE, %04Xh\n", uint16(-sym.Offset)))
		sb.WriteString("    ADD HL, DE\n")
		return nil
	}

	if sym, ok := cg.currentParams[lower]; ok {
		sb.WriteString("    PUSH IX\n    POP HL\n")
		sb.WriteString(fmt.Sprintf("    LD DE, %04Xh\n", uint16(sym.Offset)))
		sb.WriteString("    ADD HL, DE\n")
		return nil
	}

	if sym, ok := cg.globals[lower]; ok {
		sb.WriteString(fmt.Sprintf("    LD HL, %s\n", sym.Label))
		return nil
	}

	return fmt.Errorf("variável '%s' não declarada", name)
}

// varType devolve o tipo declarado de uma variável (local, parâmetro ou
// global) já conhecida, ou "" se não encontrada.
func (cg *CodeGenerator) varType(name string) string {
	lower := strings.ToLower(name)
	if sym, ok := cg.currentLocals[lower]; ok {
		return sym.Type
	}
	if sym, ok := cg.currentParams[lower]; ok {
		return sym.Type
	}
	if sym, ok := cg.globals[lower]; ok {
		return sym.Type
	}
	return ""
}

// generateStringAssign implementa "s$ = <literal>" ou "s$ = outravar$" --
// diferente de uma atribuição INTEGER/BOOLEAN (LD (dest),HL de 2 bytes fixos
// via storeVar), uma STRING é um buffer de até 256 bytes (1 de tamanho + até
// 255 de dados, SPEC.md §7), então a atribuição é uma cópia de buffer via
// StrCopyLen (MSXLIB, lib/src/string.asm) entre os ENDEREÇOS de origem e
// destino, não um valor de 16 bits.
//
// O endereço de destino é calculado e empilhado ANTES do de origem porque
// loadVarAddress usa DE como registrador de rascunho para locais/parâmetros
// -- se a origem também for uma local/parâmetro e fosse calculada primeiro,
// calcular o destino depois sobrescreveria o DE da origem antes do CALL.
func (cg *CodeGenerator) generateStringAssign(sb *strings.Builder, s *AssignStmt) error {
	cg.needsStrCopyLen = true

	if err := cg.loadVarAddress(sb, s.VarName); err != nil {
		return err
	}
	sb.WriteString("    PUSH HL\n")

	switch v := s.Value.(type) {
	case *StringExpr:
		if len(v.Value) > 255 {
			return fmt.Errorf("literal de string \"%s\" excede 255 caracteres (máximo suportado)", v.Value)
		}
		label := cg.newLabel("StrLit")
		cg.extraData = append(cg.extraData, fmt.Sprintf("%s:\n    DB %02Xh, \"%s\"\n", label, len(v.Value), escapeString(v.Value)))
		sb.WriteString(fmt.Sprintf("    LD HL, %s\n", label))
	case *VarExpr:
		srcType := cg.varType(v.Name)
		if srcType != "STRING" {
			return fmt.Errorf("atribuição a '%s' (STRING) exige um literal de texto ou outra variável STRING -- '%s' é do tipo %s", s.VarName, v.Name, srcType)
		}
		if err := cg.loadVarAddress(sb, v.Name); err != nil {
			return err
		}
	default:
		return fmt.Errorf("atribuição a '%s' (STRING) só suporta literal de texto ou outra variável STRING por enquanto -- concatenação e outras expressões de string ainda não implementadas", s.VarName)
	}

	sb.WriteString("    POP DE\n")
	sb.WriteString("    CALL StrCopyLen\n")
	return nil
}

// generateFloatAssign implementa "x! = <literal>" ou "x! = outravar!" para
// SINGLE/DOUBLE -- cópia crua de 4/8 bytes IEEE754 entre endereços (mesma
// ordem destino-antes-de-origem de generateStringAssign, mesmo motivo). Não
// existe LDIR no KAJI80 (não suportado pelo assembler), e o tamanho é uma
// constante pequena conhecida em tempo de compilação (4 ou 8), então o loop
// é desenrolado em vez de usar B/DJNZ.
func (cg *CodeGenerator) generateFloatAssign(sb *strings.Builder, s *AssignStmt, targetType string) error {
	if err := cg.loadVarAddress(sb, s.VarName); err != nil {
		return err
	}
	sb.WriteString("    PUSH HL\n")

	switch v := s.Value.(type) {
	case *FloatExpr:
		sb.WriteString(fmt.Sprintf("    LD HL, %s\n", cg.floatLiteralLabel(v, targetType)))
	case *NumberExpr:
		fe := &FloatExpr{Value: float64(v.Value), IsDouble: targetType == "DOUBLE"}
		sb.WriteString(fmt.Sprintf("    LD HL, %s\n", cg.floatLiteralLabel(fe, targetType)))
	case *VarExpr:
		srcType := cg.varType(v.Name)
		if srcType != targetType {
			return fmt.Errorf("atribuição a '%s' (%s) exige um literal numérico ou outra variável %s -- '%s' é do tipo %s", s.VarName, targetType, targetType, v.Name, srcType)
		}
		if err := cg.loadVarAddress(sb, v.Name); err != nil {
			return err
		}
	default:
		return fmt.Errorf("atribuição a '%s' (%s) só suporta literal numérico ou outra variável %s por enquanto -- aritmética de ponto flutuante ainda não implementada", s.VarName, targetType, targetType)
	}

	sb.WriteString("    POP DE\n")
	cg.emitRawCopy(sb, typeSize(targetType))
	return nil
}

// floatLiteralLabel emite um bloco DB com a representação IEEE754
// (binary32/binary64, little-endian) de um literal SINGLE/DOUBLE calculada
// em Go (math.Float32bits/Float64bits) e devolve o label criado.
func (cg *CodeGenerator) floatLiteralLabel(v *FloatExpr, targetType string) string {
	if targetType == "DOUBLE" {
		var buf [8]byte
		binary.LittleEndian.PutUint64(buf[:], math.Float64bits(v.Value))
		label := cg.newLabel("DblLit")
		cg.extraData = append(cg.extraData, fmt.Sprintf("%s:\n    DB %s\n", label, bytesToHexList(buf[:])))
		return label
	}
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], math.Float32bits(float32(v.Value)))
	label := cg.newLabel("SglLit")
	cg.extraData = append(cg.extraData, fmt.Sprintf("%s:\n    DB %s\n", label, bytesToHexList(buf[:])))
	return label
}

// emitRawCopy copia `size` bytes de (HL) para (DE), desenrolado (sem laço) --
// usado só para SINGLE/DOUBLE, onde size é sempre 4 ou 8.
func (cg *CodeGenerator) emitRawCopy(sb *strings.Builder, size int) {
	for i := 0; i < size; i++ {
		sb.WriteString("    LD A, (HL)\n")
		sb.WriteString("    LD (DE), A\n")
		if i < size-1 {
			sb.WriteString("    INC HL\n")
			sb.WriteString("    INC DE\n")
		}
	}
}

func bytesToHexList(bs []byte) string {
	parts := make([]string, len(bs))
	for i, b := range bs {
		parts[i] = fmt.Sprintf("%02Xh", b)
	}
	return strings.Join(parts, ", ")
}

func sanitizeIdent(s string) string {
	s = strings.ReplaceAll(s, "%", "_int")
	s = strings.ReplaceAll(s, "$", "_str")
	s = strings.ReplaceAll(s, "!", "_sgl") // ! agora é SINGLE, não BOOLEAN
	s = strings.ReplaceAll(s, "#", "_dbl") // # (DOUBLE) não era tratado antes -- um label com '#' de verdade não é um identificador Assembly válido
	return s
}

func containsString(list []string, item string) bool {
	for _, s := range list {
		if strings.EqualFold(s, item) {
			return true
		}
	}
	return false
}

func escapeString(s string) string {
	var sb strings.Builder
	for i := 0; i < len(s); i++ {
		b := s[i]
		if b == '"' {
			sb.WriteString(`\"`)
		} else {
			sb.WriteByte(b)
		}
	}
	return sb.String()
}
