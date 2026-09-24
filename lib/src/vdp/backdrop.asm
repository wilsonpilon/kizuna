; =============================================================================
; KIZUNA MSXLIB - vdp/backdrop
; cor de fundo e cor do texto (R#7)
; =============================================================================

MODULE vdp_backdrop
BANK 0

INCLUDE "../../inc/ports.inc"
INCLUDE "../../inc/vdpreg.inc"

PUBLIC VDP_Backdrop, VDP_TextColor
EXTERN VDP_UpdateReg

; VDP_Backdrop: cor da borda/fundo (nibble baixo de R#7), 0..15
; Entrada: A = cor
; Preserva: A, BC, DE, HL. Destrói: flags. Reabilita as interrupcoes (EI).
VDP_Backdrop:
    PUSH AF
    PUSH BC
    PUSH DE
    AND 0Fh
    LD E, A
    LD C, VDP_REG_COLORS
    LD D, 0Fh
    CALL VDP_UpdateReg
    POP DE
    POP BC
    POP AF
    RET

; VDP_TextColor: cor do texto (nibble alto de R#7) nos modos de texto, 0..15
; Entrada: A = cor
; Preserva: A, BC, DE, HL. Destrói: flags. Reabilita as interrupcoes (EI).
VDP_TextColor:
    PUSH AF
    PUSH BC
    PUSH DE
    AND 0Fh
    ADD A, A
    ADD A, A
    ADD A, A
    ADD A, A
    LD E, A
    LD C, VDP_REG_COLORS
    LD D, 0F0h
    CALL VDP_UpdateReg
    POP DE
    POP BC
    POP AF
    RET

ENDMOD
