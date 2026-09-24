; =============================================================================
; KIZUNA MSXLIB - vdp/grayscale
; saida em preto e branco (R#8, bit BW)
; =============================================================================

MODULE vdp_grayscale
BANK 0

INCLUDE "../../inc/ports.inc"
INCLUDE "../../inc/vdpreg.inc"

PUBLIC VDP_GrayScale
EXTERN VDP_UpdateReg

; VDP_GrayScale: saida em preto e branco (R#8, bit BW)
; Entrada: A = 0 = cores, qualquer outro valor = tons de cinza
; Preserva: A, BC, DE, HL. Destrói: flags. Reabilita as interrupcoes (EI).
VDP_GrayScale:
    PUSH AF
    PUSH BC
    PUSH DE
    LD E, 00h
    OR A
    JR Z, VDP_GrayScale_Go
    LD E, 01h
VDP_GrayScale_Go:
    LD C, VDP_REG_MODE2
    LD D, 01h
    CALL VDP_UpdateReg
    POP DE
    POP BC
    POP AF
    RET

ENDMOD
