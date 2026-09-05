package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

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
    --log          Gera ou anexa ao arquivo de log da compilação (<executavel>.log)
    --log-file <f> Especifica caminho customizado para o arquivo de log
    -b <endereço>  Endereço base de carregamento (padrão: 0x0100 para MSX-DOS)
    -e <símbolo>   Ponto de entrada do programa (padrão: "Start")
    -v             Modo detalhado (exibe relatório do binário e endereços na tela)
    --version      Exibe a versão atual
    -h, --help     Exibe esta ajuda completa

EXEMPLO:
    # Linkar um único objeto gerando executável .com para MSX-DOS
    musubi -v -m hello.map -o hello.com hello.mob

    # Linkar com geração/anexo de log
    musubi --log -o hello.com hello.mob msxlib.hlib

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
	logFlag := flag.Bool("log", false, "Gera ou anexa ao arquivo de log da compilação (<executavel>.log)")
	logPath := flag.String("log-file", "", "Especifica caminho customizado para o log")

	flag.Usage = printHelp
	_ = flag.CommandLine.Parse(rearrangeArgs(os.Args[1:]))

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

	getLogDest := func() string {
		if *logPath != "" {
			return *logPath
		}
		if *logFlag {
			if *outPath != "" {
				ext := filepath.Ext(*outPath)
				return strings.TrimSuffix(*outPath, ext) + ".log"
			}
			if len(args) > 0 {
				ext := filepath.Ext(args[0])
				return strings.TrimSuffix(args[0], ext) + ".log"
			}
			return "musubi.log"
		}
		return ""
	}

	cfg := musubi.LinkerConfig{
		BaseAddress: uint16(baseAddrVal),
		EntryPoint:  *entryPoint,
		MapFile:     *mapPath,
		Verbose:     *verbose,
	}

	res, err := musubi.LinkToFile(*outPath, cfg, args...)
	if err != nil {
		if logDest := getLogDest(); logDest != "" {
			_ = appendLinkerLog(logDest, args, *outPath, cfg, nil, err)
		}
		fmt.Fprintf(os.Stderr, "Erro de linkagem: %v\n", err)
		os.Exit(1)
	}

	if logDest := getLogDest(); logDest != "" {
		if logErr := appendLinkerLog(logDest, args, *outPath, cfg, res, nil); logErr == nil {
			fmt.Printf("MUSUBI: Log gravado com sucesso -> %s\n", logDest)
		}
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

func appendLinkerLog(logPath string, inputs []string, outPath string, cfg musubi.LinkerConfig, res *musubi.LinkResult, linkErr error) error {
	var sb strings.Builder
	sb.WriteString("--------------------------------------------------------------------------------\n")
	sb.WriteString("              KIZUNA TOOLCHAIN - MUSUBI LINKER LOG\n")
	sb.WriteString("--------------------------------------------------------------------------------\n")
	sb.WriteString(fmt.Sprintf("Timestamp:        %s\n", time.Now().Format("2006-01-02 15:04:05")))
	sb.WriteString(fmt.Sprintf("Linker:           MUSUBI v%s\n", version.FullVersion()))
	sb.WriteString(fmt.Sprintf("Input Files:      %s\n", strings.Join(inputs, ", ")))
	sb.WriteString(fmt.Sprintf("Output Binary:    %s\n", outPath))
	if cfg.MapFile != "" {
		sb.WriteString(fmt.Sprintf("Map File:         %s\n", cfg.MapFile))
	}
	if linkErr != nil {
		sb.WriteString(fmt.Sprintf("Status:           FAILED (%v)\n", linkErr))
	} else {
		sb.WriteString("Status:           SUCCESS\n")
	}

	if res != nil {
		sb.WriteString("\nLINK DETAILS:\n")
		sb.WriteString(fmt.Sprintf("Base Address:     0x%04X\n", res.BaseAddress))
		sb.WriteString(fmt.Sprintf("Entry Point:      0x%04X (%s)\n", res.EntryPoint, cfg.EntryPoint))
		sb.WriteString(fmt.Sprintf("Total Binary Size:%d bytes (0x%04X)\n", res.TotalSize, res.TotalSize))
		if res.IsMultiBank {
			var bStrs []string
			for _, b := range res.BanksUsed {
				bStrs = append(bStrs, fmt.Sprintf("%d", b))
			}
			sb.WriteString(fmt.Sprintf("Multi-Bank:       YES (Banks: %s)\n", strings.Join(bStrs, ", ")))
		} else {
			sb.WriteString("Multi-Bank:       NO (Single binary / Area Comum)\n")
		}

		if len(res.Segments) > 0 {
			sb.WriteString(fmt.Sprintf("\nPlaced Segments (%d):\n", len(res.Segments)))
			for i, seg := range res.Segments {
				sb.WriteString(fmt.Sprintf("  [%d] Mod %d, Seg %d: Type: %-4s Bank: %-2d Range: 0x%04X - 0x%04X (Size: %4d bytes)\n",
					i, seg.ModuleIndex, seg.SegmentIndex, seg.Type.String(), seg.Bank, seg.BaseAddr, seg.BaseAddr+seg.Size-1, seg.Size))
			}
		}

		if len(res.Symbols) > 0 {
			sb.WriteString(fmt.Sprintf("\nResolved Symbols (%d):\n", len(res.Symbols)))
			type symEntry struct {
				name string
				addr uint16
				bank uint8
				kind string
			}
			var list []symEntry
			for name, sym := range res.Symbols {
				list = append(list, symEntry{name: name, addr: sym.Address, bank: sym.Bank, kind: sym.Kind.String()})
			}
			sort.Slice(list, func(i, j int) bool {
				if list[i].addr == list[j].addr {
					return list[i].name < list[j].name
				}
				return list[i].addr < list[j].addr
			})
			for _, s := range list {
				sb.WriteString(fmt.Sprintf("  - %-24s 0x%04X (Bank %d, %s)\n", s.name, s.addr, s.bank, s.kind))
			}
		}

		if len(res.Trampolines) > 0 {
			sb.WriteString(fmt.Sprintf("\nTrampolines Generated (%d):\n", len(res.Trampolines)))
			for name, tramp := range res.Trampolines {
				sb.WriteString(fmt.Sprintf("  - %-20s Tramp: 0x%04X -> Target: 0x%04X (Bank %d)\n",
					name, tramp.Address, tramp.TargetAddress, tramp.TargetBank))
			}
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
			if (arg == "-o" || arg == "-m" || arg == "-b" || arg == "-e" || arg == "--log-file" || arg == "-log-file") && i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				i++
				flags = append(flags, args[i])
			}
		} else {
			nonFlags = append(nonFlags, arg)
		}
	}
	return append(flags, nonFlags...)
}
