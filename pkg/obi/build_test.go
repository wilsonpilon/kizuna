package obi

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wilsonpilon/kizuna/pkg/z80sim"
)

func TestParseObifile_FullExample(t *testing.T) {
	content := `
# comentario
target: app.com
entry: Start
base: 0x0100

resources:
  - file: banner.txt
    bank: 0
    symbol: Res_Banner
    size: 64

modules:
  - name: Main
    source: main.asm
    compiler: kaji80
    bank: 0
  - name: Lib
    source: lib.bas
    bank: 2

link:
  map: app.map

library:
  archive: extra.hlib

libraries:
  - msxlib.hlib
`
	cfg, err := ParseObifile(content)
	if err != nil {
		t.Fatalf("ParseObifile falhou: %v", err)
	}
	if cfg.Target != "app.com" {
		t.Errorf("Target = %q, esperado app.com", cfg.Target)
	}
	if cfg.Entry != "Start" {
		t.Errorf("Entry = %q", cfg.Entry)
	}
	if cfg.Base != 0x0100 {
		t.Errorf("Base = 0x%04X, esperado 0x0100", cfg.Base)
	}
	if len(cfg.Resources) != 1 {
		t.Fatalf("esperava 1 resource, obteve %d", len(cfg.Resources))
	}
	r := cfg.Resources[0]
	if r.File != "banner.txt" || r.Bank != 0 || r.Symbol != "Res_Banner" || !r.HasSize || r.Size != 64 {
		t.Errorf("resource inesperado: %+v", r)
	}
	if len(cfg.Modules) != 2 {
		t.Fatalf("esperava 2 modules, obteve %d", len(cfg.Modules))
	}
	if cfg.Modules[0].Compiler != "kaji80" || !cfg.Modules[0].HasBank || cfg.Modules[0].Bank != 0 {
		t.Errorf("modulo 0 inesperado: %+v", cfg.Modules[0])
	}
	if cfg.Modules[1].Compiler != "" || !cfg.Modules[1].HasBank || cfg.Modules[1].Bank != 2 {
		t.Errorf("modulo 1 inesperado: %+v", cfg.Modules[1])
	}
	if cfg.Link.Map != "app.map" {
		t.Errorf("Link.Map = %q", cfg.Link.Map)
	}
	wantLibs := []string{"extra.hlib", "msxlib.hlib"}
	if len(cfg.Libraries) != 2 || cfg.Libraries[0] != wantLibs[0] || cfg.Libraries[1] != wantLibs[1] {
		t.Errorf("Libraries = %v, esperado %v", cfg.Libraries, wantLibs)
	}
}

func TestParseObifile_MissingTarget(t *testing.T) {
	_, err := ParseObifile("modules:\n  - source: a.asm\n")
	if err == nil {
		t.Fatal("esperava erro por 'target' ausente")
	}
}

func TestParseObifile_UnknownKey(t *testing.T) {
	_, err := ParseObifile("target: a.com\nfoo: bar\n")
	if err == nil {
		t.Fatal("esperava erro por chave desconhecida")
	}
}

func TestParseSize(t *testing.T) {
	cases := map[string]int{"12K": 12 * 1024, "2048": 2048, "1k": 1024}
	for in, want := range cases {
		got, err := parseSize(in)
		if err != nil {
			t.Fatalf("parseSize(%q) erro: %v", in, err)
		}
		if got != want {
			t.Errorf("parseSize(%q) = %d, esperado %d", in, got, want)
		}
	}
}

func TestParseAddress(t *testing.T) {
	cases := map[string]uint16{"0x0100": 0x0100, "0100h": 0x0100, "256": 256}
	for in, want := range cases {
		got, err := parseAddress(in)
		if err != nil {
			t.Fatalf("parseAddress(%q) erro: %v", in, err)
		}
		if got != want {
			t.Errorf("parseAddress(%q) = 0x%04X, esperado 0x%04X", in, got, want)
		}
	}
}

