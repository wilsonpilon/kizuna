package api

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// LoadPaths carrega arquivos .api. Cada caminho pode ser um arquivo ou um
// diretório (todos os *.api dentro dele, em ordem alfabética). Caminho
// inexistente é erro -- nunca "silenciosamente sem API".
func LoadPaths(paths []string) (*Set, error) {
	set := NewSet()
	for _, p := range paths {
		if err := loadOne(set, p); err != nil {
			return nil, err
		}
	}
	return set, nil
}

func loadOne(set *Set, path string) error {
	st, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("descritor de API '%s': %w", path, err)
	}
	if !st.IsDir() {
		return loadFile(set, path)
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}
	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.EqualFold(filepath.Ext(e.Name()), ".api") {
			files = append(files, filepath.Join(path, e.Name()))
		}
	}
	sort.Strings(files)
	for _, f := range files {
		if err := loadFile(set, f); err != nil {
			return err
		}
	}
	return nil
}

func loadFile(set *Set, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("descritor de API '%s': %w", path, err)
	}
	return set.Merge(string(data), path)
}

// ForCompiler monta o conjunto usado por um compilador de linha de comando:
// os caminhos pedidos com -api mais, se nenhum foi pedido, a descoberta
// automática ao lado do executável (<exe>/../lib/api, o layout da distribuição:
// bin/dignac.exe ao lado de lib/msxlib.hlib). Sem nenhum caminho e sem
// diretório descoberto, devolve um conjunto vazio (compila como antes).
func ForCompiler(paths []string, exePath string) (*Set, error) {
	if len(paths) == 0 && exePath != "" {
		cand := filepath.Join(filepath.Dir(exePath), "..", "lib", "api")
		if st, err := os.Stat(cand); err == nil && st.IsDir() {
			paths = []string{cand}
		}
	}
	return LoadPaths(paths)
}

// PathList implementa flag.Value para um -api repetível.
type PathList []string

func (l *PathList) String() string     { return strings.Join(*l, ",") }
func (l *PathList) Set(v string) error { *l = append(*l, v); return nil }
