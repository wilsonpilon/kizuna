package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/wilsonpilon/kizuna/pkg/kaji80"
	"github.com/wilsonpilon/kizuna/pkg/mob"
	"github.com/wilsonpilon/kizuna/pkg/version"
)

func printHelp() {
	fmt.Println(version.Banner("KAJI80"))
	fmt.Print(`
USO:
    kaji80 [opções] <arquivo.asm>

OPÇÕES:
    -o <caminho>   Especifica o arquivo de saída .mob (padrão: mesmo nome com extensão .mob)
    --log          Gera ou anexa ao arquivo de log da compilação (<arquivo>.log)
    --log-file <f> Especifica caminho customizado para o arquivo de log
    -v             Modo detalhado (exibe resumo de segmentos, símbolos e relocações)
    --version      Exibe a versão atual
    -h, --help     Exibe esta ajuda completa

DIRETIVAS SUPORTADAS:
    MODULE <nome>              Define o nome do módulo
    BANK <n>                   Define o banco de memória alvo (0 = área comum, 1..N = banco paginável)
    PUBLIC <sym1> [, <sym2>]   Exporta símbolos para uso por outros módulos ou pelo linker
    EXTERN <sym1> [, <sym2>]   Declara símbolos externos importados de outros módulos
    <nome> EQU <valor>         Define uma constante simbólica
    ORG <endereço>             Define a origem/offset base do segmento
    DB <bytes/strings>         Emite bytes e/ou strings de texto (alias: DEFB, BYTE)
    DW <words>                 Emite words de 16 bits em Little-Endian (alias: DEFW, WORD)
    DS <tamanho>               Reserva bytes preenchidos com zero (alias: DEFS, BLKB)
    ENDMOD / END               Marca o fim do módulo

FORMATOS NUMÉRICOS ACEITOS:
    Hexadecimal: 0x100, 100h, 100H, $100, #100
    Binário:     %1010, 0b1010, 1010b
    Decimal:     255, 42

INSTRUÇÕES Z80 SUPORTADAS:
    Controle:    NOP, HALT, DI, EI, EXX, EX DE, HL, EX AF, AF'
    Fluxo:       RET, RET cc, CALL nn, CALL cc, nn, JP nn, JP cc, nn, JP (HL|IX|IY), JR e, JR cc, e, DJNZ e
    Pilha:       PUSH rr, POP rr (BC, DE, HL, AF, IX, IY)
    I/O:         IN A, (n), OUT (n), A
    Aritmética:  INC, DEC, ADD, ADC, SUB, SBC, AND, XOR, OR, CP
    Dados:       LD r, r' | LD r, n | LD r, (HL) | LD (HL), r | LD (HL), n | LD rr, nn | LD (nn), A | LD A, (nn) | LD SP, HL
`)
}

func main() {
	help := flag.Bool("help", false, "Exibe ajuda detalhada")
	shortHelp := flag.Bool("h", false, "Exibe ajuda detalhada")
	verFlag := flag.Bool("version", false, "Exibe versão do KAJI80")
	outPath := flag.String("o", "", "Caminho do arquivo de saída .mob")
	verbose := flag.Bool("v", false, "Modo detalhado")
	logFlag := flag.Bool("log", false, "Gera ou anexa ao arquivo de log (<arquivo>.log)")
	logPath := flag.String("log-file", "", "Especifica caminho customizado para o log")

	flag.Usage = printHelp
	_ = flag.CommandLine.Parse(rearrangeArgs(os.Args[1:]))

	if *verFlag {
		fmt.Println(version.Banner("KAJI80"))
		os.Exit(0)
	}

	if *help || *shortHelp {
		printHelp()
		os.Exit(0)
	}

	args := flag.Args()
	if len(args) < 1 {
		printHelp()
		os.Exit(1)
	}

	inPath := args[0]
	content, err := os.ReadFile(inPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao ler %s: %v\n", inPath, err)
		os.Exit(1)
	}

	if *outPath == "" {
		ext := filepath.Ext(inPath)
		*outPath = strings.TrimSuffix(inPath, ext) + ".mob"
	}

	getLogDest := func() string {
		if *logPath != "" {
			return *logPath
		}
		if *logFlag {
			ext := filepath.Ext(inPath)
			return strings.TrimSuffix(inPath, ext) + ".log"
		}
		return ""
	}

	asm := kaji80.NewAssembler()
	obj, err := asm.Assemble(string(content))
	if err != nil {
		if logDest := getLogDest(); logDest != "" {
			_ = appendAssemblerLog(logDest, inPath, *outPath, nil, err)
		}
		fmt.Fprintf(os.Stderr, "Erro de montagem: %v\n", err)
		os.Exit(1)
	}

	err = mob.SaveToFile(*outPath, obj)
	if err != nil {
		if logDest := getLogDest(); logDest != "" {
			_ = appendAssemblerLog(logDest, inPath, *outPath, obj, err)
		}
		fmt.Fprintf(os.Stderr, "Erro ao salvar %s: %v\n", *outPath, err)
		os.Exit(1)
	}

	if logDest := getLogDest(); logDest != "" {
		if logErr := appendAssemblerLog(logDest, inPath, *outPath, obj, nil); logErr == nil {
			fmt.Printf("KAJI80: Log gravado com sucesso -> %s\n", logDest)
		}
	}

	if *verbose {
		fmt.Printf("KAJI80: %s montado com sucesso -> %s\n", inPath, *outPath)
		fmt.Printf("  Segmentos:    %d\n", len(obj.Segments))
		fmt.Printf("  Símbolos:     %d\n", len(obj.Symbols))
		fmt.Printf("  Relocações:   %d\n", len(obj.Relocations))
		for i, seg := range obj.Segments {
			fmt.Printf("  Seg[%d]: %s, Banco %d, Tamanho: %d bytes\n", i, seg.Type, seg.Bank, seg.Size)
		}
	}
}

func appendAssemblerLog(logPath string, inPath string, outPath string, obj *mob.ObjectFile, asmErr error) error {
	var sb strings.Builder
	sb.WriteString("--------------------------------------------------------------------------------\n")
	sb.WriteString("              KIZUNA TOOLCHAIN - KAJI80 ASSEMBLER LOG\n")
	sb.WriteString("--------------------------------------------------------------------------------\n")
	sb.WriteString(fmt.Sprintf("Timestamp:    %s\n", time.Now().Format("2006-01-02 15:04:05")))
	sb.WriteString(fmt.Sprintf("Assembler:    KAJI80 v%s\n", version.FullVersion()))
	sb.WriteString(fmt.Sprintf("Source File:  %s\n", inPath))
	sb.WriteString(fmt.Sprintf("Output File:  %s\n", outPath))
	if asmErr != nil {
		sb.WriteString(fmt.Sprintf("Status:       FAILED (%v)\n", asmErr))
	} else {
		sb.WriteString("Status:       SUCCESS\n")
	}

	if obj != nil {
		sb.WriteString("\nOBJECT FILE DETAILS (.MOB)\n")
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
	sb.WriteString("--------------------------------------------------------------------------------\n")

	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	stat, err := f.Stat()
	if err == nil && stat.Size() == 0 {
		_, _ = f.WriteString("================================================================================\n" +
			"                    KIZUNA TOOLCHAIN COMPILATION LOG\n" +
			"================================================================================\n\n")
	} else {
		_, _ = f.WriteString("\n")
	}

	_, err = f.WriteString(sb.String())
	return err
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
