package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

	flag.Usage = printHelp
	flag.Parse()

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

	asm := kaji80.NewAssembler()
	obj, err := asm.Assemble(string(content))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro de montagem: %v\n", err)
		os.Exit(1)
	}

	err = mob.SaveToFile(*outPath, obj)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao salvar %s: %v\n", *outPath, err)
		os.Exit(1)
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
