package wirth80

import (
	"strings"
	"testing"

	"github.com/wilsonpilon/kizuna/pkg/mob"
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
