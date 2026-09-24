; =============================================================================
; KIZUNA MSXLIB - vdp/hscrollcoarse
; rolagem horizontal, parte grossa (V9958, R#26): de 8 em 8 pixels -- valor cru do registrador; o sentido do deslocamento em pixels nao foi verificado em hardware
; =============================================================================

MODULE vdp_hscrollcoarse
BANK 0

INCLUDE "../../inc/ports.inc"
INCLUDE "../../inc/vdpreg.inc"

PUBLIC VDP_SetHScrollCoarse
EXTERN VDP_SetReg

; VDP_SetHScrollCoarse: rolagem horizontal, parte grossa (V9958, R#26): de 8 em 8 pixels -- valor cru do registrador; o sentido do deslocamento em pixels nao foi verificado em hardware
; Entrada: A = deslocamento em unidades de 8 pixels (0..63)
; Preserva: A, BC, DE, HL. Destrói: flags. Reabilita as interrupcoes (EI).
VDP_SetHScrollCoarse:
    PUSH AF
    PUSH BC
    AND 3Fh
    LD B, A
    LD C, VDP_REG_HSCROLLH
    CALL VDP_SetReg
    POP BC
    POP AF
    RET

ENDMOD
