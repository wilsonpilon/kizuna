package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/wilsonpilon/kizuna/pkg/dignac"
	"github.com/wilsonpilon/kizuna/pkg/mob"
	"github.com/wilsonpilon/kizuna/pkg/version"
)

func printHelp() {
	fmt.Println(version.Banner("DIGNAC"))
	fmt.Print(`
USO:
    dignac [opções] <arquivo.bas>

OPÇÕES:
    -o <caminho>   Especifica o arquivo de saída .mob (padrão: mesmo nome com extensão .mob)
    -S             Emite o código Assembly Z80 (.asm) em vez de compilar diretamente para .mob
    --log          Gera arquivo de log da compilação (<arquivo>.log)
    --log-file <f> Especifica caminho customizado para o arquivo de log
    -v             Modo detalhado (exibe resumo da AST, símbolos e código gerado)
    --version      Exibe a versão atual
    -h, --help     Exibe esta ajuda completa

EXEMPLO:
    dignac chart.bas
    dignac --log chart.bas
    dignac -S chart.bas -o chart.asm
    musubi chart.mob msxlib.hlib -o chart.com
`)
}

func main() {
	help := flag.Bool("help", false, "Exibe ajuda detalhada")
	shortHelp := flag.Bool("h", false, "Exibe ajuda detalhada")
	verFlag := flag.Bool("version", false, "Exibe versão do DIGNAC")
	outPath := flag.String("o", "", "Caminho do arquivo de saída .mob ou .asm")
	emitAsm := flag.Bool("S", false, "Emite arquivo Assembly Z80 intermediário (.asm)")
	verbose := flag.Bool("v", false, "Modo detalhado")
	logFlag := flag.Bool("log", false, "Gera arquivo de log da compilação (<arquivo>.log)")
	logPath := flag.String("log-file", "", "Especifica caminho customizado para o log")

	_ = flag.CommandLine.Parse(rearrangeArgs(os.Args[1:]))

	if *help || *shortHelp {
		printHelp()
		os.Exit(0)
	}

	if *verFlag {
		fmt.Printf("DIGNAC v%s\n", version.FullVersion())
		os.Exit(0)
	}

	args := flag.Args()
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Erro: nenhum arquivo de entrada fornecido.")
		fmt.Fprintln(os.Stderr, "Use 'dignac --help' para instruções de uso.")
		os.Exit(1)
	}

	inputPath := args[0]
	content, err := os.ReadFile(inputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao ler arquivo de entrada '%s': %v\n", inputPath, err)
		os.Exit(1)
	}

	// 1. Lexer & Parser
	lexer := dignac.NewLexer(string(content))
	parser, err := dignac.NewParser(lexer)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao inicializar parser: %v\n", err)
		os.Exit(1)
	}

	mod, err := parser.ParseModule()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro sintático em '%s':\n  %v\n", inputPath, err)
		os.Exit(1)
	}

	// 2. Codegen
	cg := dignac.NewCodeGenerator(mod)

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
				_ = writeCompilationLog(logDest, inputPath, targetOut, mod, nil, "", err, *logPath != "")
			}
			fmt.Fprintf(os.Stderr, "Erro na geração de código Z80: %v\n", err)
			os.Exit(1)
		}

		if err := os.WriteFile(targetOut, []byte(asmCode), 0644); err != nil {
			if logDest := getLogDest(targetOut); logDest != "" {
				_ = writeCompilationLog(logDest, inputPath, targetOut, mod, nil, asmCode, err, *logPath != "")
			}
			fmt.Fprintf(os.Stderr, "Erro ao gravar arquivo assembly '%s': %v\n", targetOut, err)
			os.Exit(1)
		}

		if logDest := getLogDest(targetOut); logDest != "" {
			if logErr := writeCompilationLog(logDest, inputPath, targetOut, mod, nil, asmCode, nil, *logPath != ""); logErr == nil {
				fmt.Printf("DIGNAC: Log gravado com sucesso -> %s\n", logDest)
			}
		}

		fmt.Printf("DIGNAC: Assembly Z80 gerado com sucesso -> %s\n", targetOut)
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
			_ = writeCompilationLog(logDest, inputPath, targetOut, mod, nil, "", err, *logPath != "")
		}
		fmt.Fprintf(os.Stderr, "Erro de compilação em '%s':\n  %v\n", inputPath, err)
		os.Exit(1)
	}

	if err := mob.SaveToFile(targetOut, obj); err != nil {
		if logDest := getLogDest(targetOut); logDest != "" {
			_ = writeCompilationLog(logDest, inputPath, targetOut, mod, obj, asmCode, err, *logPath != "")
		}
		fmt.Fprintf(os.Stderr, "Erro ao salvar arquivo objeto '%s': %v\n", targetOut, err)
		os.Exit(1)
	}

	if logDest := getLogDest(targetOut); logDest != "" {
		if logErr := writeCompilationLog(logDest, inputPath, targetOut, mod, obj, asmCode, nil, *logPath != ""); logErr == nil {
			fmt.Printf("DIGNAC: Log gravado com sucesso -> %s\n", logDest)
		}
	}

	fmt.Printf("DIGNAC: %s compilado com sucesso -> %s\n", inputPath, targetOut)
	if *verbose {
		fmt.Printf("  Módulo:       %s\n", mod.Name)
		fmt.Printf("  Banco:        %d\n", mod.Bank)
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

func writeCompilationLog(logPath string, inputPath string, outPath string, mod *dignac.ModuleNode, obj *mob.ObjectFile, asmCode string, compErr error, appendMode bool) error {
	var sb strings.Builder
	sb.WriteString("================================================================================\n")
	sb.WriteString("              KIZUNA TOOLCHAIN - DIGNAC COMPILATION LOG\n")
	sb.WriteString("================================================================================\n")
	sb.WriteString(fmt.Sprintf("Timestamp:    %s\n", time.Now().Format("2006-01-02 15:04:05")))
	sb.WriteString(fmt.Sprintf("Compiler:     DIGNAC v%s\n", version.FullVersion()))
	sb.WriteString(fmt.Sprintf("Source File:  %s\n", inputPath))
	sb.WriteString(fmt.Sprintf("Output File:  %s\n", outPath))
	if compErr != nil {
		sb.WriteString(fmt.Sprintf("Status:       FAILED (%v)\n", compErr))
	} else {
		sb.WriteString("Status:       SUCCESS\n")
	}
	sb.WriteString("\n--------------------------------------------------------------------------------\n")
	sb.WriteString("MODULE INFORMATION\n")
	sb.WriteString("--------------------------------------------------------------------------------\n")
	if mod != nil {
		sb.WriteString(fmt.Sprintf("Module Name:  %s\n", mod.Name))
		sb.WriteString(fmt.Sprintf("Target Bank:  %d\n", mod.Bank))
		sb.WriteString(fmt.Sprintf("Publics:      %s\n", strings.Join(mod.Publics, ", ")))
		sb.WriteString(fmt.Sprintf("Externs:      %s\n", strings.Join(mod.Externs, ", ")))
		sb.WriteString(fmt.Sprintf("Procedures:   %d\n", len(mod.Procedures)))
		for _, proc := range mod.Procedures {
			kind := "PROCEDURE"
			if proc.IsFunction {
				kind = "FUNCTION"
			}
			var params []string
			for _, p := range proc.Params {
				params = append(params, fmt.Sprintf("%s AS %s", p.Name, p.Type))
			}
			ret := ""
			if proc.IsFunction && proc.ReturnType != "" {
				ret = " AS " + proc.ReturnType
			}
			sb.WriteString(fmt.Sprintf("  - %s %s(%s)%s\n", kind, proc.Name, strings.Join(params, ", "), ret))
		}
	} else {
		sb.WriteString("Module AST not available.\n")
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
			if (arg == "-o" || arg == "--log-file" || arg == "-log-file") && i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				i++
				flags = append(flags, args[i])
			}
		} else {
			nonFlags = append(nonFlags, arg)
		}
	}
	return append(flags, nonFlags...)
}
