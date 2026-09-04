; =============================================================================
; KIZUNA MSXLIB - VDP.ASM
; Rotinas de controle do processador de vídeo TMS9918 / V9938 / V9958
; =============================================================================

MODULE VDP
BANK 0

PUBLIC VDP_WriteReg, VDP_SetWriteAddr, VDP_SetReadAddr
PUBLIC VDP_FillVRAM, VDP_WriteVRAM, VDP_ReadVRAM, VDP_SetColor

VDP_DATA EQU 0098h
VDP_CMD  EQU 0099h

; -----------------------------------------------------------------------------
; VDP_WriteReg: Escreve valor em um registrador do VDP
; Entrada: C = número do registrador (0..23), B = valor a escrever
; -----------------------------------------------------------------------------
VDP_WriteReg:
    LD A, B
    OUT (VDP_CMD), A
    LD A, C
    OR 80h
    OUT (VDP_CMD), A
    RET

; -----------------------------------------------------------------------------
; VDP_SetWriteAddr: Configura ponteiro da VRAM para escrita sequencial
; Entrada: HL = endereço de 14 bits da VRAM (0x0000..0x3FFF)
; -----------------------------------------------------------------------------
VDP_SetWriteAddr:
    LD A, L
    OUT (VDP_CMD), A
    LD A, H
    AND 3Fh
    OR 40h
    OUT (VDP_CMD), A
    RET

; -----------------------------------------------------------------------------
; VDP_SetReadAddr: Configura ponteiro da VRAM para leitura sequencial
; Entrada: HL = endereço de 14 bits da VRAM (0x0000..0x3FFF)
; -----------------------------------------------------------------------------
VDP_SetReadAddr:
    LD A, L
    OUT (VDP_CMD), A
    LD A, H
    AND 3Fh
    OUT (VDP_CMD), A
    RET

; -----------------------------------------------------------------------------
; VDP_FillVRAM: Preenche área da VRAM com um byte repetido
; Entrada: HL = endereço inicial VRAM, BC = quantidade de bytes, A = valor
; -----------------------------------------------------------------------------
VDP_FillVRAM:
    PUSH AF
    CALL VDP_SetWriteAddr
    POP AF
VDP_Fill_Loop:
    OUT (VDP_DATA), A
    DEC BC
    LD D, A
    LD A, B
    OR C
    LD A, D
    JR NZ, VDP_Fill_Loop
    RET

; -----------------------------------------------------------------------------
; VDP_WriteVRAM: Copia bloco de dados da RAM para a VRAM
; Entrada: HL = origem na RAM, DE = destino na VRAM, BC = tamanho em bytes
; -----------------------------------------------------------------------------
VDP_WriteVRAM:
    EX DE, HL
    CALL VDP_SetWriteAddr
    EX DE, HL
VDP_Write_Loop:
    LD A, (HL)
    OUT (VDP_DATA), A
    INC HL
    DEC BC
    LD A, B
    OR C
    JR NZ, VDP_Write_Loop
    RET

; -----------------------------------------------------------------------------
; VDP_ReadVRAM: Copia bloco de dados da VRAM para a RAM
; Entrada: HL = origem na VRAM, DE = destino na RAM, BC = tamanho em bytes
; -----------------------------------------------------------------------------
VDP_ReadVRAM:
    CALL VDP_SetReadAddr
VDP_Read_Loop:
    IN A, (VDP_DATA)
    LD (DE), A
    INC DE
    DEC BC
    LD A, B
    OR C
    JR NZ, VDP_Read_Loop
    RET

; -----------------------------------------------------------------------------
; VDP_SetColor: Define as cores de texto (frente) e fundo (Reg 7)
; Entrada: H = cor de frente (0..15), L = cor de fundo (0..15)
; -----------------------------------------------------------------------------
VDP_SetColor:
    LD A, H
    SLA A
    SLA A
    SLA A
    SLA A
    AND F0h
    LD B, A
    LD A, L
    AND 0Fh
    OR B
    LD B, A
    LD C, 07h
    CALL VDP_WriteReg
    RET

ENDMOD
