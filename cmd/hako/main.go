package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/wilsonpilon/kizuna/pkg/hako"
	"github.com/wilsonpilon/kizuna/pkg/version"
)

func printHelp() {
	fmt.Println(version.Banner("HAKO"))
	fmt.Print(`
USO:
    hako -c <arquivo.hlib> <obj1.mob> [obj2.mob ...]   (Cria/Empacota biblioteca)
    hako -t <arquivo.hlib>                             (Lista módulos e símbolos)
    hako -x <arquivo.hlib> [modulo]                    (Extrai módulo ou todos)

OPÇÕES:
    -c <saida.hlib>    Cria um novo arquivo de biblioteca a partir de objetos .mob
    -t <arquivo.hlib>  Lista o conteúdo e dicionário de símbolos da biblioteca
    -x <arquivo.hlib>  Extrai módulos contidos na biblioteca
    --version          Exibe a versão atual
    -h, --help         Exibe esta ajuda completa

EXEMPLOS:
    # Criar uma biblioteca de runtime agrupando módulos matemáticos e I/O
    hako -c runtime.hlib math.mob string.mob bdos.mob

    # Listar os módulos e símbolos contidos na biblioteca
    hako -t runtime.hlib

    # Extrair todos os objetos contidos na biblioteca
    hako -x runtime.hlib
`)
}

func main() {
	help := flag.Bool("help", false, "Exibe ajuda detalhada")
	shortHelp := flag.Bool("h", false, "Exibe ajuda detalhada")
	verFlag := flag.Bool("version", false, "Exibe versão do HAKO")

	createFlag := flag.String("c", "", "Cria uma nova biblioteca .HLIB")
	listFlag := flag.String("t", "", "Lista o conteúdo da biblioteca .HLIB")
	extractFlag := flag.String("x", "", "Extrai módulos da biblioteca .HLIB")

	flag.Usage = printHelp
	flag.Parse()

	if *verFlag {
		fmt.Println(version.Banner("HAKO"))
		os.Exit(0)
	}

	if *help || *shortHelp {
		printHelp()
		os.Exit(0)
	}

	// 1. Criar biblioteca (-c)
	if *createFlag != "" {
		mobFiles := flag.Args()
		if len(mobFiles) == 0 {
			fmt.Fprintf(os.Stderr, "Erro: especifique ao menos um arquivo .mob para empacotar em %s\n", *createFlag)
			os.Exit(1)
		}

		outLib := *createFlag
		if !strings.HasSuffix(strings.ToLower(outLib), ".hlib") {
			outLib += ".hlib"
		}

		err := hako.Pack(outLib, mobFiles...)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Erro ao criar biblioteca: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("HAKO: biblioteca '%s' criada com sucesso contendo %d modulo(s).\n", outLib, len(mobFiles))
		return
	}

	// 2. Listar biblioteca (-t)
	if *listFlag != "" {
		arc, err := hako.Open(*listFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Erro ao abrir biblioteca: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("=== KIZUNA HAKO ARCHIVE: %s ===\n", *listFlag)
		fmt.Printf("Formato Versão: %d\n", arc.Version)
		fmt.Printf("Total de Módulos:  %d\n", len(arc.Modules))
		fmt.Printf("Total de Símbolos: %d\n\n", len(arc.Symbols))

		fmt.Println("--- MÓDULOS ARQUIVADOS ---")
		for i, m := range arc.Modules {
			fmt.Printf("[%2d] %-20s | Tam: %5d bytes | Offset: 0x%06X | Símbolos: %2d\n",
				i, m.Name, m.Size, m.Offset, m.SymbolCount)
		}
		fmt.Println()

		fmt.Println("--- DICIONÁRIO GLOBAL DE SÍMBOLOS (PUBLIC) ---")
		for _, s := range arc.Symbols {
			fmt.Printf("  %-24s -> Módulo: '%s' [%s]\n", s.Name, s.ModuleName, s.Kind)
		}
		return
	}

	// 3. Extrair biblioteca (-x)
	if *extractFlag != "" {
		arc, err := hako.Open(*extractFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Erro ao abrir biblioteca: %v\n", err)
			os.Exit(1)
		}

		targetModule := ""
		if len(flag.Args()) > 0 {
			targetModule = flag.Args()[0]
		}

		extracted := 0
		for _, m := range arc.Modules {
			if targetModule != "" && m.Name != targetModule {
				continue
			}

			raw, err := arc.ExtractRaw(m.Name)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Erro ao extrair '%s': %v\n", m.Name, err)
				os.Exit(1)
			}

			outName := m.Name + ".mob"
			if err := os.WriteFile(outName, raw, 0644); err != nil {
				fmt.Fprintf(os.Stderr, "Erro ao gravar '%s': %v\n", outName, err)
				os.Exit(1)
			}
			fmt.Printf("  [OK] Módulo extraído: %s (%d bytes)\n", outName, len(raw))
			extracted++
		}

		if extracted == 0 && targetModule != "" {
			fmt.Fprintf(os.Stderr, "Módulo '%s' não encontrado na biblioteca %s\n", targetModule, *extractFlag)
			os.Exit(1)
		}
		fmt.Printf("HAKO: %d modulo(s) extraido(s) com sucesso.\n", extracted)
		return
	}

	// Se nenhum comando foi fornecido, verifica argumentos posicionais legados
	args := flag.Args()
	if len(args) >= 2 && args[0] == "pack" {
		outLib := args[1]
		mobFiles := args[2:]
		if err := hako.Pack(outLib, mobFiles...); err != nil {
			fmt.Fprintf(os.Stderr, "Erro: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("HAKO: biblioteca '%s' criada com sucesso.\n", outLib)
		return
	}

	printHelp()
	os.Exit(1)
}
