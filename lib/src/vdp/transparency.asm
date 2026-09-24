; =============================================================================
; KIZUNA MSXLIB - vdp/transparency
; cor 0 transparente (R#8, bit TP; ligado = TP em 0)
; =============================================================================

MODULE vdp_transparency
BANK 0

INCLUDE "../../inc/ports.inc"
INCLUDE "../../inc/vdpreg.inc"

PUBLIC VDP_Transparency
EXTERN VDP_UpdateReg

; VDP_Transparency: cor 0 transparente (R#8, bit TP; ligado = TP em 0)
; Entrada: A = 0 = a cor 0 e opaca, qualquer outro valor = a cor 0 e transparente
; Preserva: A, BC, DE, HL. Destrói: flags. Reabilita as interrupcoes (EI).
VDP_Transparency:
    PUSH AF
    PUSH BC
    PUSH DE
    LD E, 20h
    OR A
    JR Z, VDP_Transparency_Go
    LD E, 00h
VDP_Transparency_Go:
    LD C, VDP_REG_MODE2
    LD D, 20h
    CALL VDP_UpdateReg
    POP DE
    POP BC
    POP AF
    RET

ENDMOD
