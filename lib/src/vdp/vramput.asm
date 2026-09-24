; =============================================================================
; KIZUNA MSXLIB - vdp/vramput
; um byte de/para a VRAM no ponteiro atual
; =============================================================================

MODULE vdp_vramput
BANK 0

INCLUDE "../../inc/ports.inc"
INCLUDE "../../inc/vdpreg.inc"

PUBLIC VDP_VramPut, VDP_VramGet

; VDP_VramPut: escreve o byte A na VRAM e avanca o ponteiro (ver VDP_VramSetWrite)
; Preserva: tudo.
VDP_VramPut:
    OUT (PORT_VDP_DATA), A
    RET

; VDP_VramGet: le um byte da VRAM e avanca o ponteiro (ver VDP_VramSetRead)
; Saída: A = byte
; Preserva: BC, DE, HL.
VDP_VramGet:
    IN A, (PORT_VDP_DATA)
    RET

ENDMOD
