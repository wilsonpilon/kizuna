; =============================================================================
; KIZUNA MSXLIB - vdp/hscrollfine
; rolagem horizontal, parte fina (V9958, R#27): 0 a 7 pixels -- valor cru do registrador; o sentido do deslocamento em pixels nao foi verificado em hardware
; =============================================================================

MODULE vdp_hscrollfine
BANK 0

INCLUDE "../../inc/ports.inc"
INCLUDE "../../inc/vdpreg.inc"

PUBLIC VDP_SetHScrollFine
EXTERN VDP_SetReg

; VDP_SetHScrollFine: rolagem horizontal, parte fina (V9958, R#27): 0 a 7 pixels -- valor cru do registrador; o sentido do deslocamento em pixels nao foi verificado em hardware
; Entrada: A = deslocamento fino (0..7)
; Preserva: A, BC, DE, HL. Destrói: flags. Reabilita as interrupcoes (EI).
VDP_SetHScrollFine:
    PUSH AF
    PUSH BC
    AND 07h
    LD B, A
    LD C, VDP_REG_HSCROLLL
    CALL VDP_SetReg
    POP BC
    POP AF
    RET

ENDMOD
