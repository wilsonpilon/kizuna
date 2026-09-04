package wirth80

import (
	"fmt"
	"strings"

	"github.com/wilsonpilon/kizuna/pkg/kaji80"
	"github.com/wilsonpilon/kizuna/pkg/mob"
)

// CodeGenerator gera código Assembly Z80 a partir da AST do Pascal
type CodeGenerator struct {
	prog          *ProgramNode
	asm           strings.Builder
	labelCounter  int
	stringLits    map[string]string // texto -> label (ex: "StrLit_1")
	varMap        map[string]string // nome da variável (case-insensitive) -> tipo
	varLabels     map[string]string // nome da variável -> label (ex: "Var_a")
	needsMul16    bool
	needsDiv16    bool
	needsPrintDec bool
	needsPrintStr bool
	needsPrintChr bool
}

// NewCodeGenerator cria um novo gerador de código
func NewCodeGenerator(prog *ProgramNode) *CodeGenerator {
	return &CodeGenerator{
		prog:       prog,
		stringLits: make(map[string]string),
		varMap:     make(map[string]string),
		varLabels:  make(map[string]string),
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

// GenerateAsm emite o código Assembly Z80 formatado
func (cg *CodeGenerator) GenerateAsm() (string, error) {
	cg.asm.Reset()

	modName := cg.prog.Name
	if modName == "" {
		modName = "Program"
	}

	cg.asm.WriteString(fmt.Sprintf("; =============================================================================\n"))
	cg.asm.WriteString(fmt.Sprintf("; Código gerado pelo compilador Pascal WIRTH80 - Kizuna Toolchain\n"))
	cg.asm.WriteString(fmt.Sprintf("; Programa: %s\n", modName))
	cg.asm.WriteString(fmt.Sprintf("; =============================================================================\n\n"))

	cg.asm.WriteString(fmt.Sprintf("MODULE %s\n", modName))
	cg.asm.WriteString("BANK 0\n\n")
	cg.asm.WriteString("PUBLIC Start\n\n")

	// Coleta variáveis
	for _, decl := range cg.prog.Vars {
		for _, name := range decl.Names {
			upper := strings.ToUpper(name)
			cg.varMap[upper] = decl.Type
			cg.varLabels[upper] = fmt.Sprintf("Var_%s", name)
		}
	}

	// Buffer temporário para o corpo do código
	var body strings.Builder

	// Gerar instruções do bloco principal
	if cg.prog.Block != nil {
		for _, stmt := range cg.prog.Block.Statements {
			if err := cg.generateStmt(&body, stmt); err != nil {
				return "", err
			}
		}
	}

	// Emitir lista de imports EXTERN conforme a necessidade
	var externs []string
	if cg.needsPrintStr {
		externs = append(externs, "BDOS_PrintString")
	}
	if cg.needsPrintChr {
		externs = append(externs, "BDOS_PrintChar")
	}
	if cg.needsPrintDec {
		externs = append(externs, "PrintDec16")
	}
	if cg.needsMul16 {
		externs = append(externs, "Mul16")
	}
	if cg.needsDiv16 {
		externs = append(externs, "Div16")
	}
	externs = append(externs, "BDOS_Exit")

	cg.asm.WriteString(fmt.Sprintf("EXTERN %s\n\n", strings.Join(externs, ", ")))

	// Ponto de entrada
	cg.asm.WriteString("Start:\n")
	cg.asm.WriteString(body.String())
	cg.asm.WriteString("    CALL BDOS_Exit\n\n")

	// Seção de Dados: Literais de String
	if len(cg.stringLits) > 0 {
		cg.asm.WriteString("; --- Literais de String ---\n")
		for str, lbl := range cg.stringLits {
			// Escapar para formato DB compatível com BDOS 09h ($)
			cg.asm.WriteString(fmt.Sprintf("%s:\n", lbl))
			cg.asm.WriteString(fmt.Sprintf("    DB \"%s$\"\n", escapeString(str)))
		}
		cg.asm.WriteString("\n")
	}

	// Seção de Variáveis Globais
	if len(cg.varLabels) > 0 {
		cg.asm.WriteString("; --- Variáveis Globais ---\n")
		for name, lbl := range cg.varLabels {
			t := strings.ToUpper(cg.varMap[name])
			if t == "CHAR" || t == "BOOLEAN" {
				cg.asm.WriteString(fmt.Sprintf("%s:\n    DB 00h\n", lbl))
			} else {
				// Integer (16 bits)
				cg.asm.WriteString(fmt.Sprintf("%s:\n    DW 0000h\n", lbl))
			}
		}
		cg.asm.WriteString("\n")
	}

	cg.asm.WriteString("ENDMOD\n")
	return cg.asm.String(), nil
}

// Compile compila o código Pascal diretamente para um ObjectFile .MOB
func (cg *CodeGenerator) Compile() (*mob.ObjectFile, string, error) {
	asmSource, err := cg.GenerateAsm()
	if err != nil {
		return nil, "", err
	}

	asm := kaji80.NewAssembler()
	obj, err := asm.Assemble(asmSource)
	if err != nil {
		return nil, asmSource, fmt.Errorf("erro na montagem do assembly Z80 gerado: %w\nCódigo gerado:\n%s", err, asmSource)
	}

	return obj, asmSource, nil
}

func (cg *CodeGenerator) generateStmt(sb *strings.Builder, stmt Stmt) error {
	switch s := stmt.(type) {
	case *BlockStmt:
		for _, child := range s.Statements {
			if err := cg.generateStmt(sb, child); err != nil {
				return err
			}
		}
		return nil

	case *AssignStmt:
		// var := expr
		upper := strings.ToUpper(s.VarName)
		lbl, ok := cg.varLabels[upper]
		if !ok {
			return fmt.Errorf("variável não declarada '%s' na linha %d:%d", s.VarName, s.Line, s.Column)
		}

		if err := cg.generateExpr(sb, s.Expr); err != nil {
			return err
		}

		t := strings.ToUpper(cg.varMap[upper])
		if t == "CHAR" || t == "BOOLEAN" {
			sb.WriteString(fmt.Sprintf("    LD A, L\n"))
			sb.WriteString(fmt.Sprintf("    LD (%s), A\n", lbl))
		} else {
			sb.WriteString(fmt.Sprintf("    LD (%s), HL\n", lbl))
		}
		return nil

	case *WriteStmt:
		for _, arg := range s.Args {
			if strLit, ok := arg.(*StringLiteral); ok {
				// Impressão direta de string literal
				lbl := cg.getStringLabel(strLit.Value)
				cg.needsPrintStr = true
				sb.WriteString(fmt.Sprintf("    LD DE, %s\n", lbl))
				sb.WriteString(fmt.Sprintf("    CALL BDOS_PrintString\n"))
			} else {
				// Expressão inteira: avalia em HL e chama PrintDec16
				if err := cg.generateExpr(sb, arg); err != nil {
					return err
				}
				cg.needsPrintDec = true
				sb.WriteString(fmt.Sprintf("    CALL PrintDec16\n"))
			}
		}

		if s.NewLine {
			cg.needsPrintChr = true
			sb.WriteString("    LD E, 0Dh\n")
			sb.WriteString("    CALL BDOS_PrintChar\n")
			sb.WriteString("    LD E, 0Ah\n")
			sb.WriteString("    CALL BDOS_PrintChar\n")
		}
		return nil

	case *IfStmt:
		elseLabel := cg.newLabel("Else")
		endLabel := cg.newLabel("EndIf")

		// Avalia condição em HL (0 = falso, != 0 = verdadeiro)
		if err := cg.generateExpr(sb, s.Cond); err != nil {
			return err
		}

		sb.WriteString("    LD A, H\n")
		sb.WriteString("    OR L\n")
		if s.Else != nil {
			sb.WriteString(fmt.Sprintf("    JP Z, %s\n", elseLabel))
			if err := cg.generateStmt(sb, s.Then); err != nil {
				return err
			}
			sb.WriteString(fmt.Sprintf("    JP %s\n", endLabel))
			sb.WriteString(fmt.Sprintf("%s:\n", elseLabel))
			if err := cg.generateStmt(sb, s.Else); err != nil {
				return err
			}
			sb.WriteString(fmt.Sprintf("%s:\n", endLabel))
		} else {
			sb.WriteString(fmt.Sprintf("    JP Z, %s\n", endLabel))
			if err := cg.generateStmt(sb, s.Then); err != nil {
				return err
			}
			sb.WriteString(fmt.Sprintf("%s:\n", endLabel))
		}
		return nil

	case *WhileStmt:
		loopLabel := cg.newLabel("WhileLoop")
		endLabel := cg.newLabel("WhileEnd")

		sb.WriteString(fmt.Sprintf("%s:\n", loopLabel))
		if err := cg.generateExpr(sb, s.Cond); err != nil {
			return err
		}
		sb.WriteString("    LD A, H\n")
		sb.WriteString("    OR L\n")
		sb.WriteString(fmt.Sprintf("    JP Z, %s\n", endLabel))

		if err := cg.generateStmt(sb, s.Body); err != nil {
			return err
		}
		sb.WriteString(fmt.Sprintf("    JP %s\n", loopLabel))
		sb.WriteString(fmt.Sprintf("%s:\n", endLabel))
		return nil

	default:
		return fmt.Errorf("tipo de comando não suportado: %T", stmt)
	}
}

// generateExpr avalia a expressão e deixa o resultado em HL (16 bits)
func (cg *CodeGenerator) generateExpr(sb *strings.Builder, expr Expr) error {
	switch e := expr.(type) {
	case *NumberLiteral:
		sb.WriteString(fmt.Sprintf("    LD HL, %04Xh ; %d\n", uint16(e.Value), e.Value))
		return nil

	case *VarExpr:
		upper := strings.ToUpper(e.Name)
		lbl, ok := cg.varLabels[upper]
		if !ok {
			return fmt.Errorf("variável não declarada '%s' na linha %d:%d", e.Name, e.Line, e.Column)
		}
		t := strings.ToUpper(cg.varMap[upper])
		if t == "CHAR" || t == "BOOLEAN" {
			sb.WriteString(fmt.Sprintf("    LD A, (%s)\n", lbl))
			sb.WriteString(fmt.Sprintf("    LD L, A\n    LD H, 00h\n"))
		} else {
			sb.WriteString(fmt.Sprintf("    LD HL, (%s)\n", lbl))
		}
		return nil

	case *UnaryExpr:
		if err := cg.generateExpr(sb, e.Expr); err != nil {
			return err
		}
		if e.Op == TokenMinus {
			// HL = -HL: EX DE, HL; LD HL, 0; OR A; SBC HL, DE
			sb.WriteString("    EX DE, HL\n")
			sb.WriteString("    LD HL, 0000h\n")
			sb.WriteString("    OR A\n")
			sb.WriteString("    SBC HL, DE\n")
		}
		return nil

	case *BinaryExpr:
		// 1. Avalia lado esquerdo em HL
		if err := cg.generateExpr(sb, e.Left); err != nil {
			return err
		}
		// 2. Empilha lado esquerdo
		sb.WriteString("    PUSH HL\n")
		// 3. Avalia lado direito em HL
		if err := cg.generateExpr(sb, e.Right); err != nil {
			return err
		}
		// 4. Desempilha lado esquerdo em DE (DE = Left, HL = Right)
		sb.WriteString("    POP DE\n")

		switch e.Op {
		case TokenPlus:
			// HL = DE + HL
			sb.WriteString("    ADD HL, DE\n")

		case TokenMinus:
			// HL = DE - HL
			sb.WriteString("    EX DE, HL\n") // HL = Left, DE = Right
			sb.WriteString("    OR A\n")
			sb.WriteString("    SBC HL, DE\n")

		case TokenMul:
			// HL = Left * Right (Mul16)
			cg.needsMul16 = true
			sb.WriteString("    CALL Mul16\n")

		case TokenDiv:
			// HL = Left / Right (Div16)
			cg.needsDiv16 = true
			sb.WriteString("    EX DE, HL\n") // HL = Dividendo (Left), DE = Divisor (Right)
			sb.WriteString("    CALL Div16\n")

		case TokenEqual, TokenNotEqual, TokenLess, TokenLessEq, TokenGreater, TokenGreaterEq:
			// Comparações: DE (Left) vs HL (Right) -> resultado booleano em HL (0 ou 1)
			sb.WriteString("    EX DE, HL\n") // HL = Left, DE = Right
			sb.WriteString("    OR A\n")
			sb.WriteString("    SBC HL, DE\n") // HL = Left - Right, flags Z e C ajustadas

			trueLbl := cg.newLabel("CondTrue")
			endLbl := cg.newLabel("CondEnd")

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

			// Falso: HL = 0
			sb.WriteString("    LD HL, 0000h\n")
			sb.WriteString(fmt.Sprintf("    JP %s\n", endLbl))
			// Verdadeiro: HL = 1
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

func escapeString(s string) string {
	var sb strings.Builder
	for i := 0; i < len(s); i++ {
		b := s[i]
		if b == '"' {
			sb.WriteString(`\"`)
		} else if b == '$' {
			// Se tiver literal $, pode conflitar com terminador BDOS, mas para strings literais normais mantemos
			sb.WriteByte(b)
		} else {
			sb.WriteByte(b)
		}
	}
	return sb.String()
}
