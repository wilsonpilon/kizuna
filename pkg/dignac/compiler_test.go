package dignac

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wilsonpilon/kizuna/pkg/mob"
	"github.com/wilsonpilon/kizuna/pkg/musubi"
)

func TestLexer(t *testing.T) {
	src := `
	' Comentário com apóstrofo
	MODULE TestBasic
	BANK 2
	PUBLIC Calc, Greet
	PROCEDURE Calc(a%, b%)
		LOCAL res%
		res% = (a% + b%) * 2
		PRINT "Resultado: "; res%
	END PROCEDURE
	END MODULE
	`

	lexer := NewLexer(src)
	var tokens []Token
	for {
		tok, err := lexer.NextToken()
		if err != nil {
			t.Fatalf("Erro no lexer: %v", err)
		}
		if tok.Type == TokenEOF {
			break
		}
		tokens = append(tokens, tok)
	}

	if len(tokens) == 0 {
		t.Fatalf("Nenhum token gerado")
	}

	// Primeiro token útil após newlines deve ser MODULE
	var firstUseful Token
	for _, tok := range tokens {
		if tok.Type != TokenNewline {
			firstUseful = tok
			break
		}
	}

	if firstUseful.Type != TokenModule {
		t.Errorf("Esperado TokenModule, obteve %v", firstUseful)
	}
}

func TestParser(t *testing.T) {
	src := `
	MODULE MathLib
	BANK 1
	PUBLIC Dobro

	PROCEDURE Dobro(n%)
		LOCAL r%
		r% = n% * 2
		PRINT "Dobro = "; r%
	END PROCEDURE
	END MODULE
	`

	lexer := NewLexer(src)
	parser, err := NewParser(lexer)
	if err != nil {
		t.Fatalf("Erro ao inicializar parser: %v", err)
	}

	mod, err := parser.ParseModule()
	if err != nil {
		t.Fatalf("Erro ao parsear módulo: %v", err)
	}

	if mod.Name != "MathLib" {
		t.Errorf("Esperado nome 'MathLib', obteve '%s'", mod.Name)
	}

	if mod.Bank != 1 {
		t.Errorf("Esperado banco 1, obteve %d", mod.Bank)
	}

	if len(mod.Procedures) != 1 {
		t.Fatalf("Esperado 1 procedimento, obteve %d", len(mod.Procedures))
	}

	proc := mod.Procedures[0]
	if proc.Name != "Dobro" {
		t.Errorf("Esperado nome 'Dobro', obteve '%s'", proc.Name)
	}
	if !proc.IsPublic {
		t.Errorf("Esperado que procedimento seja público")
	}
	if len(proc.Params) != 1 || proc.Params[0].Name != "n%" {
		t.Errorf("Esperado parâmetro 'n%%', obteve %v", proc.Params)
	}
	if len(proc.Locals) != 1 || proc.Locals[0].Decls[0].Name != "r%" {
		t.Errorf("Esperado local 'r%%', obteve %v", proc.Locals)
	}
}

func TestCompileChartDemo(t *testing.T) {
	// Lê o arquivo real demo/chart.bas
	content, err := os.ReadFile("../../demo/chart.bas")
	if err != nil {
		t.Fatalf("Erro ao ler demo/chart.bas: %v", err)
	}

	lexer := NewLexer(string(content))
	parser, err := NewParser(lexer)
	if err != nil {
		t.Fatalf("Erro ao criar parser: %v", err)
	}

	mod, err := parser.ParseModule()
	if err != nil {
		t.Fatalf("Erro ao analisar demo/chart.bas: %v", err)
	}

	if mod.Name != "Chart" {
		t.Errorf("Esperado nome 'Chart', obteve '%s'", mod.Name)
	}

	cg := NewCodeGenerator(mod)
	asmCode, err := cg.GenerateAsm()
	if err != nil {
		t.Fatalf("Erro ao gerar assembly: %v", err)
	}

	if !strings.Contains(asmCode, "MODULE Chart") {
		t.Errorf("Assembly gerado deve conter 'MODULE Chart'")
	}
	if !strings.Contains(asmCode, "Desenhar:") {
		t.Errorf("Assembly gerado deve conter o label do procedimento 'Desenhar:'")
	}
	if !strings.Contains(asmCode, "VDP_BoxFill") {
		t.Errorf("Assembly gerado deve conter chamada para 'VDP_BoxFill'")
	}
	if !strings.Contains(asmCode, "VDP_PSet") {
		t.Errorf("Assembly gerado deve conter chamada para 'VDP_PSet'")
	}

	// Compilação completa para .MOB através do KAJI80
	obj, asmOut, err := cg.Compile()
	if err != nil {
		t.Fatalf("Erro ao compilar demo/chart.bas para .MOB: %v\nCódigo emitido:\n%s", err, asmOut)
	}

	if obj == nil {
		t.Fatalf("Objeto .MOB gerado é nulo")
	}

	if len(obj.Segments) == 0 {
		t.Fatalf("Objeto .MOB deve conter ao menos 1 segmento")
	}

	// Verifica se exporta o símbolo 'Desenhar'
	hasDesenhar := false
	for _, sym := range obj.Symbols {
		if sym.Name == "Desenhar" {
			hasDesenhar = true
			break
		}
	}
	if !hasDesenhar {
		t.Errorf("Objeto .MOB gerado não exporta o símbolo 'Desenhar'")
	}
}

