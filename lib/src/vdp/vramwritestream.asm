; =============================================================================
; KIZUNA MSXLIB - vdp/vramwritestream
; escreve um bloco da RAM na VRAM, a partir do ponteiro atual
; =============================================================================

MODULE vdp_vramwritestream
BANK 0

INCLUDE "../../inc/ports.inc"
INCLUDE "../../inc/vdpreg.inc"

PUBLIC VDP_VramWriteStream

; VDP_VramWriteStream: copia BC bytes de HL para a VRAM, no ponteiro definido por
; VDP_VramSetWrite (usa OTIR: o mais rapido que o VDP aceita)
; Entrada: HL = origem na RAM, BC = quantidade (0 nao faz nada)
; Destrói: A, BC, D, HL, flags.
VDP_VramWriteStream:
    LD A, B
    OR C
    RET Z
    LD D, B             ; D = blocos completos de 256
    LD B, C             ; B = resto
    LD C, PORT_VDP_DATA
    LD A, B
    OR A
    JR Z, VDP_VramWriteStream_Blocks
    OTIR
VDP_VramWriteStream_Blocks:
    LD A, D
    OR A
    RET Z
VDP_VramWriteStream_Loop:
    LD B, 00h           ; OTIR com B = 0 envia 256 bytes
    OTIR
    DEC D
    JR NZ, VDP_VramWriteStream_Loop
    RET

ENDMOD
