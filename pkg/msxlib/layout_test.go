package msxlib_test

import (
	"io/fs"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wilsonpilon/kizuna/pkg/libtest"
)

// Nomes de dispositivo reservados do Windows não podem ser nome de arquivo nem de
// pasta (com qualquer extensão): "con.api" ou "lib/src/con/" simplesmente não
// abrem, e o git falha ao adicioná-los. Uma área da MSXLIB chamada "con" já passou
// dos testes no Linux-like e quebrou o commit no Windows; esta trava evita repetir.
func TestLibHasNoWindowsReservedNames(t *testing.T) {
	reserved := map[string]bool{"CON": true, "PRN": true, "AUX": true, "NUL": true}
	for i := 1; i <= 9; i++ {
		reserved["COM"+string(rune('0'+i))] = true
		reserved["LPT"+string(rune('0'+i))] = true
	}
	root := filepath.Join(libtest.RepoRoot(t), "lib")
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		base := d.Name()
		if dot := strings.Index(base, "."); dot >= 0 {
			base = base[:dot]
		}
		if reserved[strings.ToUpper(base)] {
			rel, _ := filepath.Rel(root, path)
			t.Errorf("lib/%s usa um nome reservado do Windows", filepath.ToSlash(rel))
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