func TestCompileAndLinkHelloAndCalc(t *testing.T) {
	libPath := "../../lib/msxlib.hlib"
	if _, err := os.Stat(libPath); os.IsNotExist(err) {
		t.Skipf("msxlib.hlib não encontrado em %s, pulando teste de linkagem", libPath)
		return
	}

	tmpDir := t.TempDir()

	samples := []string{"../../sample/basic/hello.bas", "../../sample/basic/calc.bas"}
	for _, path := range samples {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("Erro ao ler %s: %v", path, err)
		}

		lexer := NewLexer(string(content))
		parser, err := NewParser(lexer)
		if err != nil {
			t.Fatalf("Erro no parser para %s: %v", path, err)
		}

		mod, err := parser.ParseModule()
		if err != nil {
			t.Fatalf("Erro sintático em %s: %v", path, err)
		}

		cg := NewCodeGenerator(mod)
		obj, asmOut, err := cg.Compile()
		if err != nil {
			t.Fatalf("Erro na compilação de %s: %v\nAssembly:\n%s", path, err, asmOut)
		}

		// Salva .mob/.com num diretório temporário -- nunca em cima dos
		// fontes de sample/, que são versionados no Git (escrever ali
		// sujava o working tree a cada `go test ./...` sem nenhuma mudança
		// de fonte real, só reordenação interna da pool de strings).
		base := strings.TrimSuffix(filepath.Base(path), ".bas")
		mobPath := filepath.Join(tmpDir, base+".mob")
		if err := mob.SaveToFile(mobPath, obj); err != nil {
			t.Fatalf("Erro ao gravar %s: %v", mobPath, err)
		}

		// Linka com msxlib.hlib via MUSUBI
		outCom := filepath.Join(tmpDir, base+".com")
		cfg := musubi.LinkerConfig{
			BaseAddress: 0x0100,
			EntryPoint:  "Start",
		}
		res, err := musubi.LinkToFile(outCom, cfg, mobPath, libPath)
		if err != nil {
			t.Fatalf("Falha no smart-linking de %s: %v", path, err)
		}

		if len(res.Binary) == 0 {
			t.Fatalf("Executável binário gerado para %s está vazio", path)
		}
	}
}

func TestParseDimMixedSuffixes(t *testing.T) {
	src := `
	MODULE Types
	DIM nome$, idade%, altura!
	END MODULE
	`

	lexer := NewLexer(src)
	parser, err := NewParser(lexer)
	if err != nil {
		t.Fatalf("Erro ao criar parser: %v", err)
	}
	mod, err := parser.ParseModule()
	if err != nil {
		t.Fatalf("Erro ao parsear módulo: %v", err)
	}

	if len(mod.Globals) != 1 || len(mod.Globals[0].Decls) != 3 {
		t.Fatalf("Esperado 1 DIM com 3 variáveis, obteve %v", mod.Globals)
	}

	want := map[string]string{"nome$": "STRING", "idade%": "INTEGER", "altura!": "SINGLE"}
	for _, d := range mod.Globals[0].Decls {
		expected, ok := want[d.Name]
		if !ok {
			t.Fatalf("Variável inesperada '%s'", d.Name)
		}
		if d.Type != expected {
			t.Errorf("Variável '%s': esperado tipo %s por sufixo, obteve %s", d.Name, expected, d.Type)
		}
	}
}

