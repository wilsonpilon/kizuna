package api_test

import (
	"testing"

	"github.com/wilsonpilon/kizuna/pkg/libtest"
)

// Todo descritor de lib/api tem que apontar para uma rotina que existe de fato
// na MSXLIB (símbolo PUBLIC de algum módulo) -- senão o programa do usuário
// só descobriria na hora de ligar.
func TestRepoDescriptorsMatchLibrary(t *testing.T) {
	set := libtest.RepoAPI(t)
	if set.Len() < 30 {
		t.Fatalf("lib/api descreve só %d rotinas; esperava dezenas", set.Len())
	}
	lib := libtest.Lib(t)
	for _, name := range set.Names() {
		if _, ok := lib.FindModuleForSymbol(name); !ok {
			t.Errorf("lib/api descreve %q, que não existe na MSXLIB", name)
		}
	}
}
