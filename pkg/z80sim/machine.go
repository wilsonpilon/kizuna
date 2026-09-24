// Package z80sim é um simulador Z80 mínimo, usado nos TESTES da toolchain e da
// MSXLIB: roda um .COM inteiro (com BDOS/CALSLT simulados) ou chama uma única
// rotina com registradores dados, registrando o console e todas as escritas
// nas portas de I/O.
//
// Não é um emulador de MSX: não há slots, VDP, PSG nem ROM. A verificação de
// rotinas de hardware é feita pelo TRAÇO de escritas de porta (Ports), que é
// comparado com o que o manual do chip manda escrever.
//
// Surgiu como ferramenta de depuração de scratchpad quando a MSXLIB provou que
// só o emulador pega certas classes de erro do montador (ex.: "LD B,(nn)"
// virando "LD B,0" em silêncio) -- ver CHANGELOG.
package z80sim

import (
	"fmt"

	"github.com/remogatto/z80"

	zmem "github.com/wilsonpilon/kizuna/pkg/z80sim/_zmem"
)

// PortWrite é uma escrita OUT registrada.
type PortWrite struct {
	Port  uint8
	Value uint8
}

// Machine é um Z80 com 64KB de RAM plana.
type Machine struct {
	CPU *z80.Z80
	Mem [65536]byte

	// Console acumula o que o programa imprimiu via BDOS (funções 02h e 09h).
	Console []byte
	// Ports acumula todas as escritas OUT, em ordem.
	Ports []PortWrite
	// PortIn responde a leituras IN (padrão: 0xFF, barramento flutuante).
	PortIn func(port uint8) byte
	// CALSLTCalls registra o IX (endereço da rotina da BIOS) de cada CALL 001Ch
	// (chamada inter-slot). O simulador SEMPRE intercepta 001Ch: a ROM não
	// existe aqui, então a chamada vira um RET (a troca de tela real só dá
	// para confirmar em hardware/openMSX).
	CALSLTCalls []uint16
	// OnCALSLT, se não-nil, roda em cada CALL 001Ch antes do RET simulado --
	// para o teste devolver valores em registradores, por exemplo.
	OnCALSLT func(m *Machine)
	// Input é a fila de teclas que o BDOS entrega: funções 01h/08h (uma tecla,
	// 0 se a fila está vazia), 0Bh (A = FFh se há tecla) e 0Ah (leitura de linha
	// até o ENTER). Sem eco no Console além do da função 01h.
	Input []byte
	// BDOSCalls registra o número de função (registrador C) de cada CALL 0005h.
	BDOSCalls []uint8

	terminated bool
}

type ports struct{ m *Machine }

func (p ports) ReadPort(a uint16) byte {
	if p.m.PortIn != nil {
		return p.m.PortIn(uint8(a))
	}
	return 0xFF
}
func (p ports) ReadPortInternal(a uint16, _ bool) byte { return p.ReadPort(a) }
func (p ports) WritePort(a uint16, b byte) {
	p.m.Ports = append(p.m.Ports, PortWrite{Port: uint8(a), Value: b})
}
func (p ports) WritePortInternal(a uint16, b byte, _ bool) { p.WritePort(a, b) }
func (ports) ContendPortPreio(uint16)                      {}
func (ports) ContendPortPostio(uint16)                     {}

// New cria uma máquina zerada (SP=0xFF00, PC=0).
func New() *Machine {
	m := &Machine{}
	m.CPU = z80.NewZ80(zmem.Memory{Bytes: &m.Mem}, ports{m})
	m.CPU.SetSP(0xFF00)
	return m
}

// Load copia bytes para a memória a partir de addr.
func (m *Machine) Load(addr uint16, data []byte) {
	copy(m.Mem[addr:], data)
}

// LoadCOM carrega uma imagem .COM em 0100h e aponta PC para lá.
func (m *Machine) LoadCOM(img []byte) {
	m.Load(0x0100, img)
	m.CPU.SetPC(0x0100)
	m.CPU.SetSP(0xFF00)
}

// Terminated informa se a execução terminou por BDOS F_TERM0 (função 00h)
// ou por PC chegar em 0000h (warm boot).
func (m *Machine) Terminated() bool { return m.terminated }

// ErrStepBudget é devolvido por Run/Call quando o orçamento de passos acaba.
var ErrStepBudget = fmt.Errorf("orçamento de passos excedido (possível laço infinito)")

