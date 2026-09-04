package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/wilsonpilon/kizuna/pkg/mob"
	"github.com/wilsonpilon/kizuna/pkg/version"
	"github.com/wilsonpilon/kizuna/pkg/wirth80"
)

func printHelp() {
	fmt.Println(version.Banner("WIRTH80"))
	fmt.Print(`
USO:
    wirth80 [opções] <arquivo.pas>

OPÇÕES:
    -o <caminho>   Especifica o arquivo de saída .mob (padrão: mesmo nome com extensão .mob)
    -S             Emite o código Assembly Z80 (.asm) em vez de compilar diretamente para .mob
    -v             Modo detalhado (exibe resumo da AST, símbolos e código gerado)
    --version      Exibe a versão atual
    -h, --help     Exibe esta ajuda completa

EXEMPLO:
    wirth80 hello.pas
    wirth80 -S hello.pas -o hello.asm
    musubi hello.mob msxlib.hlib -o hello.com
`)
}

func main() {
	help := flag.Bool("help", false, "Exibe ajuda detalhada")
	shortHelp := flag.Bool("h", false, "Exibe ajuda detalhada")
	verFlag := flag.Bool("version", false, "Exibe versão do WIRTH80")
	outPath := flag.String("o", "", "Caminho do arquivo de saída .mob ou .asm")
	emitAsm := flag.Bool("S", false, "Emite arquivo Assembly Z80 intermediário (.asm)")
	verbose := flag.Bool("v", false, "Modo detalhado")

	flag.Parse()

	if *help || *shortHelp {
		printHelp()
		os.Exit(0)
	}

	if *verFlag {
		fmt.Printf("WIRTH80 v%s\n", version.FullVersion())
		os.Exit(0)
	}

	args := flag.Args()
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Erro: nenhum arquivo de entrada fornecido.")
		fmt.Fprintln(os.Stderr, "Use 'wirth80 --help' para instruções de uso.")
		os.Exit(1)
	}

	inputPath := args[0]
	content, err := os.ReadFile(inputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao ler arquivo de entrada '%s': %v\n", inputPath, err)
		os.Exit(1)
	}

	// 1. Lexer & Parser
	lexer := wirth80.NewLexer(string(content))
	parser, err := wirth80.NewParser(lexer)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao inicializar parser: %v\n", err)
		os.Exit(1)
	}

	prog, err := parser.ParseProgram()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro sintático em '%s':\n  %v\n", inputPath, err)
		os.Exit(1)
	}

	// 2. Codegen
	cg := wirth80.NewCodeGenerator(prog)

	// Se a flag -S estiver ativada, emite o código assembly (.asm)
	if *emitAsm {
		asmCode, err := cg.GenerateAsm()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Erro na geração de código Z80: %v\n", err)
			os.Exit(1)
		}

		targetOut := *outPath
		if targetOut == "" {
			ext := filepath.Ext(inputPath)
			targetOut = strings.TrimSuffix(inputPath, ext) + ".asm"
		}

		if err := os.WriteFile(targetOut, []byte(asmCode), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Erro ao gravar arquivo assembly '%s': %v\n", targetOut, err)
			os.Exit(1)
		}

		fmt.Printf("WIRTH80: Assembly Z80 gerado com sucesso -> %s\n", targetOut)
		if *verbose {
			fmt.Println("\n--- CÓDIGO ASSEMBLY GERADO ---")
			fmt.Println(asmCode)
		}
		os.Exit(0)
	}

	// Compilação completa para .MOB
	obj, asmCode, err := cg.Compile()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro de compilação em '%s':\n  %v\n", inputPath, err)
		os.Exit(1)
	}

	targetOut := *outPath
	if targetOut == "" {
		ext := filepath.Ext(inputPath)
		targetOut = strings.TrimSuffix(inputPath, ext) + ".mob"
	}

	if err := mob.SaveToFile(targetOut, obj); err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao salvar arquivo objeto '%s': %v\n", targetOut, err)
		os.Exit(1)
	}

	fmt.Printf("WIRTH80: %s compilado com sucesso -> %s\n", inputPath, targetOut)
	if *verbose {
		fmt.Printf("  Programa:     %s\n", prog.Name)
		fmt.Printf("  Segmentos:    %d\n", len(obj.Segments))
		fmt.Printf("  Símbolos:     %d\n", len(obj.Symbols))
		fmt.Printf("  Relocações:   %d\n", len(obj.Relocations))
		if len(obj.Segments) > 0 {
			seg := obj.Segments[0]
			fmt.Printf("  Seg[0]: %s, Banco %d, Tamanho: %d bytes\n", seg.Type, seg.Bank, len(seg.Data))
		}
		fmt.Println("\n--- CÓDIGO ASSEMBLY GERADO INTERNAMENTE ---")
		fmt.Println(asmCode)
	}
}
