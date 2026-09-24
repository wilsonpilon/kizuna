; =============================================================================
; KIZUNA MSXLIB - vdp/hblankline
; linha em que ocorre a interrupcao horizontal (R#19)
; =============================================================================

MODULE vdp_hblankline
BANK 0

INCLUDE "../../inc/ports.inc"
INCLUDE "../../inc/vdpreg.inc"

PUBLIC VDP_SetHBlankLine
EXTERN VDP_SetReg

; VDP_SetHBlankLine: linha em que ocorre a interrupcao horizontal (R#19)
; Entrada: A = numero da linha
; Preserva: A, BC, DE, HL. Destrói: flags. Reabilita as interrupcoes (EI).
VDP_SetHBlankLine:
    PUSH AF
    PUSH BC
    LD B, A
    LD C, VDP_REG_HINTLINE
    CALL VDP_SetReg
    POP BC
    POP AF
    RET

ENDMOD
