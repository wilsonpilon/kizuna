package libtest

import (
	"strings"
	"testing"

	"github.com/wilsonpilon/kizuna/pkg/musubi"
	"github.com/wilsonpilon/kizuna/pkg/z80sim"
)

// Regs são os registradores de uma chamada de rotina (entrada ou saída).
type Regs struct {
	A, F, B, C, D, E, H, L byte
}

func (r Regs) HL() uint16 { return uint16(r.H)<<8 | uint16(r.L) }
func (r Regs) DE() uint16 { return uint16(r.D)<<8 | uint16(r.E) }
func (r Regs) BC() uint16 { return uint16(r.B)<<8 | uint16(r.C) }

// WithHL / WithDE / WithBC devolvem uma cópia com o par carregado.
func (r Regs) WithHL(v uint16) Regs { r.H, r.L = byte(v>>8), byte(v); return r }
func (r Regs) WithDE(v uint16) Regs { r.D, r.E = byte(v>>8), byte(v); return r }
func (r Regs) WithBC(v uint16) Regs { r.B, r.C = byte(v>>8), byte(v); return r }

// Carry informa o flag C da saída.
func (r Regs) Carry() bool { return r.F&0x01 != 0 }

// Zero informa o flag Z da saída.
func (r Regs) Zero() bool { return r.F&0x40 != 0 }

// Runner chama rotinas da MSXLIB (ligadas de verdade pelo MUSUBI) numa máquina
// Z80 simulada, com registradores dados, e devolve os registradores de saída.
// A máquina é a mesma entre chamadas, então rotinas com estado (ex.: gerador
// aleatório) o mantêm de uma chamada para a outra.
type Runner struct {
	tb   testing.TB
	res  *musubi.LinkResult
	M    *z80sim.Machine
	addr map[string]uint16
}

// NewRunner liga um programa mínimo que referencia todas as rotinas pedidas
// (mais o que elas usarem da MSXLIB) e prepara a máquina.
func NewRunner(tb testing.TB, routines ...string) *Runner {
	tb.Helper()
	var sb strings.Builder
	sb.WriteString("MODULE RunnerStub\nBANK 0\nPUBLIC Start\nEXTERN " + strings.Join(routines, ", ") + "\nStart:\n    HALT\n")
	for _, r := range routines {
		sb.WriteString("    DW " + r + "\n")
	}
	res := Link(tb, sb.String())
	run := &Runner{tb: tb, res: res, M: Machine(res), addr: map[string]uint16{}}
	for _, r := range routines {
		run.addr[r] = Addr(tb, res, r)
	}
	return run
}

// Call executa a rotina com os registradores de entrada e devolve os de saída.
// IX e IY chegam com valores conhecidos; use Runner.M diretamente para conferir
// que a rotina os preservou quando isso importar.
func (r *Runner) Call(name string, in Regs) Regs {
	r.tb.Helper()
	a, ok := r.addr[name]
	if !ok {
		r.tb.Fatalf("rotina %q não foi pedida em NewRunner", name)
	}
	c := r.M.CPU
	c.A, c.F, c.B, c.C, c.D, c.E, c.H, c.L = in.A, in.F, in.B, in.C, in.D, in.E, in.H, in.L
	if err := r.M.Call(a, 2_000_000); err != nil {
		r.tb.Fatalf("%s(%+v): %v", name, in, err)
	}
	return Regs{c.A, c.F, c.B, c.C, c.D, c.E, c.H, c.L}
}

// Addr devolve o endereço final de um símbolo no binário ligado.
func (r *Runner) Addr(name string) uint16 { return Addr(r.tb, r.res, name) }
