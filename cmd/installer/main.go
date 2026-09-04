package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/wilsonpilon/kizuna/pkg/version"
)

const banner = `
==============================================================================
   ██╗  ██╗██╗███████╗██╗   ██╗███╗   ██╗ █████╗ 
   ██║ ██╔╝██║╚══███╔╝██║   ██║████╗  ██║██╔══██╗
   █████╔╝ ██║  ███╔╝ ██║   ██║██╔██╗ ██║███████║
   ██╔═██╗ ██║ ███╔╝  ██║   ██║██║╚██╗██║██╔══██║
   ██║  ██╗██║███████╗╚██████╔╝██║ ╚████║██║  ██║
   ╚═╝  ╚═╝╚═╝╚══════╝ ╚═════╝ ╚═╝  ╚═══╝╚═╝  ╚═╝
          絆 - MSX2+ Z80 Multi-Language Toolchain
==============================================================================`

func main() {
	reader := bufio.NewReader(os.Stdin)

	for {
		printMenu()
		fmt.Print("\nEscolha uma opção [1-6]: ")
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			installKizuna(reader)
		case "2":
			configurePath(reader)
		case "3":
			testInstallation()
		case "4":
			compileSample()
		case "5":
			showAbout()
		case "6", "q", "exit":
			fmt.Println("\nObrigado por usar o KIZUNA! Bons desenvolvimentos para MSX!")
			return
		default:
			fmt.Println("\nOpção inválida. Pressione ENTER para continuar...")
			reader.ReadString('\n')
		}
	}
}

func printMenu() {
	fmt.Println(banner)
	fmt.Printf(" Versão: %s | Release: %s\n", version.String(), version.Codename)
	fmt.Println("==============================================================================")
	fmt.Println("  [1] Instalar KIZUNA em um diretório do sistema")
	fmt.Println("  [2] Adicionar KIZUNA ao PATH do usuário no Windows")
	fmt.Println("  [3] Testar executáveis instalados (verificação de integridade)")
	fmt.Println("  [4] Compilar exemplo de teste (hello.asm -> hello.com)")
	fmt.Println("  [5] Sobre o KIZUNA, Documentação e Licença")
	fmt.Println("  [6] Sair do Instalador")
	fmt.Println("------------------------------------------------------------------------------")
}

func getSourceDir() string {
	exePath, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(exePath)
}

