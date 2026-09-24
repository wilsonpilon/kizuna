; =============================================================================
; KIZUNA MSXLIB - vdp/interlace
; entrelacado (R#9, bit IL)
; =============================================================================

MODULE vdp_interlace
BANK 0

INCLUDE "../../inc/ports.inc"
INCLUDE "../../inc/vdpreg.inc"

PUBLIC VDP_Interlace
EXTERN VDP_UpdateReg

; VDP_Interlace: entrelacado (R#9, bit IL)
; Entrada: A = 0 = desliga, qualquer outro valor = liga
; Preserva: A, BC, DE, HL. Destrói: flags. Reabilita as interrupcoes (EI).
VDP_Interlace:
    PUSH AF
    PUSH BC
    PUSH DE
    LD E, 00h
    OR A
    JR Z, VDP_Interlace_Go
    LD E, 08h
VDP_Interlace_Go:
    LD C, VDP_REG_MODE3
    LD D, 08h
    CALL VDP_UpdateReg
    POP DE
    POP BC
    POP AF
    RET

ENDMOD