func TestBuild_EndToEnd(t *testing.T) {
	dir := t.TempDir()

	mainAsm := `MODULE MAINT
BANK 0

PUBLIC Start
EXTERN Helper
EXTERN Res_Banner

Start:
    LD HL, Res_Banner
    CALL Helper
    RET

ENDMOD
`
	helperAsm := `MODULE HELPERT
BANK 0

PUBLIC Helper

Helper:
    RET

ENDMOD
`
	writeFile(t, filepath.Join(dir, "main.asm"), mainAsm)
	writeFile(t, filepath.Join(dir, "helper.asm"), helperAsm)
	writeFile(t, filepath.Join(dir, "banner.txt"), "Hello$")

	obifile := `target: app.com
entry: Start

resources:
  - file: banner.txt
    bank: 0
    symbol: Res_Banner

modules:
  - name: Main
    source: main.asm
  - name: Helper
    source: helper.asm

link:
  map: app.map
`
	cfg, err := ParseObifile(obifile)
	if err != nil {
		t.Fatalf("ParseObifile: %v", err)
	}

	res, err := Build(cfg, dir, BuildOptions{})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	if res.IsLibrary {
		t.Fatal("target .com não deveria ser tratado como biblioteca")
	}
	if res.LinkResult == nil {
		t.Fatal("LinkResult nulo")
	}
	if res.LinkResult.TotalSize == 0 {
		t.Fatal("binário vazio")
	}
	if _, ok := res.LinkResult.Symbols["Res_Banner"]; !ok {
		t.Error("símbolo Res_Banner não foi resolvido pela linkagem")
	}
	if _, ok := res.LinkResult.Symbols["Helper"]; !ok {
		t.Error("símbolo Helper não foi resolvido pela linkagem")
	}

	comPath := filepath.Join(dir, "app.com")
	if _, err := os.Stat(comPath); err != nil {
		t.Errorf("app.com não foi gerado: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "app.map")); err != nil {
		t.Errorf("app.map não foi gerado: %v", err)
	}
	for _, mr := range res.Modules {
		if _, err := os.Stat(mr.MobPath); err != nil {
			t.Errorf(".mob intermediário ausente para %s: %v", mr.Name, err)
		}
	}
}

func TestBuild_LibraryTarget(t *testing.T) {
	dir := t.TempDir()
	src := `MODULE LIBT
BANK 0

PUBLIC Foo

Foo:
    RET

ENDMOD
`
	writeFile(t, filepath.Join(dir, "foo.asm"), src)

	obifile := `target: foo.hlib
modules:
  - source: foo.asm
`
	cfg, err := ParseObifile(obifile)
	if err != nil {
		t.Fatalf("ParseObifile: %v", err)
	}
	res, err := Build(cfg, dir, BuildOptions{})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if !res.IsLibrary {
		t.Error("target .hlib deveria marcar IsLibrary")
	}
	if _, err := os.Stat(filepath.Join(dir, "foo.hlib")); err != nil {
		t.Errorf("foo.hlib não foi gerado: %v", err)
	}
}

func TestBuild_ResourceSizeExceeded(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "big.bin"), "0123456789")
	cfg := &Config{
		Target: "x.hlib",
		Resources: []ResourceSpec{
			{File: "big.bin", Bank: 0, Symbol: "Res_Big", Size: 4, HasSize: true},
		},
	}
	if _, err := Build(cfg, dir, BuildOptions{}); err == nil {
		t.Fatal("esperava erro por tamanho excedido")
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("falha ao escrever %s: %v", path, err)
	}
}

// A chave "api:" do Obifile carrega descritores .api para os modulos BASIC e
// Pascal. O programa chama Diff(50, 8) -- rotina de REGISTRADORES (HL - DE) --
// e imprime o resultado; rodado no simulador Z80, so com o descritor o
// resultado e 42 (sem ele, a chamada sai em convencao de pilha e a rotina le
// registradores errados).
func TestBuild_APIKey(t *testing.T) {
	dir := t.TempDir()
	write := func(name, content string) {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	nl := "\n"
	write("my.api", "func Diff(a: word in HL, b: word in DE): word out HL"+nl+"proc PrintDec16(v: word in HL)"+nl)
	write("diff.asm", "MODULE Diff"+nl+"BANK 0"+nl+"PUBLIC Diff"+nl+"Diff:"+nl+"    OR A"+nl+"    SBC HL, DE"+nl+"    RET"+nl+"ENDMOD"+nl)
	write("main.bas", "MODULE M"+nl+"BANK 0"+nl+"PUBLIC Main"+nl+"PROCEDURE Main()"+nl+"    PrintDec16(Diff(50, 8))"+nl+"END PROCEDURE"+nl+"END MODULE"+nl)
	lib, err := filepath.Abs(filepath.Join("..", "..", "lib", "msxlib.hlib"))
	if err != nil {
		t.Fatal(err)
	}
	write("Obifile", "target: app.com"+nl+"entry: Start"+nl+"modules:"+nl+"  - source: main.bas"+nl+"  - source: diff.asm"+nl+"api:"+nl+"  - my.api"+nl+"libraries:"+nl+"  - "+filepath.ToSlash(lib)+nl)

	content, _ := os.ReadFile(filepath.Join(dir, "Obifile"))
	cfg, err := ParseObifile(string(content))
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.API) != 1 || cfg.API[0] != "my.api" {
		t.Fatalf("cfg.API = %v", cfg.API)
	}

	run := func() string {
		res, err := Build(cfg, dir, BuildOptions{})
		if err != nil {
			t.Fatalf("build: %v", err)
		}
		m := z80sim.New()
		m.LoadCOM(res.LinkResult.Binary)
		m.CPU.SetPC(res.LinkResult.EntryPoint)
		if err := m.Run(1_000_000); err != nil {
			t.Fatal(err)
		}
		return string(m.Console)
	}

	if got := run(); got != "42" {
		t.Errorf("com api: console = %q, quer %q", got, "42")
	}

	// sem descritor, chamar rotina de registradores como se fosse procedure e
	// um ERRO de compilacao (antes ligava e calculava errado em silencio)
	cfg.API = nil
	if _, err := Build(cfg, dir, BuildOptions{}); err == nil || !strings.Contains(err.Error(), "nao e PROCEDURE") && !strings.Contains(err.Error(), "não é PROCEDURE") {
		t.Errorf("sem api: esperava erro claro de chamada nao declarada, veio: %v", err)
	}

	// caminho de api inexistente e erro alto, nunca silencioso
	cfg.API = []string{"nao_existe.api"}
	if _, err := Build(cfg, dir, BuildOptions{}); err == nil {
		t.Error("api: inexistente deveria falhar")
	}
}
