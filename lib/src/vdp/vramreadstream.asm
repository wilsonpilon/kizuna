; =============================================================================
; KIZUNA MSXLIB - vdp/vramreadstream
; le um bloco da VRAM para a RAM, a partir do ponteiro atual
; =============================================================================

MODULE vdp_vramreadstream
BANK 0

INCLUDE "../../inc/ports.inc"
INCLUDE "../../inc/vdpreg.inc"

PUBLIC VDP_VramReadStream

; VDP_VramReadStream: copia BC bytes da VRAM (ponteiro definido por VDP_VramSetRead)
; para DE, na RAM
; Entrada: DE = destino na RAM, BC = quantidade (0 nao faz nada)
; Destrói: A, BC, DE, HL, flags.
VDP_VramReadStream:
    LD A, B
    OR C
    RET Z
    EX DE, HL           ; HL = destino
    LD D, B
    LD B, C
    LD C, PORT_VDP_DATA
    LD A, B
    OR A
    JR Z, VDP_VramReadStream_Blocks
    INIR
VDP_VramReadStream_Blocks:
    LD A, D
    OR A
    RET Z
VDP_VramReadStream_Loop:
    LD B, 00h
    INIR
    DEC D
    JR NZ, VDP_VramReadStream_Loop
    RET

ENDMOD
