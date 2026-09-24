; =============================================================================
; KIZUNA MSXLIB - vdp/spritesize
; tamanho dos sprites (R#1, bit ST)
; =============================================================================

MODULE vdp_spritesize
BANK 0

INCLUDE "../../inc/ports.inc"
INCLUDE "../../inc/vdpreg.inc"

PUBLIC VDP_SpriteSize
EXTERN VDP_UpdateReg

; VDP_SpriteSize: tamanho dos sprites (R#1, bit ST)
; Entrada: A = 0 = 8x8, qualquer outro valor = 16x16
; Preserva: A, BC, DE, HL. Destrói: flags. Reabilita as interrupcoes (EI).
VDP_SpriteSize:
    PUSH AF
    PUSH BC
    PUSH DE
    LD E, 00h
    OR A
    JR Z, VDP_SpriteSize_Go
    LD E, 02h
VDP_SpriteSize_Go:
    LD C, VDP_REG_MODE1
    LD D, 02h
    CALL VDP_UpdateReg
    POP DE
    POP BC
    POP AF
    RET

ENDMOD
