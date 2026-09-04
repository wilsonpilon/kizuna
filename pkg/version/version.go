package version

import "fmt"

// Definição do esquema de versionamento do KIZUNA:
// MAJOR: Encerramento de cada Fase do projeto
// MINOR: Cada nova feature ou capacidade adicionada à toolchain
// BUILD: Cada compilação ou incremento de build realizado
const (
	Major    = 4
	Minor    = 4
	Build    = 0
	Codename = "Akatsuki (暁)"
)

// String retorna a versão formatada como major.minor.compilacao
func String() string {
	return fmt.Sprintf("%d.%d.%d", Major, Minor, Build)
}

// FullVersion retorna a versão completa com o codinome de release
func FullVersion() string {
	return fmt.Sprintf("%d.%d.%d [%s]", Major, Minor, Build, Codename)
}

// Banner retorna o cabeçalho oficial do Kizuna com a versão atual
func Banner(toolName string) string {
	return fmt.Sprintf("%s v%s - Toolchain KIZUNA (MSX2+ / MSX-DOS 2)", toolName, String())
}
