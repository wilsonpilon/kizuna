; =============================================================================
; KIZUNA MSXLIB - vdp/spriteson
; liga os sprites (R#8, bit SPD = 0)
; =============================================================================

MODULE vdp_spriteson
BANK 0

INCLUDE "../../inc/ports.inc"
INCLUDE "../../inc/vdpreg.inc"

PUBLIC VDP_SpritesOn
EXTERN VDP_UpdateReg

; VDP_SpritesOn: liga os sprites (R#8, bit SPD = 0)
; Preserva: A, BC, DE, HL. Destrói: flags. Reabilita as interrupcoes (EI).
VDP_SpritesOn:
    PUSH BC
    PUSH DE
    LD C, VDP_REG_MODE2
    LD D, VDP_R8_SPD
    LD E, 00h
    CALL VDP_UpdateReg
    POP DE
    POP BC
    RET

ENDMOD
