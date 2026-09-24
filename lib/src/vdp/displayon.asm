; =============================================================================
; KIZUNA MSXLIB - vdp/displayon
; liga a exibicao da tela (R#1, bit BL)
; =============================================================================

MODULE vdp_displayon
BANK 0

INCLUDE "../../inc/ports.inc"
INCLUDE "../../inc/vdpreg.inc"

PUBLIC VDP_DisplayOn
EXTERN VDP_UpdateReg

; VDP_DisplayOn: liga a exibicao da tela (R#1, bit BL)
; Preserva: A, BC, DE, HL. Destrói: flags. Reabilita as interrupcoes (EI).
VDP_DisplayOn:
    PUSH BC
    PUSH DE
    LD C, VDP_REG_MODE1
    LD D, VDP_R1_BL
    LD E, VDP_R1_BL
    CALL VDP_UpdateReg
    POP DE
    POP BC
    RET

ENDMOD
