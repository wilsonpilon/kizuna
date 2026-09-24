; =============================================================================
; KIZUNA MSXLIB - vdp/vblankinton
; liga a interrupcao de vblank do VDP (R#1, bit IE0)
; =============================================================================

MODULE vdp_vblankinton
BANK 0

INCLUDE "../../inc/ports.inc"
INCLUDE "../../inc/vdpreg.inc"

PUBLIC VDP_VBlankIntOn
EXTERN VDP_UpdateReg

; VDP_VBlankIntOn: liga a interrupcao de vblank do VDP (R#1, bit IE0)
; Preserva: A, BC, DE, HL. Destrói: flags. Reabilita as interrupcoes (EI).
VDP_VBlankIntOn:
    PUSH BC
    PUSH DE
    LD C, VDP_REG_MODE1
    LD D, VDP_R1_IE0
    LD E, VDP_R1_IE0
    CALL VDP_UpdateReg
    POP DE
    POP BC
    RET

ENDMOD
