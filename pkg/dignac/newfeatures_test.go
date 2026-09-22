package dignac

import (
	"strings"
	"testing"
)

func TestCompileSpriteMusicFileFeatures(t *testing.T) {
	src := `
MODULE NewFeatures
BANK 0
PUBLIC Main

PROCEDURE Main()
    SPRITE PATTERN 0, 255, 129, 129, 129, 129, 129, 129, 255
    PUT SPRITE 0, (10, 20), 15, 0
    SPRITE OFF

    PLAY "O4 CDEFGAB >C"

    OPEN "TEST.TXT" FOR OUTPUT AS #1
    PRINT #1, "Ola"; 42
    CLOSE #1
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
		t.Fatalf("Erro ao analisar módulo: %v", err)
	}

	cg := NewCodeGenerator(mod)
	asmCode, err := cg.GenerateAsm()
	if err != nil {
		t.Fatalf("Erro ao gerar assembly: %v", err)
	}

	for _, want := range []string{
		"CALL VDP_SpriteDefine",
		"CALL VDP_SpriteSet",
		"CALL VDP_SpriteHideAll",
		"CALL PSG_PlaySequence",
		"CALL BDOS_FileCreate",
		"CALL BDOS_FileWrite",
		"CALL BDOS_FileClose",
		"CALL PrintDec16ToBuffer",
		"DGN_FileHandle_1:",
		"EXTERN",
	} {
		if !strings.Contains(asmCode, want) {
			t.Errorf("Assembly gerado deveria conter %q, código:\n%s", want, asmCode)
		}
	}

	// Compilação completa até .MOB via KAJI80 -- pega qualquer forma de
	// instrução que o Assembler não aceite (mesma rede de segurança que
	// TestCompileChartDemo já usa para o resto do DIGNAC).
	obj, asmOut, err := cg.Compile()
	if err != nil {
		t.Fatalf("Erro ao compilar para .MOB: %v\nCódigo emitido:\n%s", err, asmOut)
	}
	if obj == nil || len(obj.Segments) == 0 {
		t.Fatalf("Objeto .MOB gerado é inválido")
	}
}

func TestPlayStmtRejectsInvalidMML(t *testing.T) {
	src := `
MODULE BadPlay
BANK 0
PUBLIC Main
PROCEDURE Main()
    PLAY "O9 C"
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
		t.Fatalf("Erro ao analisar módulo: %v", err)
	}
	cg := NewCodeGenerator(mod)
	if _, err := cg.GenerateAsm(); err == nil {
		t.Fatal("esperava erro de compilação para MML com oitava inválida (O9)")
	}
}

func TestPlayStmtRequiresStringLiteral(t *testing.T) {
	src := `
MODULE BadPlay2
BANK 0
PUBLIC Main
PROCEDURE Main()
    LOCAL s%
    PLAY s%
END PROCEDURE
END MODULE
`
	lexer := NewLexer(src)
	parser, err := NewParser(lexer)
	if err != nil {
		t.Fatalf("Erro ao criar parser: %v", err)
	}
	if _, err := parser.ParseModule(); err == nil {
		t.Fatal("esperava erro de parse: PLAY exige literal de string")
	}
}

func TestSpritePatternRequiresValidByteCount(t *testing.T) {
	src := `
MODULE BadPattern
BANK 0
PUBLIC Main
PROCEDURE Main()
    SPRITE PATTERN 0, 1, 2, 3
END PROCEDURE
END MODULE
`
	lexer := NewLexer(src)
	parser, err := NewParser(lexer)
	if err != nil {
		t.Fatalf("Erro ao criar parser: %v", err)
	}
	if _, err := parser.ParseModule(); err == nil {
		t.Fatal("esperava erro de parse: SPRITE PATTERN precisa de 8 ou 32 bytes")
	}
}

func TestOpenRequiresPathLiteral(t *testing.T) {
	src := `
MODULE BadOpen
BANK 0
PUBLIC Main
PROCEDURE Main()
    LOCAL p%
    OPEN p% FOR OUTPUT AS #1
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
		t.Fatalf("Erro ao analisar módulo: %v", err)
	}
	cg := NewCodeGenerator(mod)
	if _, err := cg.GenerateAsm(); err == nil {
		t.Fatal("esperava erro de compilação: OPEN exige caminho literal")
	}
}
