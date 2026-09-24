package z80sim

import "testing"

// Programa mínimo escrito à mão em bytes:
//
//	LD A,'H' ; LD E,A ; LD C,02h ; CALL 0005h     (BDOS F_CONOUT)
//	LD A,07h ; OUT (0A0h),A ; LD A,3Eh ; OUT (0A1h),A
//	LD IX,005Fh ; CALL 001Ch                      (CALSLT)
//	LD C,00h ; CALL 0005h                         (F_TERM0)
func TestRunCOM(t *testing.T) {
	prog := []byte{
		0x3E, 'H', 0x5F, 0x0E, 0x02, 0xCD, 0x05, 0x00,
		0x3E, 0x07, 0xD3, 0xA0, 0x3E, 0x3E, 0xD3, 0xA1,
		0xDD, 0x21, 0x5F, 0x00, 0xCD, 0x1C, 0x00,
		0x0E, 0x00, 0xCD, 0x05, 0x00,
		0x3E, 0x99, // não deve ser executado (F_TERM0 encerra)
	}
	m := New()
	m.LoadCOM(prog)
	if err := m.Run(1000); err != nil {
		t.Fatal(err)
	}
	if !m.Terminated() {
		t.Error("deveria ter terminado por F_TERM0")
	}
	if string(m.Console) != "H" {
		t.Errorf("console = %q", m.Console)
	}
	if len(m.Ports) != 2 || m.Ports[0] != (PortWrite{0xA0, 0x07}) || m.Ports[1] != (PortWrite{0xA1, 0x3E}) {
		t.Errorf("portas = %v", m.Ports)
	}
	if len(m.CALSLTCalls) != 1 || m.CALSLTCalls[0] != 0x005F {
		t.Errorf("CALSLT = %v", m.CALSLTCalls)
	}
	if len(m.BDOSCalls) != 2 || m.BDOSCalls[0] != 2 || m.BDOSCalls[1] != 0 {
		t.Errorf("BDOS = %v", m.BDOSCalls)
	}
	if m.CPU.A == 0x99 {
		t.Error("executou depois do F_TERM0")
	}
}

func TestCallReturnsAtSentinel(t *testing.T) {
	// 0200h: INC A ; ADD A,B ; RET
	m := New()
	m.Load(0x0200, []byte{0x3C, 0x80, 0xC9})
	m.CPU.A, m.CPU.B = 4, 10
	if err := m.Call(0x0200, 100); err != nil {
		t.Fatal(err)
	}
	if m.CPU.A != 15 {
		t.Errorf("A = %d, quer 15", m.CPU.A)
	}
	if m.CPU.SP() != 0xFF00 {
		t.Errorf("SP = %04X, quer FF00 (RET consumiu o endereço de retorno)", m.CPU.SP())
	}
}

func TestStepBudget(t *testing.T) {
	m := New()
	m.Load(0x0100, []byte{0x18, 0xFE}) // JR $ -- laço infinito
	m.CPU.SetPC(0x0100)
	if err := m.Run(500); err != ErrStepBudget {
		t.Errorf("erro = %v, quer ErrStepBudget", err)
	}
}

func TestPortInAndIXHelpers(t *testing.T) {
	m := New()
	m.PortIn = func(p uint8) byte { return p ^ 0xFF }
	m.Load(0x0100, []byte{0xDB, 0x99, 0x76}) // IN A,(99h) ; HALT
	m.CPU.SetPC(0x0100)
	if err := m.Run(10); err != nil {
		t.Fatal(err)
	}
	if m.CPU.A != 0x66 {
		t.Errorf("A = %02X, quer 66", m.CPU.A)
	}
	m.SetIX(0xBEEF)
	m.SetIY(0x1234)
	if m.IX() != 0xBEEF || m.IY() != 0x1234 {
		t.Errorf("IX/IY = %04X/%04X", m.IX(), m.IY())
	}
}
