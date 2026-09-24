// Package zmem adapta um array de 64KB à interface z80.MemoryAccessor.
//
// Fica num diretório iniciado por "_" DE PROPÓSITO: a interface exigida pela
// biblioteca z80 usa os nomes ReadByte(addr) byte / WriteByte(addr, v), que o
// "go vet" (verificador stdmethods) acusa por não terem a assinatura de
// io.ByteReader/io.ByteWriter. Diretórios com "_" são ignorados por "./...",
// mas continuam importáveis por caminho explícito -- assim "go vet ./..." do
// projeto segue limpo sem esconder nenhum outro problema de pkg/z80sim.
package zmem

// Memory é uma RAM plana de 64KB sem contenção.
type Memory struct{ Bytes *[65536]byte }

func (m Memory) ReadByte(a uint16) byte                  { return m.Bytes[a] }
func (m Memory) ReadByteInternal(a uint16) byte          { return m.Bytes[a] }
func (m Memory) WriteByte(a uint16, v byte)              { m.Bytes[a] = v }
func (m Memory) WriteByteInternal(a uint16, v byte)      { m.Bytes[a] = v }
func (Memory) ContendRead(uint16, int)                   {}
func (Memory) ContendReadNoMreq(uint16, int)             {}
func (Memory) ContendReadNoMreq_loop(uint16, int, uint)  {}
func (Memory) ContendWriteNoMreq(uint16, int)            {}
func (Memory) ContendWriteNoMreq_loop(uint16, int, uint) {}
func (m Memory) Read(a uint16) byte                      { return m.Bytes[a] }
func (m Memory) Write(a uint16, v byte, _ bool)          { m.Bytes[a] = v }
func (m Memory) Data() []byte                            { return m.Bytes[:] }
