package wirth80

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wilsonpilon/kizuna/pkg/kaji80"
	"github.com/wilsonpilon/kizuna/pkg/mob"
	"github.com/wilsonpilon/kizuna/pkg/musubi"
)

func TestLexer(t *testing.T) {
	src := `
	{ Comentário de bloco }
	program TestProg;
	(* Outro comentário *)
	var
		x, y: Integer;
		c: Char;
	begin
		x := 123;
		y := $2D; // 45 em hex
		WriteLn('Resultado: ', x * y);
	end.
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

	// Primeiro token útil deve ser 'program'
	if tokens[0].Type != TokenProgram {
		t.Errorf("Esperado TokenProgram, obteve %v", tokens[0])
	}
}

func TestParser(t *testing.T) {
	src := `
	program Calc;
	var
		a, b, c: Integer;
	begin
		a := 123;
		b := 45;
		c := a * b;
		WriteLn('Valor = ', c);
	end.
	`

	lexer := NewLexer(src)
	parser, err := NewParser(lexer)
	if err != nil {
		t.Fatalf("Erro ao criar parser: %v", err)
	}

	prog, err := parser.ParseProgram()
	if err != nil {
		t.Fatalf("Erro ao parsear programa: %v", err)
	}

	if prog.Name != "Calc" {
		t.Errorf("Esperado nome 'Calc', obteve '%s'", prog.Name)
	}

	if len(prog.Vars) != 1 || len(prog.Vars[0].Names) != 3 {
		t.Errorf("Esperado 3 variáveis declaradas, obteve %v", prog.Vars)
	}

	if len(prog.Block.Statements) != 4 {
		t.Errorf("Esperado 4 statements no bloco, obteve %d", len(prog.Block.Statements))
	}
}

func TestCodegenAndCompile(t *testing.T) {
	src := `
	program HelloPascal;
	var
		x: Integer;
	begin
		x := 42;
		WriteLn('Ola do Pascal!');
		WriteLn('Numero: ', x);
	end.
	`

	lexer := NewLexer(src)
	parser, err := NewParser(lexer)
	if err != nil {
		t.Fatalf("Erro ao criar parser: %v", err)
	}

	prog, err := parser.ParseProgram()
	if err != nil {
		t.Fatalf("Erro no parse: %v", err)
	}

	cg := NewCodeGenerator(prog)
	obj, asmSource, err := cg.Compile()
	if err != nil {
		t.Fatalf("Falha na compilação do Pascal para MOB:\n%v", err)
	}

	t.Logf("Assembly Z80 gerado com sucesso:\n%s", asmSource)

	if len(obj.Segments) != 1 {
		t.Fatalf("Esperado 1 segmento de código, obteve %d", len(obj.Segments))
	}

	// Verificar se Start é PUBLIC PROC
	var foundStart bool
	for _, sym := range obj.Symbols {
		if sym.Name == "Start" && sym.Class == mob.SymbolPublic {
			foundStart = true
			break
		}
	}
	if !foundStart {
		t.Errorf("Símbolo 'Start' não encontrado como PUBLIC")
	}

	// Verificar se chamadas externas como BDOS_PrintString e PrintDec16 existem
	var foundPrintStr, foundPrintDec bool
	for _, sym := range obj.Symbols {
		if sym.Name == "BDOS_PrintString" && sym.Class == mob.SymbolExtern {
			foundPrintStr = true
		}
		if sym.Name == "PrintDec16" && sym.Class == mob.SymbolExtern {
			foundPrintDec = true
		}
	}
	if !foundPrintStr {
		t.Errorf("Import EXTERN BDOS_PrintString não encontrado")
	}
	if !foundPrintDec {
		t.Errorf("Import EXTERN PrintDec16 não encontrado")
	}

	// Verificar se o texto gerado possui a string terminada em $
	if !strings.Contains(asmSource, "Ola do Pascal!$") {
		t.Errorf("Literal 'Ola do Pascal!$' não encontrado no assembly gerado")
	}
}

func TestParseProcedureWithParams(t *testing.T) {
	src := `
	program ProcTest;
	procedure Foo(a, b: Integer; c: Char);
	var
		total: Integer;
	begin
		total := a + b;
		WriteLn(total);
	end;
	begin
		Foo(1, 2, 'x');
	end.
	`

	lexer := NewLexer(src)
	parser, err := NewParser(lexer)
	if err != nil {
		t.Fatalf("Erro ao criar parser: %v", err)
	}
	prog, err := parser.ParseProgram()
	if err != nil {
		t.Fatalf("Erro ao parsear programa: %v", err)
	}

	if len(prog.Procs) != 1 {
		t.Fatalf("Esperado 1 procedure, obteve %d", len(prog.Procs))
	}
	proc := prog.Procs[0]
	if proc.Name != "Foo" || proc.IsFunction {
		t.Errorf("Esperado procedure 'Foo', obteve %+v", proc)
	}
	if len(proc.Params) != 3 {
		t.Fatalf("Esperado 3 parâmetros, obteve %d: %+v", len(proc.Params), proc.Params)
	}
	if proc.Params[0].Name != "a" || proc.Params[0].Type != "Integer" {
		t.Errorf("Parâmetro 0 inesperado: %+v", proc.Params[0])
	}
	if proc.Params[2].Name != "c" || proc.Params[2].Type != "Char" {
		t.Errorf("Parâmetro 2 inesperado: %+v", proc.Params[2])
	}

	if len(prog.Block.Statements) != 1 {
		t.Fatalf("Esperado 1 statement no bloco principal (a chamada), obteve %d", len(prog.Block.Statements))
	}
	call, ok := prog.Block.Statements[0].(*CallStmt)
	if !ok {
		t.Fatalf("Esperado CallStmt, obteve %T", prog.Block.Statements[0])
	}
	if call.Name != "Foo" || len(call.Args) != 3 {
		t.Errorf("Chamada inesperada: %+v", call)
	}
}

func TestCompileProcedureCall(t *testing.T) {
	src := `
	program ProcCall;
	procedure Dobro(n: Integer);
	begin
		WriteLn(n * 2);
	end;
	begin
		Dobro(21);
	end.
	`

	lexer := NewLexer(src)
	parser, err := NewParser(lexer)
	if err != nil {
		t.Fatalf("Erro ao criar parser: %v", err)
	}
	prog, err := parser.ParseProgram()
	if err != nil {
		t.Fatalf("Erro ao parsear programa: %v", err)
	}

	cg := NewCodeGenerator(prog)
	obj, asmSource, err := cg.Compile()
	if err != nil {
		t.Fatalf("Erro ao compilar procedure com parâmetro até .MOB: %v\nAssembly:\n%s", err, asmSource)
	}
	if obj == nil {
		t.Fatalf("Objeto .MOB gerado é nulo")
	}

	if !strings.Contains(asmSource, "Dobro:") {
		t.Errorf("Assembly gerado deveria conter o label 'Dobro:'")
	}
	if !strings.Contains(asmSource, "CALL Dobro") {
		t.Errorf("Assembly gerado deveria chamar 'Dobro' a partir de Start")
	}
	if !strings.Contains(asmSource, "PUSH IX") {
		t.Errorf("Assembly gerado deveria montar um quadro de ativação (PUSH IX) para Dobro")
	}
}

func TestCompileFunctionInExpression(t *testing.T) {
	src := `
	program FuncTest;
	var
		resultado: Integer;
	function Dobro(x: Integer): Integer;
	begin
		Dobro := x * 2;
	end;
	begin
		resultado := Dobro(5) + 1;
		WriteLn(resultado);
	end.
	`

	lexer := NewLexer(src)
	parser, err := NewParser(lexer)
	if err != nil {
		t.Fatalf("Erro ao criar parser: %v", err)
	}
	prog, err := parser.ParseProgram()
	if err != nil {
		t.Fatalf("Erro ao parsear programa: %v", err)
	}

	if len(prog.Procs) != 1 || !prog.Procs[0].IsFunction {
		t.Fatalf("Esperado 1 function, obteve %+v", prog.Procs)
	}

	cg := NewCodeGenerator(prog)
	obj, asmSource, err := cg.Compile()
	if err != nil {
		t.Fatalf("Erro ao compilar function usada em expressão até .MOB: %v\nAssembly:\n%s", err, asmSource)
	}
	if obj == nil {
		t.Fatalf("Objeto .MOB gerado é nulo")
	}

	if !strings.Contains(asmSource, "CALL Dobro") {
		t.Errorf("Assembly gerado deveria chamar 'Dobro' dentro da expressão")
	}
}

func TestCompilePublicProcedureExported(t *testing.T) {
	src := `
	program Lib;
	PUBLIC Foo;
	procedure Foo(a: Integer);
	begin
		WriteLn(a);
	end;
	begin
	end.
	`

	lexer := NewLexer(src)
	parser, err := NewParser(lexer)
	if err != nil {
		t.Fatalf("Erro ao criar parser: %v", err)
	}
	prog, err := parser.ParseProgram()
	if err != nil {
		t.Fatalf("Erro ao parsear programa: %v", err)
	}
	if len(prog.Publics) != 1 || prog.Publics[0] != "Foo" {
		t.Fatalf("Esperado PUBLIC ['Foo'], obteve %v", prog.Publics)
	}

	cg := NewCodeGenerator(prog)
	obj, asmSource, err := cg.Compile()
	if err != nil {
		t.Fatalf("Erro ao compilar PUBLIC até .MOB: %v\nAssembly:\n%s", err, asmSource)
	}
	if obj == nil {
		t.Fatalf("Objeto .MOB gerado é nulo")
	}

	var foundFoo bool
	for _, sym := range obj.Symbols {
		if sym.Name == "Foo" && sym.Class == mob.SymbolPublic {
			foundFoo = true
			break
		}
	}
	if !foundFoo {
		t.Errorf("Símbolo 'Foo' não encontrado como PUBLIC no .MOB gerado")
	}
}

func TestCompileExternCallCompiles(t *testing.T) {
	src := `
	program CallsExtern;
	EXTERN Something;
	procedure UsaExterno;
	begin
		Something(1, 2);
	end;
	begin
		UsaExterno();
	end.
	`

	lexer := NewLexer(src)
	parser, err := NewParser(lexer)
	if err != nil {
		t.Fatalf("Erro ao criar parser: %v", err)
	}
	prog, err := parser.ParseProgram()
	if err != nil {
		t.Fatalf("Erro ao parsear programa: %v", err)
	}
	if len(prog.Externs) != 1 || prog.Externs[0] != "Something" {
		t.Fatalf("Esperado EXTERN ['Something'], obteve %v", prog.Externs)
	}

	cg := NewCodeGenerator(prog)
	obj, asmSource, err := cg.Compile()
	if err != nil {
		t.Fatalf("Erro ao compilar EXTERN até .MOB: %v\nAssembly:\n%s", err, asmSource)
	}
	if obj == nil {
		t.Fatalf("Objeto .MOB gerado é nulo")
	}

	var foundExtern bool
	for _, sym := range obj.Symbols {
		if sym.Name == "Something" && sym.Class == mob.SymbolExtern {
			foundExtern = true
			break
		}
	}
	if !foundExtern {
		t.Errorf("Símbolo 'Something' não encontrado como EXTERN no .MOB gerado")
	}
}

// TestCompileLibraryModuleOmitsStart cobre o pedido de Wilson de só existir
// um único Main por executável: um módulo WIRTH80 com bloco principal
// (begin...end.) VAZIO e algum PUBLIC declarado vira uma "biblioteca" pura
// -- sem símbolo Start nenhum -- mesma regra que o DIGNAC já aplica (só gera
// Start se existir PROCEDURE Main). Antes desta correção, WIRTH80 sempre
// gerava Start incondicionalmente, então esse mesmo módulo linkado junto
// com um módulo KAJI80/DIGNAC que também define Start colidia por símbolo
// duplicado -- o pkg/musubi agora também dá um erro claro e específico
// ("múltiplos pontos de entrada") se isso ainda acontecer.
func TestCompileLibraryModuleOmitsStart(t *testing.T) {
	src := `
	program Lib;
	PUBLIC Foo;
	procedure Foo(a: Integer);
	begin
		WriteLn(a);
	end;
	begin
	end.
	`

	lexer := NewLexer(src)
	parser, err := NewParser(lexer)
	if err != nil {
		t.Fatalf("Erro ao criar parser: %v", err)
	}
	prog, err := parser.ParseProgram()
	if err != nil {
		t.Fatalf("Erro ao parsear programa: %v", err)
	}

	cg := NewCodeGenerator(prog)
	obj, asmSource, err := cg.Compile()
	if err != nil {
		t.Fatalf("Erro ao compilar módulo-biblioteca até .MOB: %v\nAssembly:\n%s", err, asmSource)
	}
	if obj == nil {
		t.Fatalf("Objeto .MOB gerado é nulo")
	}

	for _, sym := range obj.Symbols {
		if sym.Name == "Start" {
			t.Errorf("Módulo-biblioteca (begin...end. vazio) não deveria exportar 'Start', mas exporta: %+v", sym)
		}
	}

	var foundFoo bool
	for _, sym := range obj.Symbols {
		if sym.Name == "Foo" && sym.Class == mob.SymbolPublic {
			foundFoo = true
		}
	}
	if !foundFoo {
		t.Errorf("Símbolo 'Foo' não encontrado como PUBLIC no .MOB gerado")
	}
}

// TestParseAndCompileBankDirective cobre a diretiva BANK <n>, opcional, na
// mesma posição do KAJI80/DIGNAC (logo após o cabeçalho do programa). Sem
// ela, o módulo cai no banco comum (0) -- com ela, o WIRTH80 finalmente
// consegue gerar um módulo pra um banco paginável, igual o DIGNAC já fazia.
func TestParseAndCompileBankDirective(t *testing.T) {
	src := `
	program Lib;
	BANK 2;
	PUBLIC Foo;
	procedure Foo(a: Integer);
	begin
		WriteLn(a);
	end;
	begin
	end.
	`

	lexer := NewLexer(src)
	parser, err := NewParser(lexer)
	if err != nil {
		t.Fatalf("Erro ao criar parser: %v", err)
	}
	prog, err := parser.ParseProgram()
	if err != nil {
		t.Fatalf("Erro ao parsear programa: %v", err)
	}
	if prog.Bank != 2 {
		t.Fatalf("Esperado Bank=2, obteve %d", prog.Bank)
	}

	cg := NewCodeGenerator(prog)
	_, asmSource, err := cg.Compile()
	if err != nil {
		t.Fatalf("Erro ao compilar módulo com BANK 2: %v\nAssembly:\n%s", err, asmSource)
	}
	if !strings.Contains(asmSource, "BANK 2") {
		t.Errorf("Assembly gerado deveria conter 'BANK 2', obteve:\n%s", asmSource)
	}
}

// TestParseProgramWithoutBankDefaultsToZero é a regressão complementar: sem
// BANK declarado, o módulo continua caindo no banco comum, como sempre foi.
func TestParseProgramWithoutBankDefaultsToZero(t *testing.T) {
	src := `
	program Simple;
	var x: Integer;
	begin
		x := 1;
	end.
	`
	lexer := NewLexer(src)
	parser, err := NewParser(lexer)
	if err != nil {
		t.Fatalf("Erro ao criar parser: %v", err)
	}
	prog, err := parser.ParseProgram()
	if err != nil {
		t.Fatalf("Erro ao parsear programa: %v", err)
	}
	if prog.Bank != 0 {
		t.Errorf("Esperado Bank=0 (padrão) sem declaração BANK, obteve %d", prog.Bank)
	}
}

// TestLinkKaji80AndWirth80LibraryTogether é o teste end-to-end que confirma
// o objetivo final: um módulo KAJI80 que já define Start agora linka de
// verdade com um módulo-biblioteca WIRTH80 (sem Start próprio) no mesmo
// .COM -- o obstáculo real documentado em docs/manual-ferramentas.md §10
// antes desta correção.
func TestLinkKaji80AndWirth80LibraryTogether(t *testing.T) {
	libSrc := `
	program Lib;
	PUBLIC Foo;
	procedure Foo(a: Integer);
	begin
		WriteLn(a);
	end;
	begin
	end.
	`
	lexer := NewLexer(libSrc)
	parser, err := NewParser(lexer)
	if err != nil {
		t.Fatalf("Erro ao criar parser: %v", err)
	}
	prog, err := parser.ParseProgram()
	if err != nil {
		t.Fatalf("Erro ao parsear programa: %v", err)
	}
	libObj, _, err := NewCodeGenerator(prog).Compile()
	if err != nil {
		t.Fatalf("Erro ao compilar módulo-biblioteca: %v", err)
	}

	kaji80Src := `
	MODULE Main
	BANK 0
	PUBLIC Start
	EXTERN BDOS_Exit

	Start:
	    CALL BDOS_Exit
	ENDMOD
	`
	kaji80Asm := kaji80.NewAssembler()
	mainObj, err := kaji80Asm.Assemble(kaji80Src)
	if err != nil {
		t.Fatalf("Erro ao montar módulo KAJI80: %v", err)
	}

	libPath := "../../lib/msxlib.hlib"
	if _, err := os.Stat(libPath); os.IsNotExist(err) {
		t.Skipf("msxlib.hlib não encontrado em %s, pulando teste de linkagem", libPath)
		return
	}

	tmpDir := t.TempDir()
	mainMob := filepath.Join(tmpDir, "main.mob")
	libMob := filepath.Join(tmpDir, "lib.mob")
	if err := mob.SaveToFile(mainMob, mainObj); err != nil {
		t.Fatalf("Erro ao gravar %s: %v", mainMob, err)
	}
	if err := mob.SaveToFile(libMob, libObj); err != nil {
		t.Fatalf("Erro ao gravar %s: %v", libMob, err)
	}

	outCom := filepath.Join(tmpDir, "out.com")
	cfg := musubi.LinkerConfig{BaseAddress: 0x0100, EntryPoint: "Start"}
	res, err := musubi.LinkToFile(outCom, cfg, mainMob, libMob, libPath)
	if err != nil {
		t.Fatalf("Falha ao linkar KAJI80 (Start) + biblioteca WIRTH80 (sem Start) juntos: %v", err)
	}
	if len(res.Binary) == 0 {
		t.Fatalf("Executável binário gerado está vazio")
	}
}
