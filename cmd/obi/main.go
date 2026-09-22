package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/wilsonpilon/kizuna/pkg/obi"
	"github.com/wilsonpilon/kizuna/pkg/version"
)

func printHelp() {
	fmt.Println(version.Banner("OBI"))
	fmt.Print(`
USO:
    obi build [Obifile]   Lê a receita declarativa e invoca KAJI80/WIRTH80/
                           DIGNAC + MUSUBI (ou HAKO, se o alvo for .hlib) na
                           ordem certa. "Obifile" no diretório atual se omitido.

OPÇÕES:
    -v             Modo detalhado (lista cada módulo/resource montado e o
                    resultado da linkagem)
    --log          Gera ou anexa ao arquivo de log da build (<Obifile>.log)
    --log-file <f> Especifica caminho customizado para o log
    --version      Exibe a versão atual
    -h, --help     Exibe esta ajuda completa

FORMATO DO Obifile:
    target: <nome>.com | <nome>.hlib   # dita a ferramenta final
    entry: Start                        # opcional (default "Start")
    base: 0x0100                        # opcional (default 0x0100)

    resources:
      - file: <caminho>
        bank: <n>
        symbol: <Nome>                  # opcional
        size: <NNN|NNNK>                # opcional, checagem de limite

    modules:
      - name: <Nome>                    # opcional
        source: <caminho>
        compiler: kaji80|wirth80|dignac # opcional, inferido da extensão
        bank: <n>                       # opcional, valida contra o BANK do fonte

    link:
      map: <caminho.map>                # opcional

    library:
      archive: <caminho.hlib>           # forma singular
    libraries:
      - <caminho.hlib>                  # forma plural

EXEMPLO:
    obi build sample/obi/Obifile
    obi build -v --log
`)
}

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "--version":
		fmt.Println(version.Banner("OBI"))
		os.Exit(0)
	case "-h", "--help", "help":
		printHelp()
		os.Exit(0)
	case "build":
		runBuild(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "Erro: comando desconhecido '%s'\n", os.Args[1])
		printHelp()
		os.Exit(1)
	}
}

func runBuild(args []string) {
	fs := flag.NewFlagSet("build", flag.ExitOnError)
	verbose := fs.Bool("v", false, "Modo detalhado")
	logFlag := fs.Bool("log", false, "Gera ou anexa ao arquivo de log da build")
	logPath := fs.String("log-file", "", "Caminho customizado para o log")
	help := fs.Bool("help", false, "Exibe ajuda detalhada")
	shortHelp := fs.Bool("h", false, "Exibe ajuda detalhada")
	fs.Usage = printHelp
	_ = fs.Parse(args)

	if *help || *shortHelp {
		printHelp()
		os.Exit(0)
	}

	obifilePath := "Obifile"
	if fs.NArg() > 0 {
		obifilePath = fs.Arg(0)
	}

	content, err := os.ReadFile(obifilePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao ler %s: %v\n", obifilePath, err)
		os.Exit(1)
	}

	cfg, err := obi.ParseObifile(string(content))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao interpretar %s:\n  %v\n", obifilePath, err)
		os.Exit(1)
	}

	baseDir := filepath.Dir(obifilePath)

	getLogDest := func() string {
		if *logPath != "" {
			return *logPath
		}
		if *logFlag {
			ext := filepath.Ext(obifilePath)
			return strings.TrimSuffix(obifilePath, ext) + ".log"
		}
		return ""
	}

	res, err := obi.Build(cfg, baseDir, obi.BuildOptions{Verbose: *verbose})
	if err != nil {
		if logDest := getLogDest(); logDest != "" {
			_ = appendBuildLog(logDest, obifilePath, cfg, res, err)
		}
		fmt.Fprintf(os.Stderr, "Erro de build: %v\n", err)
		os.Exit(1)
	}

	if logDest := getLogDest(); logDest != "" {
		if logErr := appendBuildLog(logDest, obifilePath, cfg, res, nil); logErr == nil {
			fmt.Printf("OBI: Log gravado com sucesso -> %s\n", logDest)
		}
	}

	for _, m := range res.Modules {
		kind := strings.ToUpper(m.Compiler)
		fmt.Printf("%-8s %-24s -> %-24s (banco %d, %d bytes)\n", kind, m.Source, m.MobPath, m.Bank, m.Size)
	}

	if res.IsLibrary {
		fmt.Printf("HAKO     empacotando %d módulo(s)...\n", len(res.PackedFiles))
		fmt.Printf("-> %s\n", res.Target)
	} else {
		fmt.Printf("MUSUBI   linkando %d módulo(s)", len(res.Modules))
		if len(res.Libraries) > 0 {
			fmt.Printf(" + %d biblioteca(s)", len(res.Libraries))
		}
		fmt.Println("...")
		if res.LinkResult != nil {
			if len(res.LinkResult.Trampolines) > 0 {
				for name, tramp := range res.LinkResult.Trampolines {
					fmt.Printf("         gerando trampolim: %s -> banco %d\n", name, tramp.TargetBank)
				}
			}
			fmt.Printf("-> %s  (%d bytes", res.Target, res.LinkResult.TotalSize)
			if res.LinkResult.IsMultiBank {
				fmt.Printf(", %d banco(s)", len(res.LinkResult.BanksUsed))
			}
			fmt.Println(")")
		}
	}

	if *verbose && res.LinkResult != nil {
		fmt.Printf("  Endereço Base:    0x%04X\n", res.LinkResult.BaseAddress)
		fmt.Printf("  Ponto de Entrada: 0x%04X\n", res.LinkResult.EntryPoint)
		fmt.Printf("  Total Símbolos:   %d\n", len(res.LinkResult.Symbols))
	}
}

