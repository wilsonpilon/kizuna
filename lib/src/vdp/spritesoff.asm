; =============================================================================
; KIZUNA MSXLIB - vdp/spritesoff
; desliga os sprites (R#8, bit SPD = 1)
; =============================================================================

MODULE vdp_spritesoff
BANK 0

INCLUDE "../../inc/ports.inc"
INCLUDE "../../inc/vdpreg.inc"

PUBLIC VDP_SpritesOff
EXTERN VDP_UpdateReg

; VDP_SpritesOff: desliga os sprites (R#8, bit SPD = 1)
; Preserva: A, BC, DE, HL. Destrói: flags. Reabilita as interrupcoes (EI).
VDP_SpritesOff:
    PUSH BC
    PUSH DE
    LD C, VDP_REG_MODE2
    LD D, VDP_R8_SPD
    LD E, VDP_R8_SPD
    CALL VDP_UpdateReg
    POP DE
    POP BC
    RET

ENDMOD
