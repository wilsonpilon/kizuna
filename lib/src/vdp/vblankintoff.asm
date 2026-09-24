; =============================================================================
; KIZUNA MSXLIB - vdp/vblankintoff
; desliga a interrupcao de vblank do VDP
; =============================================================================

MODULE vdp_vblankintoff
BANK 0

INCLUDE "../../inc/ports.inc"
INCLUDE "../../inc/vdpreg.inc"

PUBLIC VDP_VBlankIntOff
EXTERN VDP_UpdateReg

; VDP_VBlankIntOff: desliga a interrupcao de vblank do VDP
; Preserva: A, BC, DE, HL. Destrói: flags. Reabilita as interrupcoes (EI).
VDP_VBlankIntOff:
    PUSH BC
    PUSH DE
    LD C, VDP_REG_MODE1
    LD D, VDP_R1_IE0
    LD E, 00h
    CALL VDP_UpdateReg
    POP DE
    POP BC
    RET

ENDMOD