func installKizuna(reader *bufio.Reader) {
	srcDir := getSourceDir()
	defaultTarget := `C:\kizuna`
	if runtime.GOOS != "windows" {
		home, _ := os.UserHomeDir()
		defaultTarget = filepath.Join(home, ".kizuna")
	}

	fmt.Printf("\n--> Diretório de instalação padrão: [%s]\n", defaultTarget)
	fmt.Print("Pressione ENTER para aceitar ou digite o novo caminho: ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	targetDir := defaultTarget
	if input != "" {
		targetDir = input
	}

	fmt.Printf("\nInstalando KIZUNA em: %s...\n", targetDir)

	// Copiar pastas bin, sample, docs e LICENSE
	itemsToCopy := []string{"bin", "sample", "docs", "LICENSE"}
	for _, item := range itemsToCopy {
		srcPath := filepath.Join(srcDir, item)
		dstPath := filepath.Join(targetDir, item)
		if fi, err := os.Stat(srcPath); err == nil {
			if fi.IsDir() {
				if err := copyDir(srcPath, dstPath); err != nil {
					fmt.Printf("  [!] Erro ao copiar %s: %v\n", item, err)
				} else {
					fmt.Printf("  [OK] Diretório '%s' copiado com sucesso.\n", item)
				}
			} else {
				if err := copyFile(srcPath, dstPath); err != nil {
					fmt.Printf("  [!] Erro ao copiar %s: %v\n", item, err)
				} else {
					fmt.Printf("  [OK] Arquivo '%s' copiado com sucesso.\n", item)
				}
			}
		}
	}

	fmt.Println("\n[SUCESSO] Instalação concluída com sucesso!")
	fmt.Println("\nDica: Utilize a opção [2] para adicionar a pasta bin ao seu PATH.")
	fmt.Print("\nPressione ENTER para voltar ao menu...")
	reader.ReadString('\n')
}

func configurePath(reader *bufio.Reader) {
	if runtime.GOOS != "windows" {
		fmt.Println("\nConfiguração automática do PATH suportada no Windows.")
		fmt.Println("No Linux/macOS, adicione 'export PATH=$PATH:~/.kizuna/bin' ao seu .bashrc/.zshrc.")
		fmt.Print("\nPressione ENTER para voltar ao menu...")
		reader.ReadString('\n')
		return
	}

	defaultBin := `C:\kizuna\bin`
	fmt.Printf("\n--> Caminho da pasta bin a adicionar ao PATH: [%s]\n", defaultBin)
	fmt.Print("Pressione ENTER para aceitar ou digite outro caminho: ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	binDir := defaultBin
	if input != "" {
		binDir = input
	}

	// Executa PowerShell para adicionar ao PATH do usuário
	psScript := fmt.Sprintf(`
		$binPath = "%s"
		$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
		if ($userPath -notlike "*$binPath*") {
			$newPath = "$userPath;$binPath".Trim(';')
			[Environment]::SetEnvironmentVariable("Path", $newPath, "User")
			Write-Output "ADDED"
		} else {
			Write-Output "ALREADY_EXISTS"
		}
	`, binDir)

	cmd := exec.Command("powershell", "-NoProfile", "-Command", psScript)
	out, err := cmd.CombinedOutput()
	outStr := strings.TrimSpace(string(out))

	if err != nil {
		fmt.Printf("\n[!] Falha ao configurar PATH: %v\nSaída: %s\n", err, outStr)
	} else if strings.Contains(outStr, "ADDED") {
		fmt.Printf("\n[SUCESSO] '%s' foi adicionado com sucesso ao seu PATH de usuário!\n", binDir)
		fmt.Println("Observação: Novos terminais abertos já reconhecerão os comandos 'kaji80' e 'musubi'.")
	} else {
		fmt.Printf("\n[INFO] O caminho '%s' já estava presente no seu PATH.\n", binDir)
	}

	fmt.Print("\nPressione ENTER para voltar ao menu...")
	reader.ReadString('\n')
}

func testInstallation() {
	fmt.Println("\nVerificando executáveis do KIZUNA...")
	srcDir := getSourceDir()
	binDir := filepath.Join(srcDir, "bin")

	tools := []string{"kaji80", "musubi", "mobdump"}
	for _, tool := range tools {
		exeName := tool
		if runtime.GOOS == "windows" {
			exeName += ".exe"
		}

		toolPath := filepath.Join(binDir, exeName)
		if _, err := os.Stat(toolPath); err != nil {
			// Tenta no PATH do sistema
			toolPath = exeName
		}

		cmd := exec.Command(toolPath, "--version")
		out, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("  [FALHA] %s: não pôde ser executado (%v)\n", tool, err)
		} else {
			fmt.Printf("  [OK] %s\n", strings.TrimSpace(string(out)))
		}
	}

	fmt.Print("\nPressione ENTER para voltar ao menu...")
	bufio.NewReader(os.Stdin).ReadString('\n')
}

func compileSample() {
	srcDir := getSourceDir()
	binDir := filepath.Join(srcDir, "bin")
	sampleDir := filepath.Join(srcDir, "sample")

	kajiExe := filepath.Join(binDir, "kaji80.exe")
	musubiExe := filepath.Join(binDir, "musubi.exe")
	if runtime.GOOS != "windows" {
		kajiExe = filepath.Join(binDir, "kaji80")
		musubiExe = filepath.Join(binDir, "musubi")
	}

	helloAsm := filepath.Join(sampleDir, "hello.asm")
	helloMob := filepath.Join(sampleDir, "hello.mob")
	helloCom := filepath.Join(sampleDir, "hello.com")

	if _, err := os.Stat(helloAsm); err != nil {
		fmt.Printf("\n[!] Exemplo %s não encontrado.\n", helloAsm)
		fmt.Print("Pressione ENTER para voltar...")
		bufio.NewReader(os.Stdin).ReadString('\n')
		return
	}

	fmt.Println("\n1. Montando sample/hello.asm via KAJI80...")
	cmd1 := exec.Command(kajiExe, "-v", "-o", helloMob, helloAsm)
	out1, err1 := cmd1.CombinedOutput()
	if err1 != nil {
		fmt.Printf("  [FALHA] Montagem: %v\n%s\n", err1, string(out1))
		fmt.Print("Pressione ENTER para voltar...")
		bufio.NewReader(os.Stdin).ReadString('\n')
		return
	}
	fmt.Printf("  %s\n", strings.TrimSpace(string(out1)))

	fmt.Println("\n2. Linkando sample/hello.mob via MUSUBI...")
	cmd2 := exec.Command(musubiExe, "-v", "-o", helloCom, helloMob)
	out2, err2 := cmd2.CombinedOutput()
	if err2 != nil {
		fmt.Printf("  [FALHA] Linkagem: %v\n%s\n", err2, string(out2))
		fmt.Print("Pressione ENTER para voltar...")
		bufio.NewReader(os.Stdin).ReadString('\n')
		return
	}
	fmt.Printf("  %s\n", strings.TrimSpace(string(out2)))

	fmt.Printf("\n[SUCESSO] Executável gerado: %s\n", helloCom)
	fmt.Println("Copie este arquivo para o seu MSX ou emulador e execute: HELLO")
	fmt.Print("\nPressione ENTER para voltar ao menu...")
	bufio.NewReader(os.Stdin).ReadString('\n')
}

func showAbout() {
	fmt.Printf("\n=== KIZUNA (絆) — %s ===\n", version.FullVersion())
	fmt.Println(`
KIZUNA ("laço", "vínculo") é uma toolchain moderna para MSX2+ / MSX-DOS 2,
projetada para permitir a escrita de programas combinando Assembly Z80,
Pascal (Turbo Pascal 4 style) e MSX-BASIC, linkados em um único binário .COM.

Destaques da Release Akatsuki (Fase 4):
  * Linker Multi-Banco com suporte oficial à Memory Mapper do MSX-DOS 2.
  * Descoberta dinâmica de vetores da BIOS via EXTBIO (Device ID 4).
  * Chaveamento seguro na Página 2 (0x8000..0xBFFF) com trampolins automáticos.
  * Alinhamento automático de slots via porta 0xA8 para cartuchos externos de DOS.

Documentação completa disponível na pasta 'docs/':
  - README.md: Apresentação e arquitetura do projeto.
  - HELP.md: Referência técnica de instruções, diretivas e formato .MOB.
  - CHANGELOG.md: Histórico completo de versões e alterações.
  - LICENSE: Licença de uso.`)

	fmt.Print("\nPressione ENTER para voltar ao menu...")
	bufio.NewReader(os.Stdin).ReadString('\n')
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0755)
		}
		return copyFile(path, target)
	})
}