func TestLexOctalAndBinaryLiterals(t *testing.T) {
	lexer := NewLexer("&O17 &B101")

	tok1, err := lexer.NextToken()
	if err != nil {
		t.Fatalf("Erro no lexer (octal): %v", err)
	}
	if tok1.Type != TokenNumber || tok1.Value != "&O17" {
		t.Errorf("Esperado token &O17, obteve %v", tok1)
	}

	tok2, err := lexer.NextToken()
	if err != nil {
		t.Fatalf("Erro no lexer (binário): %v", err)
	}
	if tok2.Type != TokenNumber || tok2.Value != "&B101" {
		t.Errorf("Esperado token &B101, obteve %v", tok2)
	}
}

func TestParseOctalAndBinaryLiteralsAsValues(t *testing.T) {
	src := `
	MODULE RadixTest
	PUBLIC Main
	PROCEDURE Main
		LOCAL o%, b%
		o% = &O17
		b% = &B101
	END PROCEDURE
	END MODULE
	`

	lexer := NewLexer(src)
	parser, err := NewParser(lexer)
	if err != nil {
		t.Fatalf("Erro ao criar parser: %v", err)
	}
	mod, err := parser.ParseModule()
	if err != nil {
		t.Fatalf("Erro ao parsear módulo: %v", err)
	}

	proc := mod.Procedures[0]
	assign1, ok := proc.Body[0].(*AssignStmt)
	if !ok {
		t.Fatalf("Esperado AssignStmt, obteve %T", proc.Body[0])
	}
	if num1, ok := assign1.Value.(*NumberExpr); !ok || num1.Value != 15 {
		t.Fatalf("Esperado &O17 = 15, obteve %v", assign1.Value)
	}

	assign2, ok := proc.Body[1].(*AssignStmt)
	if !ok {
		t.Fatalf("Esperado AssignStmt, obteve %T", proc.Body[1])
	}
	if num2, ok := assign2.Value.(*NumberExpr); !ok || num2.Value != 5 {
		t.Fatalf("Esperado &B101 = 5, obteve %v", assign2.Value)
	}
}

func TestCompileMixedTypesEndToEnd(t *testing.T) {
	src := `
	MODULE MixedTypes
	PUBLIC Main
	DIM nome$, idade%, altura!

	PROCEDURE Main
		nome$ = "Kizuna"
		idade% = 42
		altura! = 1.75
		PRINT nome$
		PRINT idade%
	END PROCEDURE
	END MODULE
	`

	lexer := NewLexer(src)
	parser, err := NewParser(lexer)
	if err != nil {
		t.Fatalf("Erro ao criar parser: %v", err)
	}
	mod, err := parser.ParseModule()
	if err != nil {
		t.Fatalf("Erro ao parsear módulo: %v", err)
	}

	cg := NewCodeGenerator(mod)
	obj, asmOut, err := cg.Compile()
	if err != nil {
		t.Fatalf("Erro ao compilar tipos mistos até .MOB: %v\nAssembly:\n%s", err, asmOut)
	}
	if obj == nil {
		t.Fatalf("Objeto .MOB gerado é nulo")
	}

	if !strings.Contains(asmOut, "StrCopyLen") {
		t.Errorf("Assembly gerado deveria referenciar StrCopyLen (EXTERN) para a atribuição de nome$")
	}
	if !strings.Contains(asmOut, "BDOS_PrintLenStr") {
		t.Errorf("Assembly gerado deveria referenciar BDOS_PrintLenStr (EXTERN) para o PRINT de nome$")
	}
	if !strings.Contains(asmOut, "DS 256") {
		t.Errorf("Global STRING 'nome$' deveria reservar 256 bytes (DS 256)")
	}
	if !strings.Contains(asmOut, "DS 4") {
		t.Errorf("Global SINGLE 'altura!' deveria reservar 4 bytes (DS 4)")
	}
}

