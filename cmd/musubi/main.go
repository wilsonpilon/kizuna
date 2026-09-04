package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/wilsonpilon/kizuna/pkg/musubi"
	"github.com/wilsonpilon/kizuna/pkg/version"
)

func printHelp() {
	fmt.Println(version.Banner("MUSUBI"))
	fmt.Print(`
USO:
    musubi [opções] <objeto.mob...> [biblioteca.hlib...]

OPÇÕES:
    -o <caminho>   Especifica o executável de saída .com (padrão: baseado no primeiro objeto)
    -m <caminho>   Gera relatório de mapa de memória e tabela de símbolos (.map)
    -b <endereço>  Endereço base de carregamento (padrão: 0x0100 para MSX-DOS)
    -e <símbolo>   Ponto de entrada do programa (padrão: "Start")
    -v             Modo detalhado (exibe relatório do binário e endereços na tela)
    --version      Exibe a versão atual
    -h, --help     Exibe esta ajuda completa

EXEMPLO:
    # Linkar um único objeto gerando executável .com para MSX-DOS
    musubi -v -m hello.map -o hello.com hello.mob

    # Linkar múltiplos módulos
    musubi -o app.com main.mob screen.mob math.mob

    # Linkar com biblioteca estática .HLIB (Smart-Linking)
    musubi -o app.com main.mob math.hlib
`)
}

func main() {
	help := flag.Bool("help", false, "Exibe ajuda detalhada")
	shortHelp := flag.Bool("h", false, "Exibe ajuda detalhada")
	verFlag := flag.Bool("version", false, "Exibe versão do MUSUBI")
	outPath := flag.String("o", "", "Caminho do executável de saída .com")
	mapPath := flag.String("m", "", "Caminho do relatório de mapa de memória .map")
	baseAddrStr := flag.String("b", "0x0100", "Endereço base de execução")
	entryPoint := flag.String("e", "Start", "Símbolo do ponto de entrada")
	verbose := flag.Bool("v", false, "Modo detalhado")

	flag.Usage = printHelp
	flag.Parse()

	if *verFlag {
		fmt.Println(version.Banner("MUSUBI"))
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

	// Converter endereço base
	baseAddrVal, err := strconv.ParseUint(strings.TrimPrefix(*baseAddrStr, "0x"), 16, 16)
	if err != nil {
		baseVal64, errDec := strconv.ParseUint(*baseAddrStr, 10, 16)
		if errDec != nil {
			fmt.Fprintf(os.Stderr, "Erro: endereço base inválido '%s'\n", *baseAddrStr)
			os.Exit(1)
		}
		baseAddrVal = baseVal64
	}

	if *outPath == "" {
		ext := filepath.Ext(args[0])
		*outPath = strings.TrimSuffix(args[0], ext) + ".com"
	}

	cfg := musubi.LinkerConfig{
		BaseAddress: uint16(baseAddrVal),
		EntryPoint:  *entryPoint,
		MapFile:     *mapPath,
		Verbose:     *verbose,
	}

	res, err := musubi.LinkToFile(*outPath, cfg, args...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro de linkagem: %v\n", err)
		os.Exit(1)
	}

	if *verbose {
		fmt.Printf("MUSUBI: linkagem concluída com sucesso -> %s\n", *outPath)
		fmt.Printf("  Endereço Base:   0x%04X\n", res.BaseAddress)
		fmt.Printf("  Ponto de Entrada: 0x%04X (%s)\n", res.EntryPoint, *entryPoint)
		fmt.Printf("  Tamanho do .COM: %d bytes (0x%04X)\n", res.TotalSize, res.TotalSize)
		fmt.Printf("  Total Símbolos:  %d\n", len(res.Symbols))
		if *mapPath != "" {
			fmt.Printf("  Mapa gerado em:  %s\n", *mapPath)
		}
	}
}
