; =============================================================================
; KIZUNA MSXLIB - vdp/pagealternate
; alterna as paginas par/impar a cada quadro (R#9, bit EO)
; =============================================================================

MODULE vdp_pagealternate
BANK 0

INCLUDE "../../inc/ports.inc"
INCLUDE "../../inc/vdpreg.inc"

PUBLIC VDP_PageAlternate
EXTERN VDP_UpdateReg

; VDP_PageAlternate: alterna as paginas par/impar a cada quadro (R#9, bit EO)
; Entrada: A = 0 = desliga, qualquer outro valor = liga
; Preserva: A, BC, DE, HL. Destrói: flags. Reabilita as interrupcoes (EI).
VDP_PageAlternate:
    PUSH AF
    PUSH BC
    PUSH DE
    LD E, 00h
    OR A
    JR Z, VDP_PageAlternate_Go
    LD E, 04h
VDP_PageAlternate_Go:
    LD C, VDP_REG_MODE3
    LD D, 04h
    CALL VDP_UpdateReg
    POP DE
    POP BC
    POP AF
    RET

ENDMOD