func TestCompileSingleDoubleDeclareAndAssign(t *testing.T) {
	src := `
	MODULE FloatTypes
	PUBLIC Main
	DIM x!, y#, z#

	PROCEDURE Main
		x! = 3.14
		y# = 2.71828d0
		z# = 100
	END PROCEDURE
	END MODULE
	`

	lexer := NewLexer(src)
	parser, err := NewParser(lexer)
	if err != nil {
		t.Fatalf("Erro ao criar parser: %v", err)
	}
	mod, err := parser.ParseModule()
	if err != nil {
		t.Fatalf("Erro ao parsear módulo: %v", err)
	}

	cg := NewCodeGenerator(mod)
	obj, asmOut, err := cg.Compile()
	if err != nil {
		t.Fatalf("Erro ao compilar SINGLE/DOUBLE (declarar+atribuir) até .MOB: %v\nAssembly:\n%s", err, asmOut)
	}
	if obj == nil {
		t.Fatalf("Objeto .MOB gerado é nulo")
	}

	if !strings.Contains(asmOut, "DS 4") {
		t.Errorf("Global SINGLE 'x!' deveria reservar 4 bytes (DS 4)")
	}
	if n := strings.Count(asmOut, "DS 8"); n != 2 {
		t.Errorf("Esperado 2 globais DOUBLE reservando 8 bytes (DS 8) cada, encontrado %d", n)
	}
}

// TestFloatMulDivNotYetImplementedError: '+'/'-' e comparação de SINGLE
// ganharam suporte real (ver TestCompileFloatAddSubAssign e
// TestCompileFloatComparison abaixo) -- '*'/'/' continuam de fora nesta
// leva (motor Z80 só tem Float_Add32/Sub32/Cmp32, Mul32/Div32 ficaram pra
// uma leva futura por complexidade de normalização descoberta durante a
// implementação). Este teste, que antes cobria TODA aritmética float como
// erro, foi ajustado pra continuar cobrindo o que ainda É erro.
func TestFloatMulDivNotYetImplementedError(t *testing.T) {
	src := `
	MODULE FloatArith
	PUBLIC Main
	DIM x!, y!

	PROCEDURE Main
		x! = 1.0
		y! = 2.0
		x! = x! * y!
	END PROCEDURE
	END MODULE
	`

	lexer := NewLexer(src)
	parser, err := NewParser(lexer)
	if err != nil {
		t.Fatalf("Erro ao criar parser: %v", err)
	}
	mod, err := parser.ParseModule()
	if err != nil {
		t.Fatalf("Erro ao parsear módulo: %v", err)
	}

	cg := NewCodeGenerator(mod)
	if _, err := cg.GenerateAsm(); err == nil {
		t.Fatalf("Esperado erro de compilação ao multiplicar dois SINGLE, mas compilou com sucesso")
	} else if !strings.Contains(err.Error(), "ponto flutuante") {
		t.Errorf("Esperado erro mencionando 'ponto flutuante', obteve: %v", err)
	}
}

// TestCompileFloatAddSubAssign: "x! = a! + b!" e "x! = a! - b!" agora
// compilam de verdade, chamando Float_Add32/Float_Sub32 do MSXLIB.
func TestCompileFloatAddSubAssign(t *testing.T) {
	for _, tc := range []struct {
		op       string
		wantCall string
		wantExt  string
	}{
		{"+", "CALL Float_Add32", "Float_Add32"},
		{"-", "CALL Float_Sub32", "Float_Sub32"},
	} {
		src := fmt.Sprintf(`
		MODULE FloatArith
		PUBLIC Main
		DIM x!, a!, b!

		PROCEDURE Main
			a! = 1.0
			b! = 2.0
			x! = a! %s b!
		END PROCEDURE
		END MODULE
		`, tc.op)

		lexer := NewLexer(src)
		parser, err := NewParser(lexer)
		if err != nil {
			t.Fatalf("[%s] Erro ao criar parser: %v", tc.op, err)
		}
		mod, err := parser.ParseModule()
		if err != nil {
			t.Fatalf("[%s] Erro ao parsear módulo: %v", tc.op, err)
		}

		cg := NewCodeGenerator(mod)
		asm, err := cg.GenerateAsm()
		if err != nil {
			t.Fatalf("[%s] Erro ao compilar: %v", tc.op, err)
		}
		if !strings.Contains(asm, tc.wantCall) {
			t.Errorf("[%s] esperava '%s' na assembly gerada, não achou:\n%s", tc.op, tc.wantCall, asm)
		}
		if !strings.Contains(asm, "EXTERN") || !strings.Contains(asm, tc.wantExt) {
			t.Errorf("[%s] esperava EXTERN '%s' na assembly gerada, não achou:\n%s", tc.op, tc.wantExt, asm)
		}
	}
}

