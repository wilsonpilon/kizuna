; =============================================================================
; KIZUNA MSXLIB - vdp/vramfillstream
; preenche a VRAM com um byte, a partir do ponteiro atual
; =============================================================================

MODULE vdp_vramfillstream
BANK 0

INCLUDE "../../inc/ports.inc"
INCLUDE "../../inc/vdpreg.inc"

PUBLIC VDP_VramFillStream

; VDP_VramFillStream: escreve BC vezes o byte A na VRAM, no ponteiro definido por
; VDP_VramSetWrite
; Entrada: A = valor, BC = quantidade (0 nao faz nada)
; Destrói: A, BC, E, flags.
VDP_VramFillStream:
    LD E, A
    LD A, B
    OR C
    RET Z
VDP_VramFillStream_Loop:
    LD A, E
    OUT (PORT_VDP_DATA), A
    DEC BC
    LD A, B
    OR C
    JR NZ, VDP_VramFillStream_Loop
    RET

ENDMOD
