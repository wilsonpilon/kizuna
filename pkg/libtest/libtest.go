// Package libtest reúne auxiliares para TESTAR rotinas da MSXLIB e o código
// gerado pelos compiladores no simulador Z80 (pkg/z80sim): monta fontes
// Assembly, monta a MSXLIB a partir de lib/src (sempre fresca, sem depender do
// lib/msxlib.hlib commitado), liga tudo com o MUSUBI e devolve o binário e a
// tabela de símbolos.
package libtest

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/wilsonpilon/kizuna/pkg/api"
	"github.com/wilsonpilon/kizuna/pkg/hako"
	"github.com/wilsonpilon/kizuna/pkg/kaji80"
	"github.com/wilsonpilon/kizuna/pkg/mob"
	"github.com/wilsonpilon/kizuna/pkg/musubi"
	"github.com/wilsonpilon/kizuna/pkg/z80sim"
)

// RepoRoot sobe a partir do diretório atual até achar o go.mod.
func RepoRoot(tb testing.TB) string {
	tb.Helper()
	dir, err := os.Getwd()
	if err != nil {
		tb.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			tb.Fatal("go.mod não encontrado subindo a partir do diretório de teste")
		}
		dir = parent
	}
}

// Assemble monta um fonte KAJI80. baseDir resolve INCLUDE/INCBIN (vazio =
// diretório atual).
func Assemble(tb testing.TB, src, baseDir string) *mob.ObjectFile {
	tb.Helper()
	a := kaji80.NewAssembler()
	if baseDir != "" {
		a.SetBaseDir(baseDir)
	}
	obj, err := a.Assemble(src)
	if err != nil {
		tb.Fatalf("falha ao montar:\n%v\n--- fonte ---\n%s", err, src)
	}
	return obj
}

var (
	libOnce sync.Once
	libArc  *hako.Archive
	libErr  error
)

// Lib monta todos os lib/src/**/*.asm e devolve a MSXLIB como arquivo .hlib
// em memória. O resultado é reaproveitado entre testes do mesmo processo.
func Lib(tb testing.TB) *hako.Archive {
	tb.Helper()
	libOnce.Do(func() {
		libArc, libErr = buildLib(RepoRoot(tb))
	})
	if libErr != nil {
		tb.Fatalf("falha ao montar a MSXLIB: %v", libErr)
	}
	return libArc
}

func buildLib(root string) (*hako.Archive, error) {
	srcDir := filepath.Join(root, "lib", "src")
	tmp, err := os.MkdirTemp("", "kizuna-libtest-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmp)

	var mobs []string
	err = filepath.WalkDir(srcDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(strings.ToLower(path), ".asm") {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		a := kaji80.NewAssembler()
		a.SetBaseDir(filepath.Dir(path))
		obj, err := a.Assemble(string(data))
		if err != nil {
			return &os.PathError{Op: "assemble", Path: path, Err: err}
		}
		rel, _ := filepath.Rel(srcDir, path)
		name := strings.TrimSuffix(strings.NewReplacer(`\`, "_", "/", "_").Replace(rel), ".asm") + ".mob"
		out := filepath.Join(tmp, name)
		if err := mob.SaveToFile(out, obj); err != nil {
			return err
		}
		mobs = append(mobs, out)
		return nil
	})
	if err != nil {
		return nil, err
	}
	hlib := filepath.Join(tmp, "msxlib.hlib")
	if err := hako.Pack(hlib, mobs...); err != nil {
		return nil, err
	}
	raw, err := os.ReadFile(hlib)
	if err != nil {
		return nil, err
	}
	return hako.FromBytes(raw)
}

// Link monta src (que deve ter o símbolo de entrada "Start"), liga contra a
// MSXLIB e devolve o resultado do linker.
func Link(tb testing.TB, src string) *musubi.LinkResult {
	tb.Helper()
	obj := Assemble(tb, src, "")
	res, err := musubi.NewLinker(musubi.DefaultConfig()).LinkWithLibraries(
		[]*mob.ObjectFile{obj}, []*hako.Archive{Lib(tb)})
	if err != nil {
		tb.Fatalf("falha ao ligar: %v\n--- fonte ---\n%s", err, src)
	}
	return res
}

// Machine devolve um simulador com o binário ligado carregado em 0100h e PC
// no início (ponto de entrada do .COM).
func Machine(res *musubi.LinkResult) *z80sim.Machine {
	m := z80sim.New()
	m.LoadCOM(res.Binary)
	m.CPU.SetPC(res.EntryPoint)
	return m
}

// Addr devolve o endereço final do símbolo, ou falha o teste.
func Addr(tb testing.TB, res *musubi.LinkResult, name string) uint16 {
	tb.Helper()
	s, ok := res.Symbols[name]
	if !ok {
		tb.Fatalf("símbolo %q não está no binário ligado", name)
	}
	return s.Address
}

// LinkObjects liga objetos já montados/compilados contra a MSXLIB.
func LinkObjects(tb testing.TB, objs ...*mob.ObjectFile) *musubi.LinkResult {
	tb.Helper()
	res, err := musubi.NewLinker(musubi.DefaultConfig()).LinkWithLibraries(objs, []*hako.Archive{Lib(tb)})
	if err != nil {
		tb.Fatalf("falha ao ligar: %v", err)
	}
	return res
}

// RepoAPI carrega os descritores reais de lib/api.
func RepoAPI(tb testing.TB) *api.Set {
	tb.Helper()
	set, err := api.LoadPaths([]string{filepath.Join(RepoRoot(tb), "lib", "api")})
	if err != nil {
		tb.Fatalf("lib/api: %v", err)
	}
	return set
}