// TestCompileFloatComparison: "IF a! > b! THEN" compila usando
// Float_Cmp32, não o caminho inteiro (SBC HL,DE).
func TestCompileFloatComparison(t *testing.T) {
	src := `
	MODULE FloatCmp
	PUBLIC Main
	DIM a!, b!

	PROCEDURE Main
		a! = 1.0
		b! = 2.0
		IF a! > b! THEN
			a! = b!
		END IF
	END PROCEDURE
	END MODULE
	`

	lexer := NewLexer(src)
	parser, err := NewParser(lexer)
	if err != nil {
		t.Fatalf("Erro ao criar parser: %v", err)
	}
	mod, err := parser.ParseModule()
	if err != nil {
		t.Fatalf("Erro ao parsear módulo: %v", err)
	}

	cg := NewCodeGenerator(mod)
	asm, err := cg.GenerateAsm()
	if err != nil {
		t.Fatalf("Erro ao compilar: %v", err)
	}
	if !strings.Contains(asm, "CALL Float_Cmp32") {
		t.Errorf("esperava 'CALL Float_Cmp32' na assembly gerada, não achou:\n%s", asm)
	}
	if !strings.Contains(asm, "EXTERN") || !strings.Contains(asm, "Float_Cmp32") {
		t.Errorf("esperava EXTERN 'Float_Cmp32' na assembly gerada, não achou:\n%s", asm)
	}
}

// TestFloatNestedExpressionStillErrors: expressões float aninhadas
// ("(a!+b!)*c!" -- aqui simplificado pra uma soma dentro de outra soma)
// continuam dando erro de compilação claro -- não existe alocação de
// temporários nesta leva, então só operandos simples são aceitos.
func TestFloatNestedExpressionStillErrors(t *testing.T) {
	src := `
	MODULE FloatNested
	PUBLIC Main
	DIM x!, a!, b!, c!

	PROCEDURE Main
		a! = 1.0
		b! = 2.0
		c! = 3.0
		x! = a! + b! + c!
	END PROCEDURE
	END MODULE
	`

	lexer := NewLexer(src)
	parser, err := NewParser(lexer)
	if err != nil {
		t.Fatalf("Erro ao criar parser: %v", err)
	}
	mod, err := parser.ParseModule()
	if err != nil {
		t.Fatalf("Erro ao parsear módulo: %v", err)
	}

	cg := NewCodeGenerator(mod)
	if _, err := cg.GenerateAsm(); err == nil {
		t.Fatalf("Esperado erro de compilação numa expressão float aninhada, mas compilou com sucesso")
	}
}

// TestFloatMixedTypeComparisonStillErrors: comparar um SINGLE com um
// INTEGER continua sendo um erro de compilação claro, não um resultado
// silenciosamente errado.
func TestFloatMixedTypeComparisonStillErrors(t *testing.T) {
	src := `
	MODULE FloatMixedCmp
	PUBLIC Main
	DIM a!, n%

	PROCEDURE Main
		a! = 1.0
		n% = 2
		IF a! > n% THEN
			n% = 0
		END IF
	END PROCEDURE
	END MODULE
	`

	lexer := NewLexer(src)
	parser, err := NewParser(lexer)
	if err != nil {
		t.Fatalf("Erro ao criar parser: %v", err)
	}
	mod, err := parser.ParseModule()
	if err != nil {
		t.Fatalf("Erro ao parsear módulo: %v", err)
	}

	cg := NewCodeGenerator(mod)
	if _, err := cg.GenerateAsm(); err == nil {
		t.Fatalf("Esperado erro de compilação numa comparação SINGLE/INTEGER mista, mas compilou com sucesso")
	}
}

