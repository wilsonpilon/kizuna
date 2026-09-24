; =============================================================================
; KIZUNA MSXLIB - vdp/leftmask
; esconde a coluna mais a esquerda, 8 pixels (V9958, R#25 bit MSK)
; =============================================================================

MODULE vdp_leftmask
BANK 0

INCLUDE "../../inc/ports.inc"
INCLUDE "../../inc/vdpreg.inc"

PUBLIC VDP_LeftMask
EXTERN VDP_UpdateReg

; VDP_LeftMask: esconde a coluna mais a esquerda, 8 pixels (V9958, R#25 bit MSK)
; Entrada: A = 0 = mostra, qualquer outro valor = esconde
; Preserva: A, BC, DE, HL. Destrói: flags. Reabilita as interrupcoes (EI).
VDP_LeftMask:
    PUSH AF
    PUSH BC
    PUSH DE
    LD E, 00h
    OR A
    JR Z, VDP_LeftMask_Go
    LD E, 02h
VDP_LeftMask_Go:
    LD C, VDP_REG_MODE4
    LD D, 02h
    CALL VDP_UpdateReg
    POP DE
    POP BC
    POP AF
    RET

ENDMOD
