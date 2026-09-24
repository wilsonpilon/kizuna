; =============================================================================
; KIZUNA MSXLIB - vdp/adjustraw
; ajuste de posicao da tela, byte cru de R#18 (nibble baixo = horizontal, alto = vertical) -- valor cru do registrador; o sentido do deslocamento em pixels nao foi verificado em hardware
; =============================================================================

MODULE vdp_adjustraw
BANK 0

INCLUDE "../../inc/ports.inc"
INCLUDE "../../inc/vdpreg.inc"

PUBLIC VDP_SetAdjustRaw
EXTERN VDP_SetReg

; VDP_SetAdjustRaw: ajuste de posicao da tela, byte cru de R#18 (nibble baixo = horizontal, alto = vertical) -- valor cru do registrador; o sentido do deslocamento em pixels nao foi verificado em hardware
; Entrada: A = byte de R#18
; Preserva: A, BC, DE, HL. Destrói: flags. Reabilita as interrupcoes (EI).
VDP_SetAdjustRaw:
    PUSH AF
    PUSH BC
    LD B, A
    LD C, VDP_REG_ADJUST
    CALL VDP_SetReg
    POP BC
    POP AF
    RET

ENDMOD