// TestCompileForNegativeStep cobre o bug real do teste de término do FOR
// assumir sempre um STEP positivo (SBC HL,DE + "Var>End encerra", sem
// considerar STEP negativo). Aqui só verificamos que o caminho de código
// consciente de direção (DGN_ForStepSign + AND 80h) é realmente emitido
// quando há STEP explícito -- a correção em si (as 5 iterações esperadas
// de "FOR i%=5 TO 1 STEP -1") foi conferida por rastreamento manual
// instrução-a-instrução (não há emulador Z80 neste repositório para
// verificar automaticamente a contagem de iterações em tempo de execução;
// recomenda-se testar em hardware/openMSX antes de dar como definitivo).
func TestCompileForNegativeStep(t *testing.T) {
	src := `
	MODULE ForDesc
	PUBLIC Main
	PROCEDURE Main
		LOCAL i%
		FOR i% = 5 TO 1 STEP -1
			PRINT i%
		NEXT i%
	END PROCEDURE
	END MODULE
	`

	lexer := NewLexer(src)
	parser, err := NewParser(lexer)
	if err != nil {
		t.Fatalf("Erro ao criar parser: %v", err)
	}
	mod, err := parser.ParseModule()
	if err != nil {
		t.Fatalf("Erro ao parsear módulo: %v", err)
	}

	cg := NewCodeGenerator(mod)
	obj, asmOut, err := cg.Compile()
	if err != nil {
		t.Fatalf("Erro ao compilar FOR com STEP negativo até .MOB: %v\nAssembly:\n%s", err, asmOut)
	}
	if obj == nil {
		t.Fatalf("Objeto .MOB gerado é nulo")
	}

	if !strings.Contains(asmOut, "DGN_ForStepSign") {
		t.Errorf("Assembly gerado deveria referenciar DGN_ForStepSign (teste de término consciente de direção)")
	}
	if !strings.Contains(asmOut, "AND 80h") {
		t.Errorf("Assembly gerado deveria conferir o bit de sinal do STEP (AND 80h)")
	}
}

// TestCompileForWithoutStepOmitsDirectionCheck é o teste de regressão
// complementar: um FOR sem STEP (incremento implícito +1, sempre
// ascendente) não deve pagar o custo nem o risco do teste de direção --
// continua gerando exatamente o teste de término simples de antes.
func TestCompileForWithoutStepOmitsDirectionCheck(t *testing.T) {
	src := `
	MODULE ForAsc
	PUBLIC Main
	PROCEDURE Main
		LOCAL i%
		FOR i% = 1 TO 3
			PRINT i%
		NEXT i%
	END PROCEDURE
	END MODULE
	`

	lexer := NewLexer(src)
	parser, err := NewParser(lexer)
	if err != nil {
		t.Fatalf("Erro ao criar parser: %v", err)
	}
	mod, err := parser.ParseModule()
	if err != nil {
		t.Fatalf("Erro ao parsear módulo: %v", err)
	}

	cg := NewCodeGenerator(mod)
	_, asmOut, err := cg.Compile()
	if err != nil {
		t.Fatalf("Erro ao compilar FOR sem STEP até .MOB: %v\nAssembly:\n%s", err, asmOut)
	}

	if strings.Contains(asmOut, "DGN_ForStepSign") {
		t.Errorf("FOR sem STEP não deveria emitir o teste de direção (DGN_ForStepSign)")
	}
}

