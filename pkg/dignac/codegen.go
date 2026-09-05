package dignac

import (
	"fmt"
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
	Type   string // "INTEGER", "STRING", "BOOLEAN"
	Offset int    // offset a partir de IX (positivo para param, positivo absoluto para local onde local está em IX - Offset)
	Label  string // label para globais
}

// CodeGenerator traduz a AST de MSX-BASIC Dignified em Assembly Z80 compatível com KAJI80
type CodeGenerator struct {
	module        *ModuleNode
	asm           strings.Builder
	labelCounter  int
	stringLits    map[string]string // texto -> label (ex: "StrLit_1")
	globals       map[string]*varSymbol
	currentLocals map[string]*varSymbol
	currentParams map[string]*varSymbol
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
}

// NewCodeGenerator cria um novo gerador de código
func NewCodeGenerator(module *ModuleNode) *CodeGenerator {
	return &CodeGenerator{
		module:     module,
		stringLits: make(map[string]string),
		globals:    make(map[string]*varSymbol),
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
		for _, vName := range g.Vars {
			lower := strings.ToLower(vName)
			cg.globals[lower] = &varSymbol{
				Name:  vName,
				Kind:  varGlobal,
				Type:  g.Type,
				Label: fmt.Sprintf("Global_%s", sanitizeIdent(vName)),
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

	// Seção de Variáveis Globais
	if len(cg.globals) > 0 {
		cg.asm.WriteString("; --- Variáveis Globais ---\n")
		for _, sym := range cg.globals {
			if strings.EqualFold(sym.Type, "BOOLEAN") {
				cg.asm.WriteString(fmt.Sprintf("%s:\n    DB 00h\n", sym.Label))
			} else {
				// Integer (16 bits)
				cg.asm.WriteString(fmt.Sprintf("%s:\n    DW 0000h\n", sym.Label))
			}
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

	// Alocação de variáveis locais no frame de pilha (IX - Offset)
	localOffset := 0
	for _, locDecl := range proc.Locals {
		for _, vName := range locDecl.Vars {
			localOffset += 2 // 16-bit
			cg.currentLocals[strings.ToLower(vName)] = &varSymbol{
				Name:   vName,
				Kind:   varLocal,
				Type:   locDecl.Type,
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
		// Avalia expressão -> HL
		if err := cg.generateExpr(sb, s.Value); err != nil {
			return err
		}
		// Armazena HL na variável
		return cg.storeVar(sb, s.VarName)

	case *ForStmt:
		// FOR var = start TO end [STEP step] ... NEXT [var]
		loopStart := cg.newLabel("For_Start")
		loopEnd := cg.newLabel("For_End")

		// 1. Inicializa variável com Start
		if err := cg.generateExpr(sb, s.Start); err != nil {
			return err
		}
		if err := cg.storeVar(sb, s.VarName); err != nil {
			return err
		}

		sb.WriteString(fmt.Sprintf("%s:\n", loopStart))

		// 2. Condição de término: var <= end (para step positivo)
		if err := cg.loadVar(sb, s.VarName); err != nil {
			return err
		}
		sb.WriteString("    PUSH HL\n")
		if err := cg.generateExpr(sb, s.End); err != nil {
			return err
		}
		sb.WriteString("    EX DE, HL\n") // DE = End
		sb.WriteString("    POP HL\n")    // HL = Var
		// Compara Var com End: se Var > End -> encerra
		sb.WriteString("    OR A\n")
		sb.WriteString("    SBC HL, DE\n")
		okLbl := cg.newLabel("For_Ok")
		sb.WriteString(fmt.Sprintf("    JP Z, %s\n", okLbl))
		sb.WriteString(fmt.Sprintf("    JP NC, %s\n", loopEnd))
		sb.WriteString(fmt.Sprintf("%s:\n", okLbl))

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
		sb.WriteString("    LD A, L\n")    // A = Color
		sb.WriteString("    POP DE\n")     // DE = Y
		sb.WriteString("    POP BC\n")     // BC = X
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
		for _, arg := range s.Args {
			switch a := arg.(type) {
			case *StringExpr:
				cg.needsPrintStr = true
				lbl := cg.getStringLabel(a.Value)
				sb.WriteString(fmt.Sprintf("    LD DE, %s\n", lbl))
				sb.WriteString("    CALL BDOS_PrintString\n")
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

func (cg *CodeGenerator) generateExpr(sb *strings.Builder, expr Expr) error {
	switch e := expr.(type) {
	case *NumberExpr:
		sb.WriteString(fmt.Sprintf("    LD HL, %04Xh\n", uint16(e.Value)))
		return nil

	case *StringExpr:
		lbl := cg.getStringLabel(e.Value)
		sb.WriteString(fmt.Sprintf("    LD HL, %s\n", lbl))
		return nil

	case *VarExpr:
		return cg.loadVar(sb, e.Name)

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

func sanitizeIdent(s string) string {
	s = strings.ReplaceAll(s, "%", "_int")
	s = strings.ReplaceAll(s, "$", "_str")
	s = strings.ReplaceAll(s, "!", "_bool")
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
