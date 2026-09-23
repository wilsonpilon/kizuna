package wirth80

import (
	"fmt"
	"sort"
	"strings"

	"github.com/wilsonpilon/kizuna/pkg/kaji80"
	"github.com/wilsonpilon/kizuna/pkg/mob"
)

// varSymbol descreve uma variável local ou parâmetro dentro do quadro de
// ativação (IX-relativo) de uma procedure/function -- variáveis globais
// continuam resolvidas por cg.varMap/cg.varLabels, como já era antes de
// existir procedure/function no WIRTH80.
type varSymbol struct {
	Type   string
	Offset int // local: IX-Offset; parâmetro: IX+Offset
}

// CodeGenerator gera código Assembly Z80 a partir da AST do Pascal
type CodeGenerator struct {
	prog          *ProgramNode
	asm           strings.Builder
	labelCounter  int
	stringLits    map[string]string // texto -> label (ex: "StrLit_1")
	varMap        map[string]string // nome da variável global (case-insensitive) -> tipo
	varLabels     map[string]string // nome da variável global -> label (ex: "Var_a")
	needsMul16    bool
	needsDiv16    bool
	needsPrintDec bool
	needsPrintStr bool
	needsPrintChr bool

	// Estado ativo só durante a geração do corpo de uma procedure/function
	// (nil enquanto gera o bloco principal) -- mesmo modelo de escopo já
	// usado em pkg/dignac/codegen.go.
	currentLocals    map[string]*varSymbol
	currentParams    map[string]*varSymbol
	currentProc      *ProcDecl
	currentRetOffset int
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

// isByteType reporta se o tipo ocupa 1 byte (Char/Boolean) em vez da
// palavra de 16 bits usada por Integer/String.
func isByteType(t string) bool {
	u := strings.ToUpper(t)
	return u == "CHAR" || u == "BOOLEAN"
}

func typeSize(t string) int {
	if isByteType(t) {
		return 1
	}
	return 2
}

func containsString(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

// GenerateAsm emite o código Assembly Z80 formatado
func (cg *CodeGenerator) GenerateAsm() (string, error) {
	cg.asm.Reset()

	modName := cg.prog.Name
	if modName == "" {
		modName = "Program"
	}

	cg.asm.WriteString("; =============================================================================\n")
	cg.asm.WriteString("; Código gerado pelo compilador Pascal WIRTH80 - Kizuna Toolchain\n")
	cg.asm.WriteString(fmt.Sprintf("; Programa: %s\n", modName))
	cg.asm.WriteString("; =============================================================================\n\n")

	cg.asm.WriteString(fmt.Sprintf("MODULE %s\n", modName))
	cg.asm.WriteString(fmt.Sprintf("BANK %d\n\n", cg.prog.Bank))

	// PUBLIC: "Start" só entra se o bloco principal (begin...end.) tiver
	// algum comando -- mesma regra que o DIGNAC já usa (só gera Start se
	// existir PROCEDURE Main). Um programa com bloco principal vazio e
	// PUBLIC declarado vira uma "biblioteca" pura, sem ponto de entrada
	// próprio -- necessário pra poder entrar num .COM que já tem Start
	// definido em outro módulo (KAJI80/DIGNAC): só pode haver UM Main por
	// executável, e MUSUBI recusa a linkagem com um erro claro se mais de
	// um módulo tentar definir o mesmo ponto de entrada.
	hasMain := cg.prog.Block != nil && len(cg.prog.Block.Statements) > 0

	var publics []string
	if hasMain {
		publics = append(publics, "Start")
	}
	for _, p := range cg.prog.Publics {
		if !containsString(publics, p) {
			publics = append(publics, p)
		}
	}
	if len(publics) == 0 {
		return "", fmt.Errorf("módulo Pascal sem bloco principal (begin...end. vazio) e sem nenhum PUBLIC declarado -- nada seria exportado")
	}
	cg.asm.WriteString(fmt.Sprintf("PUBLIC %s\n\n", strings.Join(publics, ", ")))

	// Coleta variáveis globais
	for _, decl := range cg.prog.Vars {
		for _, name := range decl.Names {
			upper := strings.ToUpper(name)
			cg.varMap[upper] = decl.Type
			cg.varLabels[upper] = fmt.Sprintf("Var_%s", name)
		}
	}

	// Gera cada procedure/function declarada, na ordem do fonte --
	// declare-antes-de-usar: uma só pode chamar as que já vêm antes dela
	// no arquivo (sem forward declarations nesta leva).
	var procsAsm strings.Builder
	for _, proc := range cg.prog.Procs {
		if err := cg.generateProcDecl(&procsAsm, proc); err != nil {
			return "", err
		}
	}

	// Gera o bloco principal (Start) -- feito depois das procedures pra
	// permitir chamar qualquer uma delas a partir do programa principal.
	var body strings.Builder
	if cg.prog.Block != nil {
		for _, stmt := range cg.prog.Block.Statements {
			if err := cg.generateStmt(&body, stmt); err != nil {
				return "", err
			}
		}
	}

	// Emitir lista de imports EXTERN conforme a necessidade + os símbolos
	// que o programa declarou via EXTERN (deduplicados contra os do
	// próprio runtime, caso coincidam).
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
	if hasMain {
		externs = append(externs, "BDOS_Exit")
	}
	for _, e := range cg.prog.Externs {
		if !containsString(externs, e) {
			externs = append(externs, e)
		}
	}

	if len(externs) > 0 {
		cg.asm.WriteString(fmt.Sprintf("EXTERN %s\n\n", strings.Join(externs, ", ")))
	}

	// Ponto de entrada -- só emitido se este módulo for de fato o Main (ver
	// o comentário acima da lista de PUBLIC).
	if hasMain {
		cg.asm.WriteString("Start:\n")
		cg.asm.WriteString(body.String())
		cg.asm.WriteString("    CALL BDOS_Exit\n\n")
	}

	// Procedures/functions definidas pelo usuário
	cg.asm.WriteString(procsAsm.String())

	// Seção de Dados: Literais de String. Nomes ordenados antes de emitir
	// -- iterar cg.stringLits (um map do Go) direto embaralharia a ordem a
	// cada remontagem, mesma classe de não-determinismo já corrigida no
	// KAJI80 e no DIGNAC nesta mesma sessão.
	if len(cg.stringLits) > 0 {
		texts := make([]string, 0, len(cg.stringLits))
		for str := range cg.stringLits {
			texts = append(texts, str)
		}
		sort.Strings(texts)

		cg.asm.WriteString("; --- Literais de String ---\n")
		for _, str := range texts {
			lbl := cg.stringLits[str]
			cg.asm.WriteString(fmt.Sprintf("%s:\n", lbl))
			cg.asm.WriteString(fmt.Sprintf("    DB \"%s$\"\n", escapeString(str)))
		}
		cg.asm.WriteString("\n")
	}

	// Seção de Variáveis Globais -- mesmo cuidado de ordenação acima.
	if len(cg.varLabels) > 0 {
		names := make([]string, 0, len(cg.varLabels))
		for name := range cg.varLabels {
			names = append(names, name)
		}
		sort.Strings(names)

		cg.asm.WriteString("; --- Variáveis Globais ---\n")
		for _, name := range names {
			lbl := cg.varLabels[name]
			t := strings.ToUpper(cg.varMap[name])
			if t == "CHAR" || t == "BOOLEAN" {
				cg.asm.WriteString(fmt.Sprintf("%s:\n    DB 00h\n", lbl))
			} else {
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

// generateProcDecl gera o código de uma procedure/function definida pelo
// usuário: prólogo com frame pointer IX (mesma ABI de pilha já usada por
// KAJI80/DIGNAC -- ver SPEC.md §7), parâmetros empilhados esquerda→direita
// pelo chamador e lidos em (IX+offset), locais em (IX-offset).
//
// Todo argumento ocupa 2 bytes na pilha independente do tipo declarado --
// PUSH no Z80 só move pares de registrador, então mesmo um parâmetro
// Char/Boolean "lógico" de 1 byte consome um slot de 2 bytes; só a leitura
// de volta (via isByteType) sabe que só o byte baixo importa. Locais já
// reservam o tamanho real (1 byte pra Char/Boolean), porque aí quem aloca o
// frame é o SP, não um PUSH.
//
// function devolve valor por atribuição ao próprio nome dentro do corpo
// (Nome := expr;, estilo Turbo Pascal). Em vez de uma célula global
// compartilhada -- que quebraria se essa function chamasse outra function
// antes do próprio epílogo ler o valor de volta -- reserva um slot LOCAL
// oculto de 2 bytes (sempre o primeiro do frame): a atribuição ao nome
// grava nele, e o epílogo carrega HL dali antes do RET.
func (cg *CodeGenerator) generateProcDecl(sb *strings.Builder, proc *ProcDecl) error {
	cg.currentLocals = make(map[string]*varSymbol)
	cg.currentParams = make(map[string]*varSymbol)
	cg.currentProc = proc

	numParams := len(proc.Params)
	for i, p := range proc.Params {
		offset := 4 + 2*(numParams-1-i)
		cg.currentParams[strings.ToUpper(p.Name)] = &varSymbol{Type: p.Type, Offset: offset}
	}

	localOffset := 0
	if proc.IsFunction {
		localOffset += 2
		cg.currentRetOffset = localOffset
	}
	for _, locDecl := range proc.Locals {
		for _, name := range locDecl.Names {
			localOffset += typeSize(locDecl.Type)
			cg.currentLocals[strings.ToUpper(name)] = &varSymbol{Type: locDecl.Type, Offset: localOffset}
		}
	}
	frameSize := localOffset

	kind := "procedure"
	if proc.IsFunction {
		kind = "function"
	}
	sb.WriteString("; -----------------------------------------------------------------------------\n")
	sb.WriteString(fmt.Sprintf("; %s %s (Locais: %d bytes, Parâmetros: %d)\n", kind, proc.Name, frameSize, numParams))
	sb.WriteString("; -----------------------------------------------------------------------------\n")
	sb.WriteString(fmt.Sprintf("%s:\n", proc.Name))
	sb.WriteString("    PUSH IX\n")
	sb.WriteString("    LD IX, 0000h\n")
	sb.WriteString("    ADD IX, SP\n")

	if frameSize > 0 {
		sb.WriteString(fmt.Sprintf("    LD HL, -%d\n", frameSize))
		sb.WriteString("    ADD HL, SP\n")
		sb.WriteString("    LD SP, HL\n")
	}

	if proc.Body != nil {
		for _, stmt := range proc.Body.Statements {
			if err := cg.generateStmt(sb, stmt); err != nil {
				return err
			}
		}
	}

	if proc.IsFunction {
		sb.WriteString(fmt.Sprintf("    LD L, (IX - %d)\n", cg.currentRetOffset))
		sb.WriteString(fmt.Sprintf("    LD H, (IX - %d)\n", cg.currentRetOffset-1))
	}
	sb.WriteString("    LD SP, IX\n")
	sb.WriteString("    POP IX\n")
	sb.WriteString("    RET\n\n")

	cg.currentLocals = nil
	cg.currentParams = nil
	cg.currentProc = nil
	return nil
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
		// var := expr (ou NomeDaFunction := expr dentro do corpo de uma
		// function -- storeVar decide o destino certo)
		if err := cg.generateExpr(sb, s.Expr); err != nil {
			return err
		}
		return cg.storeVar(sb, s.VarName)

	case *CallStmt:
		// Chamada de procedure como comando: empilha argumentos
		// esquerda→direita, CALL, limpa a pilha (convenção do chamador).
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

	case *WriteStmt:
		for _, arg := range s.Args {
			if strLit, ok := arg.(*StringLiteral); ok {
				// Impressão direta de string literal
				lbl := cg.getStringLabel(strLit.Value)
				cg.needsPrintStr = true
				sb.WriteString(fmt.Sprintf("    LD DE, %s\n", lbl))
				sb.WriteString("    CALL BDOS_PrintString\n")
			} else {
				// Expressão inteira: avalia em HL e chama PrintDec16
				if err := cg.generateExpr(sb, arg); err != nil {
					return err
				}
				cg.needsPrintDec = true
				sb.WriteString("    CALL PrintDec16\n")
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

// loadVar carrega o valor de uma variável (local, parâmetro ou global) em
// HL. Locais/parâmetros são IX-relativos -- não existe "LD HL,(IX+d)" no
// Z80, só formas de 8 bits, por isso sempre dois LDs separados (L depois
// H); uma global tem "LD HL,(nn)" disponível num instrução só. Char/
// Boolean sempre só leem o byte baixo e zeram H.
func (cg *CodeGenerator) loadVar(sb *strings.Builder, name string) error {
	upper := strings.ToUpper(name)

	if sym, ok := cg.currentLocals[upper]; ok {
		if isByteType(sym.Type) {
			sb.WriteString(fmt.Sprintf("    LD A, (IX - %d)\n", sym.Offset))
			sb.WriteString("    LD L, A\n    LD H, 00h\n")
		} else {
			sb.WriteString(fmt.Sprintf("    LD L, (IX - %d)\n", sym.Offset))
			sb.WriteString(fmt.Sprintf("    LD H, (IX - %d)\n", sym.Offset-1))
		}
		return nil
	}

	if sym, ok := cg.currentParams[upper]; ok {
		if isByteType(sym.Type) {
			sb.WriteString(fmt.Sprintf("    LD A, (IX + %d)\n", sym.Offset))
			sb.WriteString("    LD L, A\n    LD H, 00h\n")
		} else {
			sb.WriteString(fmt.Sprintf("    LD L, (IX + %d)\n", sym.Offset))
			sb.WriteString(fmt.Sprintf("    LD H, (IX + %d)\n", sym.Offset+1))
		}
		return nil
	}

	if lbl, ok := cg.varLabels[upper]; ok {
		t := strings.ToUpper(cg.varMap[upper])
		if t == "CHAR" || t == "BOOLEAN" {
			sb.WriteString(fmt.Sprintf("    LD A, (%s)\n", lbl))
			sb.WriteString("    LD L, A\n    LD H, 00h\n")
		} else {
			sb.WriteString(fmt.Sprintf("    LD HL, (%s)\n", lbl))
		}
		return nil
	}

	return fmt.Errorf("variável não declarada '%s'", name)
}

// storeVar grava HL numa variável (local, parâmetro ou global) -- mesmas
// regras de endereçamento de loadVar. Se o nome bate com o da própria
// function corrente, é a atribuição de valor de retorno (Nome := expr):
// grava no slot local oculto reservado em generateProcDecl, não numa
// variável de verdade.
func (cg *CodeGenerator) storeVar(sb *strings.Builder, name string) error {
	upper := strings.ToUpper(name)

	if cg.currentProc != nil && cg.currentProc.IsFunction && upper == strings.ToUpper(cg.currentProc.Name) {
		sb.WriteString(fmt.Sprintf("    LD (IX - %d), L\n", cg.currentRetOffset))
		sb.WriteString(fmt.Sprintf("    LD (IX - %d), H\n", cg.currentRetOffset-1))
		return nil
	}

	if sym, ok := cg.currentLocals[upper]; ok {
		if isByteType(sym.Type) {
			sb.WriteString("    LD A, L\n")
			sb.WriteString(fmt.Sprintf("    LD (IX - %d), A\n", sym.Offset))
		} else {
			sb.WriteString(fmt.Sprintf("    LD (IX - %d), L\n", sym.Offset))
			sb.WriteString(fmt.Sprintf("    LD (IX - %d), H\n", sym.Offset-1))
		}
		return nil
	}

	if sym, ok := cg.currentParams[upper]; ok {
		if isByteType(sym.Type) {
			sb.WriteString("    LD A, L\n")
			sb.WriteString(fmt.Sprintf("    LD (IX + %d), A\n", sym.Offset))
		} else {
			sb.WriteString(fmt.Sprintf("    LD (IX + %d), L\n", sym.Offset))
			sb.WriteString(fmt.Sprintf("    LD (IX + %d), H\n", sym.Offset+1))
		}
		return nil
	}

	if lbl, ok := cg.varLabels[upper]; ok {
		t := strings.ToUpper(cg.varMap[upper])
		if t == "CHAR" || t == "BOOLEAN" {
			sb.WriteString("    LD A, L\n")
			sb.WriteString(fmt.Sprintf("    LD (%s), A\n", lbl))
		} else {
			sb.WriteString(fmt.Sprintf("    LD (%s), HL\n", lbl))
		}
		return nil
	}

	return fmt.Errorf("variável não declarada '%s'", name)
}

// generateExpr avalia a expressão e deixa o resultado em HL (16 bits)
func (cg *CodeGenerator) generateExpr(sb *strings.Builder, expr Expr) error {
	switch e := expr.(type) {
	case *NumberLiteral:
		sb.WriteString(fmt.Sprintf("    LD HL, %04Xh ; %d\n", uint16(e.Value), e.Value))
		return nil

	case *VarExpr:
		return cg.loadVar(sb, e.Name)

	case *CallExpr:
		// Chamada de function como expressão: empilha argumentos, CALL,
		// limpa a pilha SEM perder o valor de retorno que o CALL deixou em
		// HL. Troca HL/DE antes de calcular o novo SP (em vez de somar o
		// tamanho de limpeza direto em HL, que destruiria o retorno) --
		// mesma necessidade que pkg/dignac/codegen.go tem, mas com a
		// sequência corrigida (a versão do DIGNAC soma o tamanho de
		// limpeza ao próprio HL antes de trocar, o que na verdade
		// corrompe o valor de retorno com esse tamanho; não repetida
		// aqui).
		for _, arg := range e.Args {
			if err := cg.generateExpr(sb, arg); err != nil {
				return err
			}
			sb.WriteString("    PUSH HL\n")
		}
		sb.WriteString(fmt.Sprintf("    CALL %s\n", e.Name))
		if len(e.Args) > 0 {
			sb.WriteString("    EX DE, HL\n") // DE = retorno
			sb.WriteString(fmt.Sprintf("    LD HL, %d\n", len(e.Args)*2))
			sb.WriteString("    ADD HL, SP\n") // HL = novo SP
			sb.WriteString("    LD SP, HL\n")
			sb.WriteString("    EX DE, HL\n") // HL = retorno de volta
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