// TestCompileOpenAppendSeeksToEnd cobre o bug real de OPEN...FOR APPEND se
// comportar exatamente igual a FOR OUTPUT (ambos caindo em
// BDOS_FileCreate, que cria OU TRUNCA um arquivo existente) -- sem nunca
// posicionar o ponteiro no fim do arquivo antes de escrever. Confere que
// APPEND agora referencia BDOS_FileSeek (a correção real) e que OUTPUT
// continua sem precisar dele (regressão).
func TestCompileOpenAppendSeeksToEnd(t *testing.T) {
	srcAppend := `
	MODULE AppendTest
	PUBLIC Main
	PROCEDURE Main
		OPEN "LOG.TXT" FOR APPEND AS #1
		PRINT #1, "linha"
		CLOSE #1
	END PROCEDURE
	END MODULE
	`
	lexer := NewLexer(srcAppend)
	parser, err := NewParser(lexer)
	if err != nil {
		t.Fatalf("Erro ao criar parser (APPEND): %v", err)
	}
	mod, err := parser.ParseModule()
	if err != nil {
		t.Fatalf("Erro ao parsear módulo (APPEND): %v", err)
	}
	cg := NewCodeGenerator(mod)
	obj, asmOut, err := cg.Compile()
	if err != nil {
		t.Fatalf("Erro ao compilar OPEN...FOR APPEND até .MOB: %v\nAssembly:\n%s", err, asmOut)
	}
	if obj == nil {
		t.Fatalf("Objeto .MOB gerado é nulo")
	}
	if !strings.Contains(asmOut, "BDOS_FileSeek") {
		t.Errorf("OPEN...FOR APPEND deveria referenciar BDOS_FileSeek (posicionar no fim do arquivo)")
	}

	srcOutput := `
	MODULE OutputTest
	PUBLIC Main
	PROCEDURE Main
		OPEN "LOG.TXT" FOR OUTPUT AS #1
		PRINT #1, "linha"
		CLOSE #1
	END PROCEDURE
	END MODULE
	`
	lexer2 := NewLexer(srcOutput)
	parser2, err := NewParser(lexer2)
	if err != nil {
		t.Fatalf("Erro ao criar parser (OUTPUT): %v", err)
	}
	mod2, err := parser2.ParseModule()
	if err != nil {
		t.Fatalf("Erro ao parsear módulo (OUTPUT): %v", err)
	}
	cg2 := NewCodeGenerator(mod2)
	_, asmOut2, err := cg2.Compile()
	if err != nil {
		t.Fatalf("Erro ao compilar OPEN...FOR OUTPUT até .MOB: %v\nAssembly:\n%s", err, asmOut2)
	}
	if strings.Contains(asmOut2, "BDOS_FileSeek") {
		t.Errorf("OPEN...FOR OUTPUT não deveria referenciar BDOS_FileSeek (sempre começa do zero)")
	}
}

// TestCompileFunctionCallInExpressionPreservesReturnValue cobre um bug real
// no CallExpr (chamada de FUNCTION usada dentro de uma expressão maior, não
// como comando isolado): a sequência de limpeza da pilha após o CALL somava
// o tamanho da limpeza (N*2 bytes) direto no próprio HL antes de trocar
// registradores, corrompendo o valor de retorno com esse tamanho. Nunca
// tinha sido pego porque nenhum teste/sample exercitava uma FUNCTION
// chamada com argumentos dentro de uma expressão (só como comando). Aqui só
// confirmamos que a sequência corrigida (EX DE,HL logo após o CALL, não
// ADD HL,DE) é a que sai -- a aritmética em si já é coberta pelo mesmo
// padrão de rastreamento manual usado no resto desta sessão.
func TestCompileFunctionCallInExpressionPreservesReturnValue(t *testing.T) {
	src := `
	MODULE FuncTest
	PUBLIC Main

	FUNCTION Dobro(n%) AS INTEGER
		RETURN n% * 2
	END FUNCTION

	PROCEDURE Main()
		LOCAL resultado%
		resultado% = Dobro(21) + 1
		PRINT resultado%
	END PROCEDURE
	END MODULE
	`

	lexer := NewLexer(src)
	parser, err := NewParser(lexer)
	if err != nil {
		t.Fatalf("Erro ao criar parser: %v", err)
	}
	mod, err := parser.ParseModule()
	if err != nil {
		t.Fatalf("Erro ao parsear módulo: %v", err)
	}

	cg := NewCodeGenerator(mod)
	obj, asmOut, err := cg.Compile()
	if err != nil {
		t.Fatalf("Erro ao compilar chamada de FUNCTION em expressão até .MOB: %v\nAssembly:\n%s", err, asmOut)
	}
	if obj == nil {
		t.Fatalf("Objeto .MOB gerado é nulo")
	}

	if !strings.Contains(asmOut, "CALL Dobro\n    EX DE, HL\n") {
		t.Errorf("Esperada a sequência corrigida (EX DE,HL logo após CALL Dobro) -- assembly gerado:\n%s", asmOut)
	}
	if strings.Contains(asmOut, "CALL Dobro\n    LD DE,") || strings.Contains(asmOut, "ADD HL, DE\n    EX DE, HL\n    ADD HL, SP") {
		t.Errorf("Sequência antiga (bugada) de limpeza de pilha ainda presente no assembly gerado:\n%s", asmOut)
	}
}
