package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/wilsonpilon/kizuna/pkg/api"
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
    --log          Gera arquivo de log da compilação (<arquivo>.log)
    --log-file <f> Especifica caminho customizado para o arquivo de log
    -api <f|dir>   Descritor(es) de API da MSXLIB (.api, arquivo ou diretório; repetível).
                   Sem -api, usa ../lib/api ao lado do executável, se existir.
    -v             Modo detalhado (exibe resumo da AST, símbolos e código gerado)
    --version      Exibe a versão atual
    -h, --help     Exibe esta ajuda completa

EXEMPLO:
    wirth80 hello.pas
    wirth80 --log hello.pas
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
	logFlag := flag.Bool("log", false, "Gera arquivo de log da compilação (<arquivo>.log)")
	logPath := flag.String("log-file", "", "Especifica caminho customizado para o log")
	var apiPaths api.PathList
	flag.Var(&apiPaths, "api", "Descritor(es) de API da MSXLIB (.api, arquivo ou diretório; repetível)")

	_ = flag.CommandLine.Parse(rearrangeArgs(os.Args[1:]))

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
	apiSet, err := api.ForCompiler(apiPaths, os.Args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro nos descritores de API: %v\n", err)
		os.Exit(1)
	}
	cg.SetAPI(apiSet)

	getLogDest := func(defaultOut string) string {
		if *logPath != "" {
			return *logPath
		}
		if *logFlag {
			ext := filepath.Ext(inputPath)
			return strings.TrimSuffix(inputPath, ext) + ".log"
		}
		return ""
	}

	// Se a flag -S estiver ativada, emite o código assembly (.asm)
	if *emitAsm {
		targetOut := *outPath
		if targetOut == "" {
			ext := filepath.Ext(inputPath)
			targetOut = strings.TrimSuffix(inputPath, ext) + ".asm"
		}

		asmCode, err := cg.GenerateAsm()
		if err != nil {
			if logDest := getLogDest(targetOut); logDest != "" {
				_ = writeCompilationLog(logDest, inputPath, targetOut, prog, nil, "", err, *logPath != "")
			}
			fmt.Fprintf(os.Stderr, "Erro na geração de código Z80: %v\n", err)
			os.Exit(1)
		}

		if err := os.WriteFile(targetOut, []byte(asmCode), 0644); err != nil {
			if logDest := getLogDest(targetOut); logDest != "" {
				_ = writeCompilationLog(logDest, inputPath, targetOut, prog, nil, asmCode, err, *logPath != "")
			}
			fmt.Fprintf(os.Stderr, "Erro ao gravar arquivo assembly '%s': %v\n", targetOut, err)
			os.Exit(1)
		}

		if logDest := getLogDest(targetOut); logDest != "" {
			if logErr := writeCompilationLog(logDest, inputPath, targetOut, prog, nil, asmCode, nil, *logPath != ""); logErr == nil {
				fmt.Printf("WIRTH80: Log gravado com sucesso -> %s\n", logDest)
			}
		}

		fmt.Printf("WIRTH80: Assembly Z80 gerado com sucesso -> %s\n", targetOut)
		if *verbose {
			fmt.Println("\n--- CÓDIGO ASSEMBLY GERADO ---")
			fmt.Println(asmCode)
		}
		os.Exit(0)
	}

	targetOut := *outPath
	if targetOut == "" {
		ext := filepath.Ext(inputPath)
		targetOut = strings.TrimSuffix(inputPath, ext) + ".mob"
	}

	// Compilação completa para .MOB
	obj, asmCode, err := cg.Compile()
	if err != nil {
		if logDest := getLogDest(targetOut); logDest != "" {
			_ = writeCompilationLog(logDest, inputPath, targetOut, prog, nil, "", err, *logPath != "")
		}
		fmt.Fprintf(os.Stderr, "Erro de compilação em '%s':\n  %v\n", inputPath, err)
		os.Exit(1)
	}

	if err := mob.SaveToFile(targetOut, obj); err != nil {
		if logDest := getLogDest(targetOut); logDest != "" {
			_ = writeCompilationLog(logDest, inputPath, targetOut, prog, obj, asmCode, err, *logPath != "")
		}
		fmt.Fprintf(os.Stderr, "Erro ao salvar arquivo objeto '%s': %v\n", targetOut, err)
		os.Exit(1)
	}

	if logDest := getLogDest(targetOut); logDest != "" {
		if logErr := writeCompilationLog(logDest, inputPath, targetOut, prog, obj, asmCode, nil, *logPath != ""); logErr == nil {
			fmt.Printf("WIRTH80: Log gravado com sucesso -> %s\n", logDest)
		}
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

func writeCompilationLog(logPath string, inputPath string, outPath string, prog *wirth80.ProgramNode, obj *mob.ObjectFile, asmCode string, compErr error, appendMode bool) error {
	var sb strings.Builder
	sb.WriteString("================================================================================\n")
	sb.WriteString("              KIZUNA TOOLCHAIN - WIRTH80 COMPILATION LOG\n")
	sb.WriteString("================================================================================\n")
	sb.WriteString(fmt.Sprintf("Timestamp:    %s\n", time.Now().Format("2006-01-02 15:04:05")))
	sb.WriteString(fmt.Sprintf("Compiler:     WIRTH80 v%s\n", version.FullVersion()))
	sb.WriteString(fmt.Sprintf("Source File:  %s\n", inputPath))
	sb.WriteString(fmt.Sprintf("Output File:  %s\n", outPath))
	if compErr != nil {
		sb.WriteString(fmt.Sprintf("Status:       FAILED (%v)\n", compErr))
	} else {
		sb.WriteString("Status:       SUCCESS\n")
	}
	sb.WriteString("\n--------------------------------------------------------------------------------\n")
	sb.WriteString("PROGRAM INFORMATION\n")
	sb.WriteString("--------------------------------------------------------------------------------\n")
	if prog != nil {
		sb.WriteString(fmt.Sprintf("Program Name: %s\n", prog.Name))
		sb.WriteString(fmt.Sprintf("Variables:    %d\n", len(prog.Vars)))
		if prog.Block != nil {
			sb.WriteString(fmt.Sprintf("Statements:   %d\n", len(prog.Block.Statements)))
		}
	} else {
		sb.WriteString("Program AST not available.\n")
	}

	if obj != nil {
		sb.WriteString("\n--------------------------------------------------------------------------------\n")
		sb.WriteString("OBJECT FILE DETAILS (.MOB)\n")
		sb.WriteString("--------------------------------------------------------------------------------\n")
		sb.WriteString(fmt.Sprintf("Segments (%d):\n", len(obj.Segments)))
		for i, seg := range obj.Segments {
			sb.WriteString(fmt.Sprintf("  [%d] Type: %-6s Bank: %-2d Size: %4d bytes\n",
				i, seg.Type.String(), seg.Bank, seg.Size))
		}
		sb.WriteString(fmt.Sprintf("\nSymbols (%d):\n", len(obj.Symbols)))
		for _, sym := range obj.Symbols {
			sb.WriteString(fmt.Sprintf("  - %-20s %-7s %-6s Offset: %04Xh (Seg %d)\n",
				sym.Name, sym.Class.String(), sym.Kind.String(), sym.Offset, sym.SegmentIndex))
		}
		sb.WriteString(fmt.Sprintf("\nRelocations (%d):\n", len(obj.Relocations)))
		for i, rel := range obj.Relocations {
			symName := "N/A"
			if int(rel.SymbolIndex) < len(obj.Symbols) {
				symName = obj.Symbols[rel.SymbolIndex].Name
			}
			sb.WriteString(fmt.Sprintf("  [%2d] Offset: %04Xh Target: %-16s Type: %s (Seg %d)\n",
				i, rel.Offset, symName, rel.Type.String(), rel.SegmentIndex))
		}
	}

	if asmCode != "" {
		sb.WriteString("\n--------------------------------------------------------------------------------\n")
		sb.WriteString("GENERATED Z80 ASSEMBLY CODE\n")
		sb.WriteString("--------------------------------------------------------------------------------\n")
		sb.WriteString(asmCode)
		sb.WriteString("\n")
	}
	sb.WriteString("================================================================================\n")
	if appendMode {
		f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		defer f.Close()
		stat, err := f.Stat()
		if err == nil && stat.Size() > 0 {
			_, _ = f.WriteString("\n")
		}
		_, err = f.WriteString(sb.String())
		return err
	}

	return os.WriteFile(logPath, []byte(sb.String()), 0644)
}

func rearrangeArgs(args []string) []string {
	var flags []string
	var nonFlags []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "-") {
			flags = append(flags, arg)
			if (arg == "-o" || arg == "--log-file" || arg == "-log-file" || arg == "-api" || arg == "--api") && i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				i++
				flags = append(flags, args[i])
			}
		} else {
			nonFlags = append(nonFlags, arg)
		}
	}
	return append(flags, nonFlags...)
}