// Run executa até o programa terminar (F_TERM0 ou PC=0000h) ou estourar
// maxSteps instruções. Uma execução que só estoura o orçamento devolve
// ErrStepBudget mas mantém console e portas já registrados -- útil para
// programas que terminam num laço "espera tecla".
func (m *Machine) Run(maxSteps int) error {
	return m.run(maxSteps, nil)
}

// sentinel é o endereço de retorno usado por Call. Está em RAM alta,
// longe de qualquer código de teste.
const sentinel = 0xFFF0

// Call executa a rotina em addr como se fosse um CALL: empilha um endereço de
// retorno e roda até o RET correspondente. Registradores devem ser preparados
// antes por m.CPU. O SP é restaurado para 0xFF00 antes de começar.
func (m *Machine) Call(addr uint16, maxSteps int) error {
	sp := uint16(0xFF00 - 2)
	m.CPU.SetSP(sp)
	m.Mem[sp] = byte(sentinel & 0xFF)
	m.Mem[sp+1] = byte(sentinel >> 8)
	m.CPU.SetPC(addr)
	stop := uint16(sentinel)
	return m.run(maxSteps, &stop)
}

func (m *Machine) run(maxSteps int, stopAt *uint16) error {
	cpu := m.CPU
	for i := 0; i < maxSteps; i++ {
		pc := cpu.PC()
		if stopAt != nil && pc == *stopAt {
			return nil
		}
		if stopAt == nil && pc == 0x0000 {
			m.terminated = true
			return nil
		}
		if cpu.Halted {
			return nil
		}
		switch pc {
		case 0x0005:
			m.BDOSCalls = append(m.BDOSCalls, cpu.C)
			switch cpu.C {
			case 0x00:
				m.terminated = true
				return nil
			case 0x01, 0x08:
				cpu.A = m.nextKey()
				if cpu.C == 0x01 && cpu.A != 0 {
					m.Console = append(m.Console, cpu.A)
				}
			case 0x0B:
				cpu.A = 0
				if len(m.Input) > 0 {
					cpu.A = 0xFF
				}
			case 0x0A:
				m.readLine(cpu.DE())
			case 0x02:
				m.Console = append(m.Console, cpu.E)
			case 0x09:
				for a := cpu.DE(); m.Mem[a] != '$'; a++ {
					m.Console = append(m.Console, m.Mem[a])
				}
			}
			m.ret()
			continue
		case 0x001C:
			m.CALSLTCalls = append(m.CALSLTCalls, m.IX())
			if m.OnCALSLT != nil {
				m.OnCALSLT(m)
			}
			m.ret()
			continue
		}
		cpu.DoOpcode()
	}
	return ErrStepBudget
}

func (m *Machine) nextKey() byte {
	if len(m.Input) == 0 {
		return 0
	}
	k := m.Input[0]
	m.Input = m.Input[1:]
	return k
}

// readLine implementa a função 0Ah: buffer = [máximo][tamanho][dados...].
// Consome a fila até um CR ou LF (e o LF que segue um CR).
func (m *Machine) readLine(buf uint16) {
	max := int(m.Mem[buf])
	n := 0
	for len(m.Input) > 0 {
		c := m.Input[0]
		m.Input = m.Input[1:]
		if c == '\r' || c == '\n' {
			if c == '\r' && len(m.Input) > 0 && m.Input[0] == '\n' {
				m.Input = m.Input[1:]
			}
			break
		}
		if n < max {
			m.Mem[int(buf)+2+n] = c
			n++
		}
	}
	m.Mem[buf+1] = byte(n)
}

func (m *Machine) ret() {
	sp := m.CPU.SP()
	m.CPU.SetPC(uint16(m.Mem[sp+1])<<8 | uint16(m.Mem[sp]))
	m.CPU.SetSP(sp + 2)
}

// IX / IY como valores de 16 bits.
func (m *Machine) IX() uint16 { return uint16(m.CPU.IXH)<<8 | uint16(m.CPU.IXL) }
func (m *Machine) IY() uint16 { return uint16(m.CPU.IYH)<<8 | uint16(m.CPU.IYL) }

// SetIX / SetIY.
func (m *Machine) SetIX(v uint16) { m.CPU.IXH, m.CPU.IXL = byte(v>>8), byte(v) }
func (m *Machine) SetIY(v uint16) { m.CPU.IYH, m.CPU.IYL = byte(v>>8), byte(v) }
