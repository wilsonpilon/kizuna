; =============================================================================
; KIZUNA MSXLIB - vdp/displayoff
; desliga a exibicao (a tela fica so com a cor de fundo; a VRAM pode ser gravada mais rapido)
; =============================================================================

MODULE vdp_displayoff
BANK 0

INCLUDE "../../inc/ports.inc"
INCLUDE "../../inc/vdpreg.inc"

PUBLIC VDP_DisplayOff
EXTERN VDP_UpdateReg

; VDP_DisplayOff: desliga a exibicao (a tela fica so com a cor de fundo; a VRAM pode ser gravada mais rapido)
; Preserva: A, BC, DE, HL. Destrói: flags. Reabilita as interrupcoes (EI).
VDP_DisplayOff:
    PUSH BC
    PUSH DE
    LD C, VDP_REG_MODE1
    LD D, VDP_R1_BL
    LD E, 00h
    CALL VDP_UpdateReg
    POP DE
    POP BC
    RET

ENDMOD
