; =============================================================================
; KIZUNA MSXLIB - vdp/verticaloffset
; rolagem vertical por hardware (R#23): a tela comeca na linha A da VRAM
; =============================================================================

MODULE vdp_verticaloffset
BANK 0

INCLUDE "../../inc/ports.inc"
INCLUDE "../../inc/vdpreg.inc"

PUBLIC VDP_SetVerticalOffset
EXTERN VDP_SetReg

; VDP_SetVerticalOffset: rolagem vertical por hardware (R#23): a tela comeca na linha A da VRAM
; Entrada: A = linha inicial (0..255)
; Preserva: A, BC, DE, HL. Destrói: flags. Reabilita as interrupcoes (EI).
VDP_SetVerticalOffset:
    PUSH AF
    PUSH BC
    LD B, A
    LD C, VDP_REG_VOFFSET
    CALL VDP_SetReg
    POP BC
    POP AF
    RET

ENDMOD