func appendBuildLog(logPath, obifilePath string, cfg *obi.Config, res *obi.BuildResult, buildErr error) error {
	var sb strings.Builder
	sb.WriteString("--------------------------------------------------------------------------------\n")
	sb.WriteString("              KIZUNA TOOLCHAIN - OBI BUILD LOG\n")
	sb.WriteString("--------------------------------------------------------------------------------\n")
	sb.WriteString(fmt.Sprintf("Timestamp:    %s\n", time.Now().Format("2006-01-02 15:04:05")))
	sb.WriteString(fmt.Sprintf("Orchestrator: OBI v%s\n", version.FullVersion()))
	sb.WriteString(fmt.Sprintf("Obifile:      %s\n", obifilePath))
	if cfg != nil {
		sb.WriteString(fmt.Sprintf("Target:       %s\n", cfg.Target))
	}
	if buildErr != nil {
		sb.WriteString(fmt.Sprintf("Status:       FAILED (%v)\n", buildErr))
	} else {
		sb.WriteString("Status:       SUCCESS\n")
	}

	if res != nil {
		sb.WriteString(fmt.Sprintf("\nMódulos/Resources (%d):\n", len(res.Modules)))
		for _, m := range res.Modules {
			sb.WriteString(fmt.Sprintf("  - [%-8s] %-20s Banco %d, %4d bytes -> %s\n",
				m.Compiler, m.Source, m.Bank, m.Size, m.MobPath))
		}
		if len(res.Libraries) > 0 {
			sb.WriteString(fmt.Sprintf("\nBibliotecas (%d):\n", len(res.Libraries)))
			for _, lib := range res.Libraries {
				sb.WriteString(fmt.Sprintf("  - %s\n", lib))
			}
		}
		if res.LinkResult != nil {
			sb.WriteString("\nLINK DETAILS:\n")
			sb.WriteString(fmt.Sprintf("Base Address:      0x%04X\n", res.LinkResult.BaseAddress))
			sb.WriteString(fmt.Sprintf("Entry Point:       0x%04X\n", res.LinkResult.EntryPoint))
			sb.WriteString(fmt.Sprintf("Total Binary Size: %d bytes\n", res.LinkResult.TotalSize))
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
