package dignac

import (
	"os"
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
	if len(proc.Locals) != 1 || proc.Locals[0].Vars[0] != "r%" {
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

		// Salva temporariamente .mob
		mobPath := strings.TrimSuffix(path, ".bas") + ".mob"
		if err := mob.SaveToFile(mobPath, obj); err != nil {
			t.Fatalf("Erro ao gravar %s: %v", mobPath, err)
		}

		// Linka com msxlib.hlib via MUSUBI
		outCom := strings.TrimSuffix(path, ".bas") + ".com"
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
